package handlers

import (
	"fmt"
	"html"
	"log"
	"math/rand/v2"
	"strconv"
	"strings"
	"sync"

	"github.com/amarnathcjd/gogram/telegram"
	"github.com/sandeep97217890-droid/ReactionBot/store"
)

const maxPremiumReactions = 3

type Session struct {
	Client    *telegram.Client
	IsPremium bool
}

func Register(sessions []Session, st *store.Store) {
	if len(sessions) == 0 {
		return
	}
	var seen sync.Map
	for _, sess := range sessions {
		sess := sess
		sess.Client.On(telegram.OnNewMessage, func(m *telegram.NewMessage) error {
			if !st.IsEnabled() || !st.HasChat(m.ChatID()) {
				return nil
			}
			chatID := m.ChannelID()
			msgID := m.ID
			key := [2]int64{chatID, int64(msgID)}
			if _, loaded := seen.LoadOrStore(key, struct{}{}); loaded {
				return nil
			}
			fmt.Println("Received message in chat", m.ChatID(), "– reacting with all sessions")
			for _, s := range sessions {
				sendReaction(s, st, chatID, msgID)
			}
			return nil
		})
	}
}

func sendReaction(sess Session, st *store.Store, chatID int64, msgID int32) {
	var reaction []string
	if sess.IsPremium {
		emojis, err := st.GetPremEmojis()
		if err != nil || len(emojis) == 0 {
			return
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
			return
		}
		reaction = []string{emojis[rand.IntN(len(emojis))]}
	}
	if err := sess.Client.SendReaction(chatID, msgID, reaction, true); err != nil {
		log.Printf("SendReaction failed (isPremium=%v, chatID=%d, msgID=%d, reaction=%v): %v", sess.IsPremium, chatID, msgID, reaction, err)
	}
}

const helpText = `🤖 <b>ReactionBot Commands</b>

/start - Show welcome message
/help - Show this help message
/react on|off - Enable or disable auto-reactions
/joinchat &lt;link&gt; - Join a chat via private (<code>+Hash</code>) or public (<code>@username</code>) invite link
/addchat &lt;chat_id&gt; - Add a chat to the auto-react list
/removechat &lt;chat_id&gt; - Remove a chat from the auto-react list
/listchats - List all monitored chats
/addpremoji &lt;emoji…&gt; - Add one or more premium reaction emojis (space-separated)
/addnpemoji &lt;emoji…&gt; - Add one or more non-premium reaction emojis (space-separated)
/listemojis - List all configured emojis
/validreactions - Show all valid Telegram reaction emojis
/status - Show current bot status`

func RegisterBot(client *telegram.Client, st *store.Store, ownerIDs []int64, userClients []*telegram.Client) {
	f := telegram.FromUser(ownerIDs...)

	client.On("cmd:start", func(m *telegram.NewMessage) error {
		_, _ = m.Reply("👋 Welcome to <b>ReactionBot</b>!\n\nI automatically react to messages in configured chats.\nSend /help to see all available commands.")
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
			if st.IsEnabled() {
				_, _ = m.Reply("ℹ️ Auto-reactions are already enabled.")
				return nil
			}
			if err := st.SetEnabled(true); err != nil {
				_, _ = m.Reply("❌ Failed to enable: " + err.Error())
				return err
			}
			_, _ = m.Reply("✅ Auto-reactions enabled.")
		case "off":
			if !st.IsEnabled() {
				_, _ = m.Reply("ℹ️ Auto-reactions are already disabled.")
				return nil
			}
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

	client.On("cmd:joinchat", func(m *telegram.NewMessage) error {
		arg := strings.TrimSpace(m.Args())
		if arg == "" {
			_, _ = m.Reply("Usage: /joinchat &lt;invite_link&gt;\n\nSupports:\n• Private: <code>+AbCdEfGh</code> or <code>https://t.me/+AbCdEfGh</code>\n• Public: <code>@username</code> or <code>https://t.me/username</code>")
			return nil
		}
		if len(userClients) == 0 {
			_, _ = m.Reply("❌ No userbot sessions configured. Add <code>PREM_SESSIONS</code> or <code>NPREM_SESSIONS</code>.")
			return nil
		}

		link := arg
		if strings.HasPrefix(link, "+") {
			link = "https://t.me/" + link
		}

		var lastErr error
		joined := 0
		for _, uc := range userClients {
			if _, err := uc.JoinChannel(link); err != nil {
				lastErr = err
			} else {
				joined++
			}
		}

		if joined == 0 {
			errMsg := "unknown error"
			if lastErr != nil {
				errMsg = lastErr.Error()
			}
			_, _ = m.Reply("❌ Failed to join chat: " + html.EscapeString(errMsg))
			return nil
		}

		_, _ = m.Reply(fmt.Sprintf(
			"✅ Joined chat via invite link (%d/%d sessions succeeded).\nUse /addchat &lt;chat_id&gt; to start monitoring.",
			joined, len(userClients),
		))
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
		args := strings.Fields(m.Args())
		if len(args) == 0 {
			_, _ = m.Reply("Usage: /addpremoji <emoji…>\nEmojis must be space-separated valid Telegram reactions.\nSee /validreactions for the full list.")
			return nil
		}
		var added, invalid []string
		for _, emoji := range args {
			if !IsValidReaction(emoji) {
				invalid = append(invalid, emoji)
				continue
			}
			if err := st.AddPremEmoji(emoji); err != nil {
				_, _ = m.Reply("❌ Failed to add premium emoji: " + err.Error())
				return err
			}
			added = append(added, emoji)
		}
		var parts []string
		if len(added) > 0 {
			parts = append(parts, "✅ Premium emoji(s) added: "+strings.Join(added, " "))
		}
		if len(invalid) > 0 {
			parts = append(parts, "❌ Invalid reaction emoji(s): "+strings.Join(invalid, " ")+"\nUse /validreactions to see valid options.")
		}
		_, _ = m.Reply(strings.Join(parts, "\n"))
		return nil
	}, f)

	client.On("cmd:addnpemoji", func(m *telegram.NewMessage) error {
		args := strings.Fields(m.Args())
		if len(args) == 0 {
			_, _ = m.Reply("Usage: /addnpemoji <emoji…>\nEmojis must be space-separated valid Telegram reactions.\nSee /validreactions for the full list.")
			return nil
		}
		var added, invalid []string
		for _, emoji := range args {
			if !IsValidReaction(emoji) {
				invalid = append(invalid, emoji)
				continue
			}
			if err := st.AddNpremEmoji(emoji); err != nil {
				_, _ = m.Reply("❌ Failed to add non-premium emoji: " + err.Error())
				return err
			}
			added = append(added, emoji)
		}
		var parts []string
		if len(added) > 0 {
			parts = append(parts, "✅ Non-premium emoji(s) added: "+strings.Join(added, " "))
		}
		if len(invalid) > 0 {
			parts = append(parts, "❌ Invalid reaction emoji(s): "+strings.Join(invalid, " ")+"\nUse /validreactions to see valid options.")
		}
		_, _ = m.Reply(strings.Join(parts, "\n"))
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

	client.On("cmd:validreactions", func(m *telegram.NewMessage) error {
		list := ValidReactionList()
		_, _ = m.Reply(fmt.Sprintf(
			"✅ <b>Valid Telegram reaction emojis (%d):</b>\n%s\n\nUse these with /addnpemoji or /addpremoji (space-separated).",
			len(list), strings.Join(list, " "),
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

