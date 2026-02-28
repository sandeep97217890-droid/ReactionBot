package handlers

import (
"fmt"
"math/rand/v2"
"strconv"
"strings"

"github.com/amarnathcjd/gogram/telegram"
"github.com/sandeep97217890-droid/ReactionBot/store"
)

func Register(client *telegram.Client, st *store.Store, ownerID int64, isPremium bool) {
client.AddCommandHandler("react", func(m *telegram.NewMessage) error {
if m.SenderID() != ownerID {
return nil
}
arg := strings.ToLower(strings.TrimSpace(m.Args()))
switch arg {
case "on":
if err := st.SetEnabled(true); err != nil {
_, _ = m.Reply("❌ Failed to enable: " + err.Error())
return err
}
_, _ = m.Reply("✅ Auto-reactions enabled.")
case "off":
if err := st.SetEnabled(false); err != nil {
_, _ = m.Reply("❌ Failed to disable: " + err.Error())
return err
}
_, _ = m.Reply("🚫 Auto-reactions disabled.")
default:
_, _ = m.Reply("Usage: /react on|off")
}
return nil
})

client.AddCommandHandler("addchat", func(m *telegram.NewMessage) error {
if m.SenderID() != ownerID {
return nil
}
arg := strings.TrimSpace(m.Args())
if arg == "" {
_, _ = m.Reply("Usage: /addchat <chat_id>")
return nil
}
chatID, err := strconv.ParseInt(arg, 10, 64)
if err != nil {
_, _ = m.Reply("❌ Invalid chat ID: must be a number.")
return nil
}
if err := st.AddChat(chatID); err != nil {
_, _ = m.Reply("❌ Failed to add chat: " + err.Error())
return err
}
_, _ = m.Reply(fmt.Sprintf("✅ Chat %d added to auto-react list.", chatID))
return nil
})

client.AddCommandHandler("removechat", func(m *telegram.NewMessage) error {
if m.SenderID() != ownerID {
return nil
}
arg := strings.TrimSpace(m.Args())
if arg == "" {
_, _ = m.Reply("Usage: /removechat <chat_id>")
return nil
}
chatID, err := strconv.ParseInt(arg, 10, 64)
if err != nil {
_, _ = m.Reply("❌ Invalid chat ID: must be a number.")
return nil
}
if err := st.RemoveChat(chatID); err != nil {
_, _ = m.Reply("❌ Failed to remove chat: " + err.Error())
return err
}
_, _ = m.Reply(fmt.Sprintf("✅ Chat %d removed from auto-react list.", chatID))
return nil
})

client.AddCommandHandler("addpremoji", func(m *telegram.NewMessage) error {
if m.SenderID() != ownerID {
return nil
}
emoji := strings.TrimSpace(m.Args())
if emoji == "" {
_, _ = m.Reply("Usage: /addpremoji <emoji>")
return nil
}
if err := st.AddPremEmoji(emoji); err != nil {
_, _ = m.Reply("❌ Failed to add premium emoji: " + err.Error())
return err
}
_, _ = m.Reply("✅ Premium emoji added: " + emoji)
return nil
})

client.AddCommandHandler("addnpemoji", func(m *telegram.NewMessage) error {
if m.SenderID() != ownerID {
return nil
}
emoji := strings.TrimSpace(m.Args())
if emoji == "" {
_, _ = m.Reply("Usage: /addnpemoji <emoji>")
return nil
}
if err := st.AddNpremEmoji(emoji); err != nil {
_, _ = m.Reply("❌ Failed to add non-premium emoji: " + err.Error())
return err
}
_, _ = m.Reply("✅ Non-premium emoji added: " + emoji)
return nil
})

client.AddCommandHandler("listchats", func(m *telegram.NewMessage) error {
if m.SenderID() != ownerID {
return nil
}
chats, err := st.GetChats()
if err != nil {
_, _ = m.Reply("❌ Error: " + err.Error())
return err
}
if len(chats) == 0 {
_, _ = m.Reply("No chats added yet. Use /addchat <chat_id>.")
return nil
}
parts := make([]string, len(chats))
for i, id := range chats {
parts[i] = strconv.FormatInt(id, 10)
}
_, _ = m.Reply("📋 Monitored chats:\n" + strings.Join(parts, "\n"))
return nil
})

client.AddCommandHandler("listemojis", func(m *telegram.NewMessage) error {
if m.SenderID() != ownerID {
return nil
}
prem, err := st.GetPremEmojis()
if err != nil {
_, _ = m.Reply("❌ Error: " + err.Error())
return err
}
nprem, err := st.GetNpremEmojis()
if err != nil {
_, _ = m.Reply("❌ Error: " + err.Error())
return err
}
reply := fmt.Sprintf(
"⭐ Premium emojis (%d):\n%s\n\n👤 Non-premium emojis (%d):\n%s",
len(prem), strings.Join(prem, " "),
len(nprem), strings.Join(nprem, " "),
)
_, _ = m.Reply(reply)
return nil
})

client.AddCommandHandler("status", func(m *telegram.NewMessage) error {
if m.SenderID() != ownerID {
return nil
}
state := "🚫 OFF"
if st.IsEnabled() {
state = "✅ ON"
}
acctType := "👤 Non-Premium"
if isPremium {
acctType = "⭐ Premium (3 reactions per message)"
}
chats, _ := st.GetChats()
_, _ = m.Reply(fmt.Sprintf(
"🤖 ReactionBot Status\nAuto-react: %s\nAccount: %s\nMonitored chats: %d",
state, acctType, len(chats),
))
return nil
})

client.AddMessageHandler(telegram.OnNewMessage, func(m *telegram.NewMessage) error {
if m.SenderID() == ownerID {
return nil
}
if !st.IsEnabled() {
return nil
}
if !st.HasChat(m.ChatID()) {
return nil
}
reaction, err := pickReaction(st, isPremium)
if err != nil || len(reaction) == 0 {
return nil
}
reactions := make([]any, len(reaction))
for i, r := range reaction {
reactions[i] = r
}
return m.React(reactions...)
})
}

func pickReaction(st *store.Store, isPremium bool) ([]string, error) {
if isPremium {
emojis, err := st.GetPremEmojis()
if err != nil || len(emojis) == 0 {
return nil, err
}
rand.Shuffle(len(emojis), func(i, j int) { emojis[i], emojis[j] = emojis[j], emojis[i] })
count := 3
if len(emojis) < count {
count = len(emojis)
}
return emojis[:count], nil
}
emojis, err := st.GetNpremEmojis()
if err != nil || len(emojis) == 0 {
return nil, err
}
return []string{emojis[rand.IntN(len(emojis))]}, nil
}
