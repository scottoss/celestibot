package cmds

import (
	"equestriaunleashed.com/eclipsingr/celestibot/audio"
	"equestriaunleashed.com/eclipsingr/celestibot/core"
	"fmt"
	"strings"
	"equestriaunleashed.com/eclipsingr/celestibot/draw"
)

// SayCommand puts neigh in the chat.
func PlayCommand(a core.CommandArgs, v []string) bool {
	chn, err := a.Session.Channel(a.Event.ChannelID)
	if err != nil {
		a.SendMessage("An error happened while joining [join get guild]: " + err.Error())
	}

	guild, err := a.Session.Guild(chn.GuildID)
	if err != nil {
		a.SendMessage("An error happened while joining [join get guild]: " + err.Error())
	}

	if a.UsedTag == "join" {

		for _, vs := range guild.VoiceStates {
			fmt.Println(vs.ChannelID + "...")
			if vs.UserID == a.Event.Author.ID {
				player, err := audio.NewPlayer(a.Session, guild.ID, vs.ChannelID)
				if err != nil {
					a.SendMessage("An error happened while joining [join create audio]: " + err.Error())
				} else {

					_, err = audio.CreateList(player)
					if err != nil {
						a.SendMessage("An error happened while joining [join create playlist]: " + err.Error())
					}

					fmt.Println("Done.")
					return true
				}
			}
		}
		fmt.Println("Done.")

	}
	if a.UsedTag == "dj" {
		if len(v) > 0 {
			fmt.Println("Getting list...")
			l, exist := audio.Playlists[guild.ID]
			if exist {
				if v[0] == "pause" {
					l.Pause()
					if l.Paused() {
						a.SendMessage("**<@" + a.Event.Author.ID + "> paused the music!**")
					} else {
						a.SendMessage("**<@" + a.Event.Author.ID + "> unpaused the music!**")
					}
				}

				if v[0] == "skip" && (a.Event.Author.ID == "155004943072362496" || a.Event.Author.ID == "220279471285075970"){
					le := len(l.BorrowPlaylist())
					l.UnBorrowPlaylist()
					if le > 0 {
						a.SendMessage("**<@" + a.Event.Author.ID + "> skipped current track [" + l.BorrowPlaylist()[0].TrackName + "]!**")
						l.UnBorrowPlaylist()
						l.Skip()
					}
				}

				if v[0] == "list" {
					file, err := draw.DrawPlaylist(l.BorrowPlaylist()); l.UnBorrowPlaylist()
					if err != nil {
						fmt.Println(err.Error())
					}
					a.SendFile("playlist.png", file)
				}

				if len(v) == 1 {

					fmt.Println("Arguments passed...")

					fmt.Println("NOT A SCAM!...")
					if validStartsWith(v[0], "youtube.com") || validStartsWith(v[0], "youtu.be") || validStartsWith(v[0], "soundcloud.com") {
						fmt.Println("Adding new track...")
						track, err := audio.NewTrack(v[0], "neigh", audio.TRACK_WEB_DOWN)
						if err != nil {
							a.SendMessage("Failed to add track: " + err.Error())
							return false
						}
						l.Add(track)
						a.SendMessage("**<@" + a.Event.Author.ID + "> added " + track.TrackName + " to the playlist!**")
					}
					if a.HasPermissionLevel(RANK_ROYAL_GUARD_LEVEL) {
						if strings.HasPrefix(v[0], "radio->") {
							aud := v[0][len("radio->"):]
							track, err := audio.NewTrack(aud, "livefeed from " + aud, audio.TRACK_WEB_LIVE)
							if err != nil {
								a.SendMessage("Failed to add track: " + err.Error())
								return false
							}
							l.Add(track)
							a.SendMessage("**<@" + a.Event.Author.ID + "> added " + track.TrackName + " to the playlist!**")
						}
					}

				}

			}

		}
	}
	return false
}

func validStartsWith(arg, site string) bool {
	s := strings.ToLower(arg)
	if strings.HasPrefix(s, "https://"+site) || strings.HasPrefix(s, "https://www."+site) {
		return true
	}
	return false
}
