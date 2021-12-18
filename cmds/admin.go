package cmds

import (
	"equestriaunleashed.com/eclipsingr/celestibot/core"
	"equestriaunleashed.com/eclipsingr/celestibot/tools"
	"strconv"
	"equestriaunleashed.com/eclipsingr/celestibot/perm"
	"strings"
)

// SayCommand puts neigh in the chat.
func SetRankCommand(a core.CommandArgs, v []string) bool {
	if a.HandlePermission() {
		if len(v) > 1 {
			person := v[0]
			vv := v[1:]
			rolename := core.SliceToString(vv)
			rank := a.RoleToRank("[" + rolename + "]")
			if rank != nil {
				if (a.GetPermissionLevel() > rank.PermissionLevel) || a.GetGuild().OwnerID == a.Event.Author.ID  {
					gid := a.GetGuild().ID

					aper := person[2:len(person)-1]
					if strings.HasPrefix(aper, "!") {
						aper = aper[1:]
					}

					rm, err := perms.GetRankMap(gid)
					if err != nil {
						a.SendMessage("**Server is not setup correctly, try running /genconfig**")
					}
					for _, rb := range rm {
						a.Session.GuildMemberRoleRemove(gid, aper, rb.RankID)
					}
					a.Session.GuildMemberRoleAdd(gid, aper, rank.RankID)
					a.SendMessage("**Congratulations " + person + ", you have been promoted to " + rolename + "!**")
					return true
				} else {
					a.SendMessage("**<@" + a.Event.Author.ID + "> Rank " + rolename + " Is higher than your current rank, therefore I can not allow you to do this.**")
					return false
				}
			}
			a.SendMessage("**<@" + a.Event.Author.ID + "> Rank " + rolename + " was not found!**")
			return false
		}
		a.SendMessage("**<@" + a.Event.Author.ID + "> Not enough arguments.**")
		return true
	}
	return false
}

const (
	RANK_SUBJECT_LEVEL = 50
	RANK_HONORED_SUBJECT_LEVEL = 100
	RANK_GUARD_LEVEL = 4000
	RANK_ROYAL_GUARD_LEVEL = 5000
	RANK_PRINCESS_LEVEL = 9001
)

var BasicRoleNames []string = []string {"Princess", "Royal Staff", "Royal Guard", "Guard", "Honered Subject", "Subject"}
var BasicRoleColors []int = []int {0x2222BB, 0x00BBBB, 0xBB2222, 0x2244DD, 0xBB55BB, 0xBBBBBB}
var BasicRoleLevels []int = []int {RANK_PRINCESS_LEVEL, RANK_ROYAL_GUARD_LEVEL, RANK_ROYAL_GUARD_LEVEL, RANK_GUARD_LEVEL, RANK_HONORED_SUBJECT_LEVEL, RANK_SUBJECT_LEVEL}

// SayCommand puts neigh in the chat.
func GenConfCommand(a core.CommandArgs, v []string) bool {
	if a.HandlePermission() {
		a.SendMessage("Generating configuration folder...")
		a.SendTyping()
		c, err := a.Session.Channel(a.Event.ChannelID)
		if err != nil {
			return false
		}
		s, err := a.Session.Guild(c.GuildID)
		if err != nil {
			return false
		}

		if !tools.FSExists("config/" + s.ID + "/") {
			err := tools.CreateDirectory("config/" + s.ID + "/")
			if err != nil {
				a.SendMessage("ERROR " + err.Error())
				return false
			}
			a.SendMessage("Created " + "config/" + s.ID + "/")
		} else {
			a.SendMessage("Warning, config directory for this server already exists.")
			a.SendTyping()
		}

		a.SendMessage("Generating ranks...")
		a.SendTyping()
		msg := ""
		var ranks map[string]*perms.Rank = make(map[string]*perms.Rank)
		var Princess *perms.Rank
		for i := len(BasicRoleNames); i > 0; i-- {
			index := len(BasicRoleNames) - i
			role := BasicRoleNames[index]
			r, err := a.Session.GuildRoleCreate(s.ID)
			if err != nil {
				a.SendMessage("ERROR " + err.Error())
				return false
			}
			r.Name = role
			r, err = a.Session.GuildRoleEdit(s.ID, r.ID, "["+role+"]", BasicRoleColors[index], true, r.Permissions, true)
			if err != nil {
				a.SendMessage("ERROR " + err.Error())
				return false
			}
			ranks[r.ID] = &perms.Rank{r.Name, r.ID, BasicRoleLevels[index]}
			if BasicRoleLevels[index] == RANK_PRINCESS_LEVEL {
				Princess = ranks[r.ID]
			}
			msg += "`" + strconv.Itoa(i) + "` - <@&" + r.ID + ">\t`[LVL " + strconv.Itoa(BasicRoleLevels[index]) + "]`\n"

			a.Session.GuildRoleEdit(s.ID, r.ID, r.Name, r.Color, r.Hoist, r.Permissions, false)
		}
		a.SendMessage("**>>Rank Layout<<\n" + msg + "`0` - Everypony (Unranked)\t`[LVL NO LEVEL]`**\n\nWriting configuration...")
		a.SendTyping()
		tools.WriteToFile("config/"+s.ID+"/ranks.json", []byte(perms.DeserializeRanks(ranks)))
		a.SendMessage("Saved configuration to `config/"+s.ID+"/ranks.json`.")
		a.SendMessage("Applying settings...")
		a.Session.GuildMemberRoleAdd(a.GetGuild().ID, a.Event.Author.ID, Princess.RankID)
		a.SendTyping()
		a.SendMessage("Configuration has been completed.")
		return true
	}
	return false
}
