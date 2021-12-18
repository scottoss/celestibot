package cmds

import (
	"equestriaunleashed.com/eclipsingr/celestibot/core"
	"strings"
)

// SayCommand puts neigh in the chat.
func SayCommand(a core.CommandArgs, v []string) bool {
	if a.HasPermission() {
		txt := v
		if len(v) > 0 {
			chnid := a.Event.ChannelID
			if strings.HasPrefix(v[0], "<#") && strings.HasSuffix(v[0], ">") {
				chnid = v[0][2:len(v[0])-1]
				txt = v[1:]
			}
			a.Session.ChannelMessageSend(chnid, core.SliceToString(txt))
		}
	} else {
		a.SendMessage("<@" + a.Event.Author.ID + "> What, no! I won't say that.")
	}
	return true
}
