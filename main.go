package main

import (
"log"
"os"
"strconv"
"strings"
"sync"
"sync/atomic"

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

total := len(premSessions) + len(npremSessions)
if total == 0 {
log.Fatal("No sessions configured. Set PREM_SESSIONS and/or NPREM_SESSIONS.")
}

var (
wg      sync.WaitGroup
started int64
)

for _, sess := range premSessions {
wg.Add(1)
go func(s string) {
defer wg.Done()
if runClient(int32(appID), appHash, s, true, st) {
atomic.AddInt64(&started, 1)
}
}(sess)
}

for _, sess := range npremSessions {
wg.Add(1)
go func(s string) {
defer wg.Done()
if runClient(int32(appID), appHash, s, false, st) {
atomic.AddInt64(&started, 1)
}
}(sess)
}

wg.Wait()

if started == 0 {
log.Fatal("All clients failed to start.")
}
}

func runClient(appID int32, appHash, session string, isPremium bool, st *store.Store) bool {
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

if err := client.Start(); err != nil {
log.Printf("Failed to start client: %v", err)
return false
}

me, err := client.GetMe()
if err != nil {
log.Printf("Failed to get self user: %v", err)
return false
}

log.Printf("Logged in as: %s %s (id=%d, premium=%v)", me.FirstName, me.LastName, me.ID, isPremium)

handlers.Register(client, st, me.ID, isPremium)

client.Idle()
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

func mustEnv(key string) string {
v := os.Getenv(key)
if v == "" {
log.Fatalf("Required environment variable %s is not set", key)
}
return v
}
