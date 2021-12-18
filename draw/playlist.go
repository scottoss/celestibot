package draw

import (
	"github.com/fogleman/gg"
	"github.com/wcharczuk/go-chart"
	"equestriaunleashed.com/eclipsingr/celestibot/audio"
	"strconv"
	"crypto/md5"
	"math/rand"
)

func a() {
	gg.Scale(0, 0)
	_ = chart.BoxZero
}

func longestName(list []audio.Track) string {
	longest := ""
	for _, n := range list {
		if len(n.TrackName) > len(longest) {
			longest = n.TrackName
		}
	}
	if len(longest) > 82 {
		longest = longest[:79] + "..."
	}
	if longest == "" {
		longest = "No tracks in playlist... :c"
	}
	return longest
}

func DrawPlaylist(list []audio.Track) (string, error) {
	height := 256 + (len(list) * 16)

	var width float64 = 0

	if height > 1024 {
		height = 1024
	}

	if len(list) == 0 {
		height = (64)
	}



	dc := gg.NewContext(512, height)
	err := dc.LoadFontFace("/usr/share/fonts/TTF/Ubuntu-B.ttf", 16)
	if err != nil {
		return "", err
	}
	midpart :=  " ::    "
	width, _ = dc.MeasureString("99" + midpart + longestName(list))

	width = width + 128


	dc = gg.NewContext(int(width), height)

	dc.SetRGB255 (35,39,42)
	dc.Clear()


	dc.SetRGB255 (44,47,51)

	dc.DrawRectangle(0, 32, width, float64(height))
	dc.Fill()

	dc.SetRGB255 (114,137,218)

	err = dc.LoadFontFace("/usr/share/fonts/TTF/Ubuntu-B.ttf", 16)
	if err != nil {
		return "", err
	}

	w, h := dc.MeasureString("Playlist")

	yoff := 42


	dc.DrawString("Playlist", (width/2)-(w/2), 24)

	if len(list) > 0 {

		dc.SetRGB255 (255,255,255)

		err = dc.LoadFontFace("/usr/share/fonts/TTF/Ubuntu-M.ttf", 16)
		if err != nil {
			return "", err
		}

		for i, t := range list {
			if i == 0 {
				dc.DrawString(strconv.Itoa(i)+" ::    "+t.TrackName+ " ▶", 16, h+float64(yoff)+(float64(i)*h))
			} else if (i < 64) {
				dc.DrawString(strconv.Itoa(i)+" ::    "+t.TrackName, 16, h+float64(yoff)+(float64(i)*h) + 4)
			} else {
				dc.DrawString("And more!", 16, h+float64(yoff)+(float64(i)*h))
			}
		}
	} else {
		dc.SetRGB255 (128,128,128)
		w, h := dc.MeasureString("No tracks in playlist... :c")
		dc.DrawString("No tracks in playlist... :c", (width/2)-(w/2), float64(yoff)+(h))
	}
	hash := md5.Sum([]byte(strconv.Itoa(rand.Int())))
	file := "tmp/" + hashToString(hash[:]) + ".png"
	dc.SavePNG(file)
	return file, nil
}

func hashToString(input []byte) string {
	s := ""
	for _, b := range input {
		s += strconv.Itoa(int(b))
	}
	return s
}
