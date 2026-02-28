package handlers

import "strings"

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

func ValidReactionList() []string {
	list := make([]string, 0, len(validReactions))
	for e := range validReactions {
		list = append(list, e)
	}
	return list
}

func stripVariationSelector(s string) string {
	return strings.NewReplacer("\uFE0F", "", "\uFE0E", "").Replace(s)
}

func IsValidReaction(emoji string) bool {
	_, ok := validReactions[stripVariationSelector(emoji)]
	return ok
}

