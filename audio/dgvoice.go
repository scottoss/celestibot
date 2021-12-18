/*******************************************************************************
 * This is very experimental code and probably a long way from perfect or
 * ideal.  Please provide feed back on areas that would improve performance
 *
 */

// **********************************************************************
// NOTE :: This is getting closer to a Opus<->PCM layer for Discordgo and will
// probably eventually move into a sub-folder of the Discordgo package.
// **********************************************************************

// Package dgvoice provides opus encoding and audio file playback for the
// Discordgo package.
package audio

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io"
	"os/exec"
	"strconv"
	"sync"

	"github.com/bwmarrin/discordgo"
	"github.com/layeh/gopus"
)

// NOTE: This API is not final and these are likely to change.

// Technically the below settings can be adjusted however that poses
// a lot of other problems that are not handled well at this time.
// These below values seem to provide the best overall performance
const (
	channels  int = 2                   // 1 for mono, 2 for stereo
	frameRate int = 48000               // audio sampling rate
	frameSize int = 960                 // uint16 size of each audio frame
	maxBytes  int = (frameSize * 2) * 2 // max size of opus data
)

type Voice struct {
	conn        *discordgo.VoiceConnection
	speakers    map[uint32]*gopus.Decoder
	opusEncoder *gopus.Encoder
	run         *exec.Cmd
	sendpcm     bool
	recvpcm     bool
	recv        chan *discordgo.Packet
	send        chan []int16
	mu          *sync.Mutex
	mup         *sync.RWMutex
	muprl       *sync.Mutex
	vol         float32
	paused      bool
	rlen        int64
	flen        int64
}

func NewVoice(conn *discordgo.VoiceConnection) *Voice {
	return &Voice{conn: conn, muprl: &sync.Mutex{}, mup: &sync.RWMutex{}, mu: &sync.Mutex{}, vol: 0.5}
}

// GetTimeLeft gets the time left of the current stream on the server.
func (v *Voice) GetTimeLeft() (Mins int64, Secs int64) {
	v.muprl.Lock()
	v.mup.Lock()
	defer v.mup.Unlock()
	defer v.muprl.Unlock()
	var sc = float64(((float64(v.rlen) / float64(frameRate)) / 60))
	var mn = float64(sc / 60)
	return int64(mn), int64(sc)
}

// Unpause unpauses the stream on the server
func (v *Voice) Unpause() {
	v.muprl.Lock()
	v.mup.RLock()
	defer v.mup.RUnlock()
	defer v.muprl.Unlock()
	v.paused = false
}

// Pause pauses the stream on the server
func (v *Voice) Pause() {
	v.muprl.Lock()
	v.mup.RLock()
	defer v.mup.RUnlock()
	defer v.muprl.Unlock()
	v.paused = true
}

// Paused returns the pause state of the server
func (v *Voice) Paused() bool {
	v.muprl.Lock()
	v.mup.Lock()
	defer v.mup.Unlock()
	defer v.muprl.Unlock()
	return v.paused
}

// SetVolume sets the volume on the server
func (vo *Voice) SetVolume(volume float32) {
	v := volume
	if v > 1.0 {
		v = 1.0
	} else if v < 0.0 {
		v = 0.0
	}
	vo.vol = v
}

// SendPCM will receive on the provied channel encode
// received PCM data into Opus then send that to Discordgo
func (vo *Voice) SendPCM(pcm <-chan []int16) {

	// make sure this only runs one instance at a time.
	vo.mu.Lock()
	if vo.sendpcm || pcm == nil {
		vo.mu.Unlock()
		return
	}
	vo.sendpcm = true
	vo.mu.Unlock()

	defer func() { vo.sendpcm = false }()

	var err error

	vo.opusEncoder, err = gopus.NewEncoder(frameRate, channels, gopus.Audio)

	if err != nil {
		fmt.Println("NewEncoder Error:", err)
		return
	}

	for {

		// read pcm from chan, exit if channel is closed.
		recv, ok := <-pcm
		if !ok {
			fmt.Println("PCM Channel closed.")
			return
		}

		// try encoding pcm frame with Opus
		opus, err := vo.opusEncoder.Encode(recv, frameSize, maxBytes)
		if err != nil {
			fmt.Println("Encoding Error:", err)
			return
		}

		if vo.conn.Ready == false || vo.conn.OpusSend == nil {
			fmt.Printf("Discordgo not ready for opus packets. %+v : %+v", vo.conn.Ready, vo.conn.OpusSend)
			return
		}
		// send encoded opus data to the sendOpus channel
		vo.conn.OpusSend <- opus
	}
}

