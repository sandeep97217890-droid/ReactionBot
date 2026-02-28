// Package store provides a SQLite-backed persistent state for the reaction bot.
package store

import (
	"database/sql"
	"fmt"
	"sync"

	_ "modernc.org/sqlite" // SQLite driver (CGO-free)
)

// Store wraps a SQLite database with a mutex for safe concurrent access.
type Store struct {
	mu sync.RWMutex
	db *sql.DB
}

// New opens (or creates) the SQLite database at path and initialises the schema.
// Auto-react is enabled by default when the database is first created.
func New(path string) (*Store, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("opening sqlite db: %w", err)
	}
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("connecting to sqlite db: %w", err)
	}

	s := &Store{db: db}
	if err := s.migrate(); err != nil {
		return nil, fmt.Errorf("running migrations: %w", err)
	}
	return s, nil
}

func (s *Store) migrate() error {
	_, err := s.db.Exec(`
		CREATE TABLE IF NOT EXISTS settings (
			key   TEXT PRIMARY KEY,
			value TEXT NOT NULL
		);

		CREATE TABLE IF NOT EXISTS chats (
			chat_id INTEGER PRIMARY KEY
		);

		CREATE TABLE IF NOT EXISTS prem_emojis (
			emoji TEXT PRIMARY KEY
		);

		CREATE TABLE IF NOT EXISTS nprem_emojis (
			emoji TEXT PRIMARY KEY
		);

		-- Seed default setting: auto-react ON by default
		INSERT OR IGNORE INTO settings (key, value) VALUES ('enabled', '1');

		-- Seed default premium emojis
		INSERT OR IGNORE INTO prem_emojis (emoji) VALUES ('🐳');
		INSERT OR IGNORE INTO prem_emojis (emoji) VALUES ('❤️');
		INSERT OR IGNORE INTO prem_emojis (emoji) VALUES ('👍');
		INSERT OR IGNORE INTO prem_emojis (emoji) VALUES ('🎉');
		INSERT OR IGNORE INTO prem_emojis (emoji) VALUES ('👌');

		-- Seed default non-premium emojis
		INSERT OR IGNORE INTO nprem_emojis (emoji) VALUES ('👍');
		INSERT OR IGNORE INTO nprem_emojis (emoji) VALUES ('❤️');
		INSERT OR IGNORE INTO nprem_emojis (emoji) VALUES ('🔥');
	`)
	return err
}

// IsEnabled returns whether auto-react is currently turned on.
func (s *Store) IsEnabled() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var v string
	_ = s.db.QueryRow(`SELECT value FROM settings WHERE key = 'enabled'`).Scan(&v)
	return v == "1"
}

// SetEnabled turns auto-react on or off.
func (s *Store) SetEnabled(enabled bool) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	v := "0"
	if enabled {
		v = "1"
	}
	_, err := s.db.Exec(`INSERT INTO settings (key, value) VALUES ('enabled', ?) ON CONFLICT(key) DO UPDATE SET value = excluded.value`, v)
	return err
}

// HasChat returns true if chatID is being monitored.
func (s *Store) HasChat(chatID int64) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var id int64
	err := s.db.QueryRow(`SELECT chat_id FROM chats WHERE chat_id = ?`, chatID).Scan(&id)
	return err == nil
}

// AddChat adds chatID to the monitored list (no-op if already present).
func (s *Store) AddChat(chatID int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	_, err := s.db.Exec(`INSERT OR IGNORE INTO chats (chat_id) VALUES (?)`, chatID)
	return err
}

// RemoveChat removes chatID from the monitored list.
func (s *Store) RemoveChat(chatID int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	_, err := s.db.Exec(`DELETE FROM chats WHERE chat_id = ?`, chatID)
	return err
}

// AddPremEmoji adds an emoji to the premium list (no-op if already present).
func (s *Store) AddPremEmoji(emoji string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	_, err := s.db.Exec(`INSERT OR IGNORE INTO prem_emojis (emoji) VALUES (?)`, emoji)
	return err
}

// AddNpremEmoji adds an emoji to the non-premium list (no-op if already present).
func (s *Store) AddNpremEmoji(emoji string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	_, err := s.db.Exec(`INSERT OR IGNORE INTO nprem_emojis (emoji) VALUES (?)`, emoji)
	return err
}

// GetPremEmojis returns all premium emojis.
func (s *Store) GetPremEmojis() ([]string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.queryEmojis(`SELECT emoji FROM prem_emojis`)
}

// GetNpremEmojis returns all non-premium emojis.
func (s *Store) GetNpremEmojis() ([]string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.queryEmojis(`SELECT emoji FROM nprem_emojis`)
}

// GetChats returns all monitored chat IDs.
func (s *Store) GetChats() ([]int64, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	rows, err := s.db.Query(`SELECT chat_id FROM chats`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var ids []int64
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}

// Close releases the database connection.
func (s *Store) Close() error {
	return s.db.Close()
}

func (s *Store) queryEmojis(query string) ([]string, error) {
	rows, err := s.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var emojis []string
	for rows.Next() {
		var e string
		if err := rows.Scan(&e); err != nil {
			return nil, err
		}
		emojis = append(emojis, e)
	}
	return emojis, rows.Err()
}
