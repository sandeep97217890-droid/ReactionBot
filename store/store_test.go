package store_test

import (
"os"
"path/filepath"
"testing"

"github.com/sandeep97217890-droid/ReactionBot/store"
)

func TestStore(t *testing.T) {
dbPath := filepath.Join(os.TempDir(), "test_reactions_unit.db")
os.Remove(dbPath)

st, err := store.New(dbPath)
if err != nil {
t.Fatalf("New: %v", err)
}
defer st.Close()

// Enabled by default
if !st.IsEnabled() {
t.Error("expected IsEnabled=true by default")
}

// AddChat / HasChat
if err := st.AddChat(123456789); err != nil {
t.Fatalf("AddChat: %v", err)
}
if !st.HasChat(123456789) {
t.Error("expected HasChat=true after AddChat")
}
if st.HasChat(999) {
t.Error("expected HasChat=false for unknown chat")
}

// Idempotent AddChat
if err := st.AddChat(123456789); err != nil {
t.Fatalf("AddChat duplicate: %v", err)
}
chats, err := st.GetChats()
if err != nil {
t.Fatalf("GetChats: %v", err)
}
if len(chats) != 1 {
t.Errorf("expected 1 chat, got %d", len(chats))
}

// RemoveChat
if err := st.RemoveChat(123456789); err != nil {
t.Fatalf("RemoveChat: %v", err)
}
if st.HasChat(123456789) {
t.Error("expected HasChat=false after RemoveChat")
}

// Premium emojis (5 seeded by default)
prem, err := st.GetPremEmojis()
if err != nil {
t.Fatalf("GetPremEmojis: %v", err)
}
if len(prem) != 5 {
t.Errorf("expected 5 default prem emojis, got %d", len(prem))
}
if err := st.AddPremEmoji("🎊"); err != nil {
t.Fatalf("AddPremEmoji: %v", err)
}
prem2, _ := st.GetPremEmojis()
if len(prem2) != 6 {
t.Errorf("expected 6 prem emojis after add, got %d", len(prem2))
}
// Idempotent add
_ = st.AddPremEmoji("🎊")
prem3, _ := st.GetPremEmojis()
if len(prem3) != 6 {
t.Errorf("expected 6 prem emojis after duplicate add, got %d", len(prem3))
}

// Non-premium emojis (3 seeded by default)
nprem, err := st.GetNpremEmojis()
if err != nil {
t.Fatalf("GetNpremEmojis: %v", err)
}
if len(nprem) != 3 {
t.Errorf("expected 3 default nprem emojis, got %d", len(nprem))
}

// SetEnabled toggle
if err := st.SetEnabled(false); err != nil {
t.Fatalf("SetEnabled false: %v", err)
}
if st.IsEnabled() {
t.Error("expected IsEnabled=false after SetEnabled(false)")
}
if err := st.SetEnabled(true); err != nil {
t.Fatalf("SetEnabled true: %v", err)
}
if !st.IsEnabled() {
t.Error("expected IsEnabled=true after SetEnabled(true)")
}
}
