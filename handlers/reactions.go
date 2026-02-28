package handlers

import "strings"

// validReactions is the set of emoji strings accepted for Telegram reactions.
// Variation selector U+FE0F is stripped before lookup so both "❤" and "❤️" match.
var validReactions = map[string]struct{}{
	"👍":      {},
	"👎":      {},
	"❤":       {},
	"🔥":      {},
	"🥰":      {},
	"👏":      {},
	"😁":      {},
	"🤔":      {},
	"🤯":      {},
	"😱":      {},
	"🤬":      {},
	"😢":      {},
	"🎉":      {},
	"🤩":      {},
	"🤮":      {},
	"💩":      {},
	"🙏":      {},
	"👌":      {},
	"🕊":      {},
	"🤡":      {},
	"🥱":      {},
	"🥴":      {},
	"😍":      {},
	"🐳":      {},
	"❤‍🔥":    {},
	"🌚":      {},
	"🌭":      {},
	"💯":      {},
	"🤣":      {},
	"⚡":      {},
	"🍌":      {},
	"🏆":      {},
	"💔":      {},
	"🤨":      {},
	"😐":      {},
	"🍓":      {},
	"🍾":      {},
	"💋":      {},
	"🖕":      {},
	"😈":      {},
	"😴":      {},
	"😭":      {},
	"🤓":      {},
	"👻":      {},
	"👨‍💻": {},
	"👀":      {},
	"🎃":      {},
	"🙈":      {},
	"😇":      {},
	"😨":      {},
	"🤝":      {},
	"✍":       {},
	"🤗":      {},
	"🫡":      {},
	"🎅":      {},
	"🎄":      {},
	"☃":       {},
	"💅":      {},
	"🤪":      {},
	"🗿":      {},
	"💀":      {},
	"🌹":      {},
	"🌊":      {},
	"😆":      {},
}

// ValidReactionList returns a slice of all valid reaction emojis.
func ValidReactionList() []string {
	list := make([]string, 0, len(validReactions))
	for e := range validReactions {
		list = append(list, e)
	}
	return list
}

// stripVariationSelector removes Unicode variation selector characters
// (U+FE0F and U+FE0E) so both "❤" and "❤️" match the same entry.
func stripVariationSelector(s string) string {
	return strings.NewReplacer("\uFE0F", "", "\uFE0E", "").Replace(s)
}

// IsValidReaction reports whether emoji is a known Telegram reaction emoji.
func IsValidReaction(emoji string) bool {
	_, ok := validReactions[stripVariationSelector(emoji)]
	return ok
}
