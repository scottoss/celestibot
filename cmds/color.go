package cmds

import (
	"equestriaunleashed.com/eclipsingr/celestibot/core"
	"strconv"
	"github.com/bwmarrin/discordgo"
	"sort"
	"strings"
)


// SayCommand puts neigh in the chat.
func ColorCommand(a core.CommandArgs, v []string) bool {
	if len(v) == 1 {
		col, err := strconv.ParseUint(v[0], 0, 64)
		if err != nil {
			a.SendMessage("**<@" + a.Event.Author.ID + "> Invalid color, please use a valid uint64 value! (Hexidecimal, base64, etc)")
			return false
		}
		roles, err := a.Session.GuildRoles(a.GetGuild().ID)

		if err != nil {
			a.SendMessage("**<@" + a.Event.Author.ID + "> Could not find roles... ??????")
			return false
		}

		r := GetRoleFromName(roles, a.Event.Author.ID)
		role := r
		if r == nil {
			role, err = a.Session.GuildRoleCreate(a.GetGuild().ID)
			if err != nil {
				a.SendMessage("Failed creating role for color!\n" + err.Error())
				return false
			}
		}
		role, err = a.Session.GuildRoleEdit(a.GetGuild().ID, role.ID, a.Event.Author.ID, int(col), false, 0, false)
		if err != nil {
			a.SendMessage("Failed creating role for color!\n" + err.Error())
			return false
		}
		if r == nil {
			roles = a.GetGuild().Roles
			sort.Sort(discordgo.Roles(roles))

			var froles []*discordgo.Role = make([]*discordgo.Role, len(roles))
			gm, err := a.Session.GuildMember(a.GetGuild().ID, a.GetMe().ID)
			if err != nil {
				a.SendMessage("Failed getting roles!\n" + err.Error())
				return false
			}
			i := 0
			l := len(roles) - 1
			for _, r := range roles {
				if r.ID != role.ID {
					r.Position = l - i
					froles[l-i] = r
					if gm.Roles[0] == r.ID {
						i = i + 1
						role.Position = l - i
						froles[l-i] = role
					}
					i = i + 1
				}
			}
			gr, err := a.Session.GuildRoleReorder(a.GetGuild().ID, froles)
			sort.Sort(discordgo.Roles(gr))
			if err != nil {
				a.SendMessage("Failed reorganizing roles!\n" + err.Error())
				return false
			}
			a.Session.GuildMemberRoleAdd(a.GetGuild().ID, a.Event.Author.ID, role.ID)
		}
	} else if len(v) == 0 {
		roles, err := a.Session.GuildRoles(a.GetGuild().ID)
		if err != nil {
			a.SendMessage("**<@" + a.Event.Author.ID + "> Could not find roles... ??????")
			return false
		}

		r := GetRoleFromName(roles, a.Event.Author.ID)

		if r != nil {
			a.Session.GuildRoleDelete(a.GetGuild().ID, r.ID)
		}
	} else {
		a.SendMessage("**<@" + a.Event.Author.ID + "> Too many arguments!")
	}
	return true
}

func GetRoleFromName(slice []*discordgo.Role, value string) *discordgo.Role {
	for _, v := range slice {
		if strings.ToLower(v.Name) == strings.ToLower(value) {
			return v
		}
	}
	return nil
}