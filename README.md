# ReactionBot

A Telegram userbot built with [gogram](https://github.com/AmarnathCJD/gogram) (Go) that automatically reacts to messages in configured chats.

- **Premium accounts** send 3 randomly-picked reactions from the premium emoji pool.
- **Non-premium accounts** send 1 randomly-picked reaction from the non-premium pool.
- State (enabled flag, chat list, emoji pools) is stored in an **SQLite** database.
- Auto-react is **enabled by default** when the bot starts for the first time.

---

## Setup

```bash
# 1. Clone & enter the repo
git clone https://github.com/sandeep97217890-droid/ReactionBot.git
cd ReactionBot

# 2. Copy and fill in credentials
cp .env.example .env
$EDITOR .env        # set APP_ID and APP_HASH from https://my.telegram.org

# 3. Build
go build -o reactionbot .

# 4. Run (interactive login prompt on first run if no SESSION_STRING is set)
./reactionbot
```

---

## Commands

All commands are only accepted from the account owner (the logged-in user).

| Command | Description |
|---|---|
| `/react on` | Enable auto-reactions |
| `/react off` | Disable auto-reactions |
| `/addchat <chat_id>` | Add a chat/channel to the monitored list |
| `/removechat <chat_id>` | Remove a chat/channel from the monitored list |
| `/addpremoji <emoji>` | Add an emoji to the **premium** reaction pool |
| `/addnpemoji <emoji>` | Add an emoji to the **non-premium** reaction pool |
| `/listchats` | Show all monitored chats |
| `/listemojis` | Show all configured emojis |
| `/status` | Show current bot state |

---

## Environment Variables

| Variable | Required | Default | Description |
|---|---|---|---|
| `APP_ID` | ✅ | — | Telegram API ID from my.telegram.org |
| `APP_HASH` | ✅ | — | Telegram API hash from my.telegram.org |
| `SESSION_STRING` | ❌ | — | Pre-exported session string (skips interactive login) |
| `SESSION_FILE` | ❌ | `session.session` | Path to the session file |
| `DB_PATH` | ❌ | `reactions.db` | Path to the SQLite database |

---

## Default Emoji Pools

These are seeded on first run and can be extended with `/addpremoji` / `/addnpemoji`.

| Pool | Default emojis |
|---|---|
| Premium | 🐳 ❤️ 👍 🎉 👌 |
| Non-premium | 👍 ❤️ 🔥 |
