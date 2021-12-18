package audio

import (
	"crypto/md5"
	"errors"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

type trackType uint32

const (
	TRACK_WEB_LIVE trackType = iota
	TRACK_WEB_DOWN
	TRACK_LOCAL
)

type Track struct {
	TrackFile string
	TrackName string
	Repeat    bool
	Locked    bool
	Type      trackType
}

func NewTrack(stream string, streamName string, ttype trackType) (Track, error) {
	if ttype == TRACK_WEB_DOWN {
		fmt.Println("Generating hash...")
		h := md5.Sum([]byte(stream))
		var file = "tmp/" + hashToString(h[:]) + ".mp3"

		fmt.Println("Getting name...")
		ynm, err := exec.Command("youtube-dl", "-e", stream).Output()
		if err != nil {
			return Track{}, err
		}

		fmt.Println("Downloading track...")
		_, err = exec.Command("youtube-dl", "--extract-audio", "--audio-format", "mp3", "--output", file, stream).Output()
		if err != nil {
			return Track{}, err
		}

		fmt.Println("Done.")
		var t = Track{file, strings.TrimRight(string(ynm), "\n"), false, false, TRACK_WEB_DOWN}
		return t, nil
	} else if ttype == TRACK_WEB_LIVE {
		var t = Track{stream, streamName, false, false, TRACK_WEB_LIVE}
		return t, nil
	} else if ttype == TRACK_LOCAL {
		var t = Track{stream, streamName, false, false, TRACK_LOCAL}
		return t, nil
	}
	return Track{}, errors.New("Unknown error happened! This should not be returned, invalid trackType?")
}

func hashToString(input []byte) string {
	s := ""
	for _, b := range input {
		s += strconv.Itoa(int(b))
	}
	return s
}
