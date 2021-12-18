package cmds

import (
	"equestriaunleashed.com/eclipsingr/celestibot/core"
	"strconv"
	"math/rand"
	"strings"
	"github.com/bwmarrin/discordgo"
	"fmt"
	"time"
)

func DiceCommand(a core.CommandArgs, v []string) bool {
	rand.Seed(time.Now().Unix())
	args := core.SanitizeArgs(v)
	if len(args) > 0 {
		if args[0] == "users" {
			members := a.GetGuild().Members
			if len(args) > 1 {
				roles, err := a.Session.GuildRoles(a.GetGuild().ID)
				if err != nil {
					a.SendMessage("**<@" + a.Event.Author.ID + "> Could not find roles... ??????")
					return false
				}
				names := args[1:]
				var users []*discordgo.User = make([]*discordgo.User, 0)
				var rs []*discordgo.Role = make([]*discordgo.Role, 0)
				for _, n := range names {
					un := n
					if strings.HasPrefix(un, "<@") {
						un = un[2:len(un)-1]
						if strings.HasPrefix(un, "!") {
							un = un[1:]
						}
						fmt.Println("Adding role " + un + "...")
						m := GetMemberFromID(members, un)
						if m != nil {
							users = append(users, m.User)
						}
						continue
					}
					r := GetRoleFromName(roles, un)
					if r != nil {
						fmt.Println("Adding role " + r.Name + "...")
						rs = append(rs, r)
					}
				}
				var userlist []*discordgo.User = users
				for _, m := range members {
					if GetUserHasRole(m, rs) == true {
						fmt.Println("Adding member "+m.User.Username+"...")
						userlist = append(userlist, m.User)
					}
				}
				if len(userlist) > 0 {
					droll := (rand.Int() % len(userlist))
					a.SendMessage("<@" + a.Event.Author.ID + ">'s dice landed on " + strconv.Itoa(droll) + ", which means <@" + userlist[droll].ID + "> won!")
					return true
				}
				a.SendMessage("<@" + a.Event.Author.ID + "> Rolled a dice, but it had no sides! (Try checking your spelling)")
				return false
			}
			droll := (rand.Int() % len(members))
			a.SendMessage("<@" + a.Event.Author.ID + ">'s dice landed on " + strconv.Itoa(droll)+ ", which means <@" + members[droll].User.ID+"> won!")
			return true
		} else {

			if strings.Contains(args[0], "d") {
				split := strings.Split(args[0], "d")
				if len(split) == 2 {
					output := ""
					dices, err := strconv.Atoi(split[0])
					if err != nil {
						a.SendMessage("<@" + a.Event.Author.ID + "> Sorry, please select a numeric value. (for dices)")
						return false
					}
					if dices > 5 {
						a.SendMessage("<@" + a.Event.Author.ID + "> Sorry, that's too many dices. (max 5)")
						return false
					}
					if dices < 1 {
						a.SendMessage("<@" + a.Event.Author.ID + "> Sorry, that's... no dices? (min 1)")
						return false
					}
					sides, err := strconv.Atoi(split[1])
					if err != nil {
						a.SendMessage("<@" + a.Event.Author.ID + "> Sorry, please select a numeric value. (for dices)")
						return false
					}
					output += "**<@"+a.Event.Author.ID+"> rolled "+strconv.Itoa(dices)+" dice(s)!\n"
					for i := 1; i < dices+1; i++ {
						droll := (rand.Int() % sides) + 1
						output += "Dice "+strconv.Itoa(i)+" landed on: " + strconv.Itoa(droll) + "!\n"
					}
					a.SendMessage(output + "**")
				} else {
					a.SendMessage("Invalid format! either choose /roll (dices)d(sides) or /roll (sides)!")
				}
			} else {
				Roll(a, args[0])
			}
		}
	}
	return true
}

func GetMemberFromID(slice []*discordgo.Member, value string) *discordgo.Member {
	for _, v := range slice {
		if v.User.ID == value {
			return v
		}
	}
	return nil
}

func GetUserHasRole(user *discordgo.Member, roles []*discordgo.Role) bool {
	for _, v := range user.Roles {
		for _, val := range roles {
			if v == val.ID {
				return true
			}
		}
	}
	return false
}

func Roll(a core.CommandArgs, v string) bool {
	i, err := strconv.Atoi(v)
	if err != nil {
		a.SendMessage("<@" + a.Event.Author.ID + "> Sorry, please select a numeric value")
		return false
	}
	droll := (rand.Int() % i) + 1
	a.SendMessage("<@" + a.Event.Author.ID + ">'s dice landed on " + strconv.Itoa(droll) + ".")
	return true
}