// ReceivePCM will receive on the the Discordgo OpusRecv channel and decode
// the opus audio into PCM then send it on the provided channel.
func (vo *Voice) ReceivePCM(c chan *discordgo.Packet) {
	// make sure this only runs one instance at a time.
	vo.mu.Lock()
	if vo.recvpcm || c == nil {
		vo.mu.Unlock()
		return
	}
	vo.recvpcm = true

	vo.mu.Unlock()

	defer func() { vo.sendpcm = false }()
	var err error

	for {

		if vo.conn.Ready == false || vo.conn.OpusRecv == nil {
			fmt.Printf("Discordgo not ready to receive opus packets. %+v : %+v", vo.conn.Ready, vo.conn.OpusRecv)
			return
		}

		p, ok := <-vo.conn.OpusRecv
		if !ok {
			return
		}

		if vo.speakers == nil {
			vo.speakers = make(map[uint32]*gopus.Decoder)
		}

		_, ok = vo.speakers[p.SSRC]
		if !ok {
			vo.speakers[p.SSRC], err = gopus.NewDecoder(48000, 2)
			if err != nil {
				fmt.Println("error creating opus decoder:", err)
				continue
			}
		}

		p.PCM, err = vo.speakers[p.SSRC].Decode(p.Opus, 960, false)
		if err != nil {
			fmt.Println("Error decoding opus data: ", err)
			continue
		}

		c <- p
	}
}

// PlayAudioFile will play the given filename to the already connected
// Discord voice server/channel.  voice websocket and udp socket
// must already be setup before this will work.
func (vo *Voice) PlayAudioFile(filename string) error {

	vo.muprl.Lock()
	vo.mup = &sync.RWMutex{}
	vo.rlen = 0
	vo.muprl.Unlock()

	// Create a shell command "object" to run.
	vo.run = exec.Command("ffmpeg", "-i", filename, "-f", "s16le", "-ar", strconv.Itoa(frameRate), "-ac", strconv.Itoa(channels), "pipe:1")
	defer vo.run.Process.Kill()
	ffmpegout, err := vo.run.StdoutPipe()
	if err != nil {
		fmt.Println("StdoutPipe Error:", err)
		return err
	}

	ffmpegbuf := bufio.NewReaderSize(ffmpegout, 16384)

	// Starts the ffmpeg command
	err = vo.run.Start()
	if err != nil {
		fmt.Println("RunStart Error:", err)
		return err
	}

	// Send "speaking" packet over the voice websocket
	vo.conn.Speaking(true)

	// Send not "speaking" packet over the websocket when we finish
	defer vo.conn.Speaking(false)

	// will actually only spawn one instance, a bit hacky.
	if vo.send == nil {
		vo.send = make(chan []int16, 2)
	}
	go vo.SendPCM(vo.send)

	for {
		if !vo.Paused() {

			// read data from ffmpeg stdout
			audiobuf := make([]int16, frameSize*channels)
			err = binary.Read(ffmpegbuf, binary.LittleEndian, &audiobuf)
			if err == io.EOF || err == io.ErrUnexpectedEOF {
				return nil
			}

			if err != nil {
				fmt.Println("error reading from ffmpeg stdout :", err)
				return err
			}

			vo.muprl.Lock()
			vo.mup.RLock()
			for indx, elm := range audiobuf {
				audiobuf[indx] = int16(float32(elm) * vo.vol)
			}
			vo.mup.RUnlock()
			vo.muprl.Unlock()

			// Send received PCM to the sendPCM channel
			vo.send <- audiobuf

		}
	}
}

// KillPlayer forces the player to stop by killing the ffmpeg cmd process
// this method may be removed later in favor of using chans or bools to
// request a stop.
func (vo *Voice) KillPlayer() {
	vo.run.Process.Kill()
}
