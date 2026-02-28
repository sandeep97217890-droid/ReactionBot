package main

import (
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/amarnathcjd/gogram/telegram"
	"github.com/joho/godotenv"
	"github.com/sandeep97217890-droid/ReactionBot/handlers"
	"github.com/sandeep97217890-droid/ReactionBot/store"
)

func main() {
	// Load .env file if present (errors are fine – env vars may already be set)
	_ = godotenv.Load()

	appIDStr := mustEnv("APP_ID")
	appHash := mustEnv("APP_HASH")

	appID, err := strconv.ParseInt(appIDStr, 10, 32)
	if err != nil {
		log.Fatalf("APP_ID must be a valid integer: %v", err)
	}

	// Optional: provide a pre-exported string session via SESSION_STRING
	// or rely on the interactive prompt on first run.
	stringSession := os.Getenv("SESSION_STRING")
	sessionFile := os.Getenv("SESSION_FILE")
	if sessionFile == "" {
		sessionFile = "session.session"
	}

	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "reactions.db"
	}

	// Open / create the SQLite store (auto-react is ON by default on first run).
	st, err := store.New(dbPath)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer st.Close()

	cfg := telegram.ClientConfig{
		AppID:         int32(appID),
		AppHash:       appHash,
		Session:       sessionFile,
		StringSession: stringSession,
		LogLevel:      telegram.LogInfo,
	}

	client, err := telegram.NewClient(cfg)
	if err != nil {
		log.Fatalf("Failed to create Telegram client: %v", err)
	}

	// Connect and authenticate (interactive prompt on first run if no session present).
	if err := client.Start(); err != nil {
		log.Fatalf("Failed to start client: %v", err)
	}

	me, err := client.GetMe()
	if err != nil {
		log.Fatalf("Failed to get self user: %v", err)
	}

	isPremium := me.Premium
	fmt.Printf("Logged in as: %s %s (id=%d, premium=%v)\n",
		me.FirstName, me.LastName, me.ID, isPremium)
	fmt.Printf("Auto-react enabled: %v  |  DB: %s\n", st.IsEnabled(), dbPath)

	handlers.Register(client, st, me.ID, isPremium)

	fmt.Println("ReactionBot is running. Commands: /react on|off  /addchat  /removechat  /addpremoji  /addnpemoji  /listchats  /listemojis  /status")
	client.Idle()
}

func mustEnv(key string) string {
	v := os.Getenv(key)
	if v == "" {
		log.Fatalf("Required environment variable %s is not set", key)
	}
	return v
}
