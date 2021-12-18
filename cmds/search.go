package cmds

import (
	"encoding/json"
	"equestriaunleashed.com/eclipsingr/celestibot/core"
	"equestriaunleashed.com/eclipsingr/celestibot/rest"
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"time"
	"github.com/bwmarrin/discordgo"
)

type DerpiResponse struct {
	Search []struct {
		ID               string        `json:"id"`
		CreatedAt        time.Time     `json:"created_at"`
		UpdatedAt        time.Time     `json:"updated_at"`
		DuplicateReports []interface{} `json:"duplicate_reports"`
		FirstSeenAt      time.Time     `json:"first_seen_at"`
		UploaderID       interface{}   `json:"uploader_id"`
		Score            int           `json:"score"`
		CommentCount     int           `json:"comment_count"`
		Width            int           `json:"width"`
		Height           int           `json:"height"`
		FileName         string        `json:"file_name"`
		Description      string        `json:"description"`
		Uploader         string        `json:"uploader"`
		Image            string        `json:"image"`
		Upvotes          int           `json:"upvotes"`
		Downvotes        int           `json:"downvotes"`
		Faves            int           `json:"faves"`
		Tags             string        `json:"tags"`
		TagIds           []string      `json:"tag_ids"`
		AspectRatio      float64       `json:"aspect_ratio"`
		OriginalFormat   string        `json:"original_format"`
		MimeType         string        `json:"mime_type"`
		Sha512Hash       string        `json:"sha512_hash"`
		OrigSha512Hash   string        `json:"orig_sha512_hash"`
		SourceURL        string        `json:"source_url"`
		Representations  struct {
			ThumbTiny  string `json:"thumb_tiny"`
			ThumbSmall string `json:"thumb_small"`
			Thumb      string `json:"thumb"`
			Small      string `json:"small"`
			Medium     string `json:"medium"`
			Large      string `json:"large"`
			Tall       string `json:"tall"`
			Full       string `json:"full"`
		} `json:"representations"`
		IsRendered  bool `json:"is_rendered"`
		IsOptimized bool `json:"is_optimized"`
	} `json:"search"`
	Total        int           `json:"total"`
	Interactions []interface{} `json:"interactions"`
}

var crest = celrest.New()

var blockedterms = []string{"pingas", "suicide", "bruh", "weed", "420", "vore", "sanic", "meme", "big breasts", "suggestive", "porn", "vulgar",
	"mlg", "meme", "dank", "maymays", "exploitable meme", "spoiler", "semi-grimdark", "grimdark", "suggestive", "questionable", "politics"}

func SearchCommand(a core.CommandArgs, v []string) bool {

	tags := core.SliceToString(v)
	for _, term := range blockedterms {
		for _, tag := range v {
			if strings.ToLower(tag) == strings.ToLower(term) {
				a.SendMessage("**Sorry, a tag you tried to search for has been blocked.**")
				return false
			}
		}
	}
	resp := Request(v, 0)
	if resp.Total != 0 && len(resp.Search) != 0 {
		fmt.Println("Image search returned " + strconv.Itoa(resp.Total) + " pages and " + strconv.Itoa(len(resp.Search)) + " elements.")
		pages := resp.Total / len(resp.Search)

		resp = Request(v, rand.Int()%pages)
		si := rand.Int() % len(resp.Search)
		element := resp.Search[si]
		for _, term := range blockedterms {
			t := strings.ToLower(element.Tags)
			if strings.Contains(t, term) {
				if si+1 >= len(resp.Search) {
					a.SendMessage("**Sorry, all images in the current query queue had one or more blocked tags.**")
					return false
				}
				element = resp.Search[si+1]

			}
		}

		embi := discordgo.MessageEmbedImage{
			"https://www." + element.Representations.Medium[2:],
			"https://www." + element.Representations.Medium[2:],
			element.Width/2,
			element.Height/2,
		}
		emba := discordgo.MessageEmbedAuthor{
			element.SourceURL,
			"<Image Source>",
			"",
			"",
		}
		emb := discordgo.MessageEmbed{
			"https://www." + element.Representations.Full[2:],
			"rich",
			"["+element.ID+"] (Right click & Copy Link)]",
			"**Warning: source might be NSFW, Stay safe!\nSearched Tags: " + tags + "**\n" + element.Description,
			"",
			0xfcfbbd,
			nil,
			&embi,
			nil,
			nil,
			nil,
			&emba,
			nil,
		}
		a.SendEmbed(&emb)
	} else {
		a.SendMessage("**No images with the tags: " + tags + " was found. Sorry!**")
		return false
	}
	return true
}

func Request(v []string, page int) DerpiResponse {
	request, data := getReqHeader(v, page)
	out, err := crest.Request(request, data)
	if err != nil {
		fmt.Printf(err.Error())
		return DerpiResponse{}
	}
	var resp DerpiResponse
	json.Unmarshal([]byte(out), &resp)
	return resp
}

func getReqHeader(v []string, page int) (celrest.RestRequest, []celrest.RestRequestData) {
	str := core.SliceToString(v)
	reqh := celrest.RestHeader{
		Tag:   "Accept",
		Value: "application/json",
	}

	req := celrest.RestRequest{
		Header:      reqh,
		AuthToken:   "",
		Directory:   "https://derpibooru.org/search.json",
		RequestType: celrest.R_GET,
	}

	return req, []celrest.RestRequestData{celrest.RestRequestData{
		Tag:   "q",
		Value: str,
	}, celrest.RestRequestData{
		Tag:   "page",
		Value: strconv.Itoa(page),
	}}
}
