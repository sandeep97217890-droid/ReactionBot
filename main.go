package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"

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

	var clients []*telegram.Client
	var userClients []*telegram.Client
	var sessions []handlers.Session
	startedCount := 0

	for _, sess := range premSessions {
		if client := startSession(int32(appID), appHash, sess, true); client != nil {
			sessions = append(sessions, handlers.Session{Client: client, IsPremium: true})
			clients = append(clients, client)
			userClients = append(userClients, client)
			startedCount++
		}
	}
	for _, sess := range npremSessions {
		if client := startSession(int32(appID), appHash, sess, false); client != nil {
			sessions = append(sessions, handlers.Session{Client: client, IsPremium: false})
			clients = append(clients, client)
			userClients = append(userClients, client)
			startedCount++
		}
	}

	if len(sessions) > 0 {
		handlers.Register(sessions, st)
	}

	if botToken != "" {
		ownerIDs := parseOwnerIDs(mustEnv("OWNER_IDS"))
		if len(ownerIDs) == 0 {
			log.Fatal("OWNER_IDS must contain at least one valid user ID when BOT_TOKEN is configured.")
		}
		client, err := telegram.NewClient(telegram.ClientConfig{
			AppID:    int32(appID),
			AppHash:  appHash,
			LogLevel: telegram.LogInfo,
		})
		if err != nil {
			log.Printf("Failed to create bot client: %v", err)
		} else if err := client.LoginBot(botToken); err != nil {
			log.Printf("Failed to login bot: %v", err)
			_ = client.Disconnect()
		} else {
			me, err := client.GetMe()
			if err != nil {
				log.Printf("Failed to get bot user: %v", err)
				_ = client.Disconnect()
			} else {
				log.Printf("Bot logged in as: @%s (id=%d)", me.Username, me.ID)
				handlers.RegisterBot(client, st, ownerIDs, userClients)
				clients = append(clients, client)
				startedCount++
			}
		}
	}

	if startedCount == 0 {
		log.Fatal("All clients failed to start.")
	}

	<-ctx.Done()
	for _, c := range clients {
		_ = c.Stop()
	}
}

func startSession(appID int32, appHash, sess string, isPremium bool) *telegram.Client {
	client, err := telegram.NewClient(telegram.ClientConfig{
		AppID:         appID,
		AppHash:       appHash,
		StringSession: sess,
		MemorySession: true,
		LogLevel:      telegram.LogInfo,
	})
	if err != nil {
		log.Printf("Failed to create client: %v", err)
		return nil
	}
	authorized, err := client.IsAuthorized()
	if err != nil {
		log.Printf("Session authorization check failed (%v), skipping", err)
		_ = client.Disconnect()
		return nil
	}
	if !authorized {
		log.Printf("Session not authorized, skipping")
		_ = client.Disconnect()
		return nil
	}
	me, err := client.GetMe()
	if err != nil {
		log.Printf("Failed to get self user: %v", err)
		_ = client.Disconnect()
		return nil
	}
	log.Printf("Logged in as: %s %s (id=%d, premium=%v)", me.FirstName, me.LastName, me.ID, isPremium)
	return client
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
