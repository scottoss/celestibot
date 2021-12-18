package audio

import (
	"fmt"
	"strconv"
	"sync"
	"time"
	"os"
)

var Playlists map[string]*Playlist = make(map[string]*Playlist, 0)

var mux = sync.Mutex{}

type Playlist struct {
	Tracks    []Track
	Player    *Player
	mutex     sync.Mutex
	destroy   bool
	destroyed bool
}

// Mass Playlist genocide
func KillAll() {
	mux.Lock()
	defer mux.Unlock()
	for i, x := range Playlists {
		fmt.Println("Killing player " + i)
		x.destroy = true
		for !x.destroyed {

		}
	}
	fmt.Println("All has been killed.")
}

func (pl *Playlist) Handle() {
	go func() {
		fmt.Println("PlaylistHandleBegin!...")
		for !pl.destroy {
			if len(pl.Tracks) > 0 {
				fmt.Println("pre-play")
				tf := pl.Tracks[0].TrackFile
				trm := pl.Tracks[0].Type
				fmt.Println("play! " + tf)
				pl.Player.Play(tf)
				if trm == TRACK_LOCAL {
					os.Remove(tf)
				}
				pl.PushNext()
			} else {
				time.Sleep(2 * time.Second)
			}

		}
		pl.destroyed = true
		fmt.Println("Quit Audio Loop!")
	}()
}

func (pl *Playlist) GetCurrentPlaying() Track {
	return pl.Tracks[0]
}

func (pl *Playlist) Stop() {
	pl.destroy = true
}

// AddTrack thread safetly adds a track to the playlist
func (pl *Playlist) Add(track Track) Track {
	pl.mutex.Lock()
	pl.Tracks = append(pl.Tracks, track)
	fmt.Println("New length: " + strconv.Itoa(len(pl.Tracks)))
	pl.mutex.Unlock()
	return track
}


// Skip skips a track in the playlist and plays the next track if possible.
func (pl *Playlist) Pause() Track {
	pl.mutex.Lock()
	if len(pl.Tracks) > 0 {
		pl.Player.TogglePause()
	}
	defer pl.mutex.Unlock()
	return pl.Tracks[0]
}


// Skip skips a track in the playlist and plays the next track if possible.
func (pl *Playlist) Paused() bool {
	pl.mutex.Lock()
	defer pl.mutex.Unlock()
	return pl.Player.Paused()
}

// Skip a track in the playlist and plays the next track if possible.
func (pl *Playlist) Skip(){
	pl.mutex.Lock()
	pl.Player.Stop()
	defer pl.mutex.Unlock()
}

// Skip skips a track in the playlist and plays the next track if possible.
func (pl *Playlist) BorrowPlaylist() []Track {
	pl.mutex.Lock()
	return pl.Tracks
}

func (pl *Playlist) UnBorrowPlaylist() {
	pl.mutex.Unlock()
}

// PullNext pulls the next track in the playlist. (thread safetly)
func (pl *Playlist) Count() int {
	pl.mutex.Lock()
	defer pl.mutex.Unlock()
	return len(pl.Tracks)
}

// PullNext pulls the next track in the playlist. (thread safetly)
func (pl *Playlist) PullNext() Track {
	var t Track = Track{}
	pl.mutex.Lock()
	if len(pl.Tracks) > 0 {
		t = pl.Tracks[1]
	}
	defer pl.mutex.Unlock()
	return t
}

// PushNext pulls the next track in the playlist. (thread safetly)
func (pl *Playlist) PushNext() {
	pl.mutex.Lock()
	if len(pl.Tracks) > 0 {
		pl.Tracks = pl.Tracks[1:]
	}
	defer pl.mutex.Unlock()
}

// CreateList creates a playlist.
func CreateList(player *Player) (*Playlist, error) {
	p := &Playlist{Tracks: make([]Track, 0), Player: player, mutex: *new(sync.Mutex), destroy: false, destroyed: false}
	p.Handle()
	mux.Lock()
	Playlists[player.Guild.ID] = p
	mux.Unlock()
	return p, nil
}
