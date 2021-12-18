package audio

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	"sync"
	"time"
)

var Players []*Player = make([]*Player, 0)

var mtx = sync.Mutex{}

type Player struct {
	Player  *Voice
	Guild   *discordgo.Guild
	Channel *discordgo.Channel
}

func GetPlayer(guild string) (success bool, player *Player) {
	mtx.Lock()
	for _, x := range Players {
		if x.Guild.ID == guild {
			return true, x
		}
	}
	defer mtx.Unlock()
	return false, &Player{}
}

func NewPlayer(session *discordgo.Session, guild, channel string) (*Player, error) {
	fmt.Println("Starting voice session on " + guild)
	a, b := GetPlayer(guild)
	if a == true {
		fmt.Println("Voice session found, updating!...")
		mtx.Lock()
		vs, err := session.ChannelVoiceJoin(guild, channel, false, true)
		if err != nil {
			fmt.Printf(err.Error())
			return &Player{}, err
		}

		for !vs.Ready {
			fmt.Println("Voice not ready, waiting 500 msec...")
			time.Sleep(time.Duration(500))
		}
		b.Player = NewVoice(vs)
		defer mtx.Unlock()
		return b, nil
	}

	chn, err := session.Channel(channel)
	if err != nil {
		return &Player{}, err
	}

	gld, err := session.Guild(chn.GuildID)
	if err != nil {
		return &Player{}, err
	}

	fmt.Println("Trying to join...")

	vs, err := session.ChannelVoiceJoin(guild, channel, false, true)
	if err != nil {
		fmt.Printf(err.Error())
		return &Player{}, err
	}

	for !vs.Ready {
		fmt.Println("Voice not ready, waiting 500 msec...")
		time.Sleep(time.Duration(500))
	}
	p := &Player{Player: NewVoice(vs), Channel: chn, Guild: gld}
	mtx.Lock()
	Players = append(Players, p)
	defer mtx.Unlock()
	fmt.Println("Returning and unlocking mutexes...")
	return p, nil
}

func (a *Player) Volume(vol float32) {
	a.Player.SetVolume(vol)
}

func (a *Player) Play(file string) {
	a.Player.PlayAudioFile(file)
}

func (a *Player) Pause() {
	a.Player.Pause()
}


func (a *Player) Paused() bool {
	return a.Player.Paused()
}

func (a *Player) UnPause() {
	a.Player.Unpause()
}

func (a *Player) TogglePause() {
	if a.Player.Paused() {
		a.Player.Unpause()
	} else {
		a.Player.Pause()
	}
}

func (a *Player) Stop() {
	a.Player.KillPlayer()
}

func (a *Player) Disconnect() {
	//a.Player.KillPlayer()
}
