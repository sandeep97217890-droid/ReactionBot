package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/amarnathcjd/gogram/telegram"
	"github.com/joho/godotenv"
	"github.com/sandeep97217890-droid/ReactionBot/handlers"
	"github.com/sandeep97217890-droid/ReactionBot/store"
)

func main() {
	_ = godotenv.Load()

	appIDStr := mustEnv("APP_ID")
	appHash := mustEnv("APP_HASH")

	appID, err := strconv.ParseInt(appIDStr, 10, 32)
	if err != nil {
		log.Fatalf("APP_ID must be a valid integer: %v", err)
	}

	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "reactions.db"
	}

	st, err := store.New(dbPath)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer st.Close()

	premSessions := parseSessions(os.Getenv("PREM_SESSIONS"))
	npremSessions := parseSessions(os.Getenv("NPREM_SESSIONS"))
	botToken := os.Getenv("BOT_TOKEN")

	if len(premSessions)+len(npremSessions) == 0 && botToken == "" {
		log.Fatal("No sessions or bot token configured. Set PREM_SESSIONS, NPREM_SESSIONS and/or BOT_TOKEN.")
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	var (
		wg      sync.WaitGroup
		started int64
	)

	for _, sess := range premSessions {
		wg.Add(1)
		go func(s string) {
			defer wg.Done()
			if runClient(ctx, int32(appID), appHash, s, true, st) {
				atomic.AddInt64(&started, 1)
			}
		}(sess)
	}

	for _, sess := range npremSessions {
		wg.Add(1)
		go func(s string) {
			defer wg.Done()
			if runClient(ctx, int32(appID), appHash, s, false, st) {
				atomic.AddInt64(&started, 1)
			}
		}(sess)
	}

	if botToken != "" {
		ownerIDs := parseOwnerIDs(mustEnv("OWNER_IDS"))
		if len(ownerIDs) == 0 {
			log.Fatal("OWNER_IDS must contain at least one valid user ID when BOT_TOKEN is configured.")
		}
		wg.Add(1)
		go func() {
			defer wg.Done()
			if runBot(ctx, int32(appID), appHash, botToken, ownerIDs, st) {
				atomic.AddInt64(&started, 1)
			}
		}()
	}

	wg.Wait()

	if started == 0 {
		log.Fatal("All clients failed to start.")
	}
}

func runClient(ctx context.Context, appID int32, appHash, session string, isPremium bool, st *store.Store) bool {
	cfg := telegram.ClientConfig{
		AppID:         appID,
		AppHash:       appHash,
		StringSession: session,
		MemorySession: true,
		LogLevel:      telegram.LogInfo,
	}

	client, err := telegram.NewClient(cfg)
	if err != nil {
		log.Printf("Failed to create client: %v", err)
		return false
	}

	startErr := make(chan error, 1)
	go func() { startErr <- client.Start() }()
	select {
	case err := <-startErr:
		if err != nil {
			log.Printf("Failed to start client: %v", err)
			return false
		}
	case <-time.After(30 * time.Second):
		log.Printf("Client start timed out, skipping session")
		_ = client.Stop()
		return false
	}

	me, err := client.GetMe()
	if err != nil {
		log.Printf("Failed to get self user: %v", err)
		return false
	}

	log.Printf("Logged in as: %s %s (id=%d, premium=%v)", me.FirstName, me.LastName, me.ID, isPremium)

	handlers.Register(client, st, me.ID, isPremium)

	<-ctx.Done()
	_ = client.Stop()
	return true
}

func runBot(ctx context.Context, appID int32, appHash, botToken string, ownerIDs []int64, st *store.Store) bool {
	client, err := telegram.NewClient(telegram.ClientConfig{
		AppID:         appID,
		AppHash:       appHash,
		MemorySession: true,
		LogLevel:      telegram.LogInfo,
	})
	if err != nil {
		log.Printf("Failed to create bot client: %v", err)
		return false
	}

	botErr := make(chan error, 1)
	go func() {
		if err := client.Connect(); err != nil {
			botErr <- err
			return
		}
		botErr <- client.LoginBot(botToken)
	}()
	select {
	case err := <-botErr:
		if err != nil {
			log.Printf("Failed to start bot: %v", err)
			return false
		}
	case <-time.After(30 * time.Second):
		log.Printf("Bot start timed out")
		_ = client.Stop()
		return false
	}

	me, err := client.GetMe()
	if err != nil {
		log.Printf("Failed to get bot user: %v", err)
		return false
	}

	log.Printf("Bot logged in as: @%s (id=%d)", me.Username, me.ID)

	handlers.RegisterBot(client, st, ownerIDs)

	<-ctx.Done()
	_ = client.Stop()
	return true
}

func parseSessions(raw string) []string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil
	}
	parts := strings.Split(raw, ",")
	var out []string
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

func parseOwnerIDs(raw string) []int64 {
	var ids []int64
	for _, p := range strings.Split(raw, ",") {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		id, err := strconv.ParseInt(p, 10, 64)
		if err != nil {
			log.Printf("Skipping invalid owner ID %q: %v", p, err)
			continue
		}
		ids = append(ids, id)
	}
	return ids
}

func mustEnv(key string) string {
	v := os.Getenv(key)
	if v == "" {
		log.Fatalf("Required environment variable %s is not set", key)
	}
	return v
}
