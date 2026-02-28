package handlers

import (
	"fmt"
	"math/rand/v2"
	"strconv"
	"strings"

	"github.com/amarnathcjd/gogram/telegram"
	"github.com/sandeep97217890-droid/ReactionBot/store"
)

const maxPremiumReactions = 3

func Register(client *telegram.Client, st *store.Store, ownerID int64, isPremium bool) {
	client.On(telegram.OnNewMessage, func(m *telegram.NewMessage) error {
		if m.SenderID() == ownerID {
			return nil
		}
		if !st.IsEnabled() {
			return nil
		}
		if !st.HasChat(m.ChatID()) {
			return nil
		}
		var reaction []string
		if isPremium {
			emojis, err := st.GetPremEmojis()
			if err != nil || len(emojis) == 0 {
				return nil
			}
			rand.Shuffle(len(emojis), func(i, j int) { emojis[i], emojis[j] = emojis[j], emojis[i] })
			count := maxPremiumReactions
			if len(emojis) < count {
				count = len(emojis)
			}
			reaction = emojis[:count]
		} else {
			emojis, err := st.GetNpremEmojis()
			if err != nil || len(emojis) == 0 {
				return nil
			}
			reaction = []string{emojis[rand.IntN(len(emojis))]}
		}
		reactions := make([]any, len(reaction))
		for i, r := range reaction {
			reactions[i] = r
		}
		return m.React(reactions...)
	})
}

const helpText = `🤖 *ReactionBot Commands*

/start - Show welcome message
/help - Show this help message
/react on|off - Enable or disable auto-reactions
/addchat <chat_id> - Add a chat to the auto-react list
/removechat <chat_id> - Remove a chat from the auto-react list
/listchats - List all monitored chats
/addpremoji <emoji> - Add a premium emoji
/addnpemoji <emoji> - Add a non-premium emoji
/listemojis - List all configured emojis
/status - Show current bot status`

func RegisterBot(client *telegram.Client, st *store.Store, ownerIDs []int64) {
	f := telegram.FromUser(ownerIDs...)

	client.On("cmd:start", func(m *telegram.NewMessage) error {
		_, _ = m.Reply("👋 Welcome to *ReactionBot*!\n\nI automatically react to messages in configured chats.\nSend /help to see all available commands.")
		return nil
	})

	client.On("cmd:help", func(m *telegram.NewMessage) error {
		_, _ = m.Reply(helpText)
		return nil
	})

	client.On("cmd:react", func(m *telegram.NewMessage) error {
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
	}, f)

	client.On("cmd:addchat", func(m *telegram.NewMessage) error {
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
	}, f)

	client.On("cmd:removechat", func(m *telegram.NewMessage) error {
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
	}, f)

	client.On("cmd:addpremoji", func(m *telegram.NewMessage) error {
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
	}, f)

	client.On("cmd:addnpemoji", func(m *telegram.NewMessage) error {
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
	}, f)

	client.On("cmd:listchats", func(m *telegram.NewMessage) error {
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
	}, f)

	client.On("cmd:listemojis", func(m *telegram.NewMessage) error {
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
		_, _ = m.Reply(fmt.Sprintf(
			"⭐ Premium emojis (%d):\n%s\n\n👤 Non-premium emojis (%d):\n%s",
			len(prem), strings.Join(prem, " "),
			len(nprem), strings.Join(nprem, " "),
		))
		return nil
	}, f)

	client.On("cmd:status", func(m *telegram.NewMessage) error {
		state := "🚫 OFF"
		if st.IsEnabled() {
			state = "✅ ON"
		}
		chats, _ := st.GetChats()
		_, _ = m.Reply(fmt.Sprintf(
			"🤖 ReactionBot Status\nAuto-react: %s\nAccount: 🤖 Bot\nMonitored chats: %d",
			state, len(chats),
		))
		return nil
	}, f)
}

