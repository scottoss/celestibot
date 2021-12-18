package rp

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"equestriaunleashed.com/eclipsingr/celestibot/db"
	"errors"
	"fmt"
	"github.com/boltdb/bolt"
	"github.com/bwmarrin/discordgo"
	"html/template"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

type Authentication struct {
	AuthModeID string `json:"auth_mode_id"`
	Password   string `json:"password"`
}

type RPMetadata struct {
	ServerParent string         `json: "server"`
	RPChannel    string         `json:"rp_channel"`
	Owner        string         `json: "owner"`
	Title        string         `json: "title"`
	Description  string         `json: "description"`
	Members      []string       `json: "member_ids"`
	Auth         Authentication `json:"auth"`
	JoinID       string         `json: "join_id"`
}

type RPCreationInfo struct {
	Auth        Authentication
	Owner       string
	Title       string
	Description string
}

type RPLoginTemplate struct {
	UserName   string
	ServerName string
}

type LoginSession struct {
	ServerID string
	UserID   string
	Key      string
}

func (mtd *RPMetadata) CommitChanges() error {
	db, err := db.Open("celestibot")
	defer db.Close()
	if err != nil {
		return err
	}

	err = db.Database.Update(func(tx *bolt.Tx) error {
		broot, err := tx.CreateBucketIfNotExists([]byte("rp"))
		if err != nil {
			broot = tx.Bucket([]byte("rp"))
		}
		b, err := broot.CreateBucketIfNotExists([]byte(mtd.ServerParent))
		if err != nil {
			b = tx.Bucket([]byte(mtd.ServerParent))
		}
		j, err := json.Marshal(mtd)
		if err != nil {
			b = tx.Bucket([]byte(mtd.ServerParent))
		}
		err = b.Put([]byte(mtd.RPChannel), j)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return err
	}
	return nil
}

func GetRPMetadata(guildid string) (*RPMetadata, error) {
	data := RPMetadata{}
	db, err := db.Open("celestibot")
	defer db.Close()
	if err != nil {
		return nil, err
	}
	err = db.Database.Update(func(tx *bolt.Tx) error {
		broot, err := tx.CreateBucketIfNotExists([]byte("rp"))
		if err != nil {
			broot = tx.Bucket([]byte("rp"))
		}
		b, err := broot.CreateBucketIfNotExists([]byte(guildid))
		if err != nil {
			b = tx.Bucket([]byte(guildid))
		}
		bdata := b.Get([]byte("metadata"))
		if err != nil {
			return err
		}
		err = json.Unmarshal(bdata, &data)
		if err != nil {
			b = tx.Bucket([]byte(guildid))
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &data, nil
}

func CreateRPServer(session *discordgo.Session, info RPCreationInfo) (*RPMetadata, error) {
	// Init base metadata
	data := RPMetadata{"", info.Owner, info.Title, info.Description, []string{info.Owner}, info.Auth, ""}

	//create the guild in question.
	g, err := session.GuildCreate(info.Title)
	if err != nil {
		return nil, errors.New("GUILD ERR: " + err.Error())
	}
	data.Server = g.ID
	ch, err := session.GuildChannelCreate(g.ID, "description", "text")
	if err != nil {
		return nil, err
	}
	session.ChannelMessageSend(ch.ID, info.Description)

	db, err := db.Open("celestibot")
	defer db.Close()
	if err != nil {
		session.GuildDelete(g.ID)
		return nil, err
	}

	hs := sha256.Sum256([]byte(g.ID))
	data.JoinID = base64.StdEncoding.EncodeToString(hs[:])

	err = db.Database.Update(func(tx *bolt.Tx) error {
		broot, err := tx.CreateBucketIfNotExists([]byte("rp"))
		if err != nil {
			broot = tx.Bucket([]byte("rp"))
		}
		b, err := broot.CreateBucketIfNotExists([]byte(g.ID))
		if err != nil {
			b = tx.Bucket([]byte(g.ID))
		}
		j, err := json.Marshal(data)
		if err != nil {
			b = tx.Bucket([]byte(g.ID))
		}
		err = b.Put([]byte("metadata"), j)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		session.GuildDelete(g.ID)
		return nil, err
	}
	return &data, nil
}

func ApplySession(s *discordgo.Session) {
	if session == nil {
		session = s
	}
}

var loginSessions map[string]*LoginSession = make(map[string]*LoginSession)
var mtx sync.Mutex
var session *discordgo.Session

func setSession(name string, session *LoginSession) {
	loginSessions[name] = session
}

func getSession(name string) *LoginSession {
	return loginSessions[name]
}

func CreateSession(serverid, userid string) *LoginSession {
	c := false
	var rnd string = ""
	for !c {
		hs := sha256.Sum256([]byte(strconv.Itoa(rand.Int())))
		rnd = base64.StdEncoding.EncodeToString(hs[:])
		if getSession(rnd) == nil {
			c = true
		}
	}
	s := LoginSession{serverid, userid, rnd}
	setSession(rnd, &s)
	go func() {
		// Timeout
		time.Sleep(time.Minute)
		setSession(rnd, nil)
	}()
	return &s
}

func DebugHandler(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		w.Write([]byte("Invalid request. (err 4) [Invalid form data]"))
		return
	}
	action := r.Form.Get("action")
	if action == "create" {
		owner := r.Form.Get("owner")
		name := r.Form.Get("name")
		desc := r.Form.Get("desc")
		s, err := CreateRPServer(session, RPCreationInfo{Authentication{"password", ""}, owner, name, desc})
		if err != nil {
			w.Write([]byte("<html>Invalid request. (err) [" + err.Error() + "]<br>Listing servers:<br><center>"))
			for i, server := range session.State.Guilds {
				w.Write([]byte(strconv.Itoa(i) + "<br>" + server.Name + " [" + server.ID + "]<br>"))
			}
			w.Write([]byte("</center></html>"))
			return
		}
		sess := CreateSession(s.Server, owner)
		w.Write([]byte("Your Authkey is: " + sess.Key + ", server id is " + sess.ServerID))
	} else if action == "delete" {
		s := r.Form.Get("id")
		_, err := session.GuildDelete(s)
		if err != nil {
			w.Write([]byte("Failed removing server. " + err.Error()))
			return
		}
		w.Write([]byte("Server removed."))
	} else if action == "list" {
		w.Write([]byte("<html>Listing servers:<br><center>"))
		for i, server := range session.State.Guilds {
			w.Write([]byte(strconv.Itoa(i) + "<br>" + server.Name + " [" + server.ID + "]<br>"))
		}
		w.Write([]byte("</center></html>"))
		return
	}
}

func RPServeHandler(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		w.Write([]byte("Invalid request. (err 4) [Invalid form data]"))
		return
	}
	pass := r.PostForm.Get("pass")
	luid := r.Form.Get("luid")
	luid = strings.Replace(luid, " ", "+", -1)
	fmt.Println(luid)
	if pass == "" {
		if getSession(luid) == nil {
			w.Write([]byte("<!DOCTYPE html><html><body>Invalid request. (err 3)<br>Invite timed out.</body></html>"))
			return
		}
		s, err := session.User(getSession(luid).UserID)
		if err != nil {
			w.Write([]byte("Invalid request. (err 2)"))
			return
		}
		g, err := session.Guild(getSession(luid).ServerID)
		if err != nil {
			w.Write([]byte("Invalid request. (err 2)"))
			return
		}
		if getSession(luid).ServerID != g.ID || getSession(luid).UserID != s.ID {
			w.Write([]byte("Invalid request. (err 3)"))
			return
		}

		showPage(w, "login.html", RPLoginTemplate{s.Username, g.Name})
	} else {
		w.Write([]byte("Invalid request. (err 1)"))
		return
	}
}

func showPage(w http.ResponseWriter, fread string, t RPLoginTemplate) error {
	ret, err := template.ParseFiles("res/" + fread)
	if err != nil {
		return err
	}
	ret.Execute(w, t)
	return nil
}

func StartWebServer() {
	http.HandleFunc("/rp/login", RPServeHandler)
	http.HandleFunc("/rp/debug_login", DebugHandler)
	fs := justFilesFilesystem{http.Dir("res/")}
	http.Handle("/res/", http.StripPrefix("/res/", http.FileServer(fs)))
	go http.ListenAndServe(":2048", nil)
}

//From stackoverflow lmao
type justFilesFilesystem struct {
	fs http.FileSystem
}

func (fs justFilesFilesystem) Open(name string) (http.File, error) {
	f, err := fs.fs.Open(name)
	if err != nil {
		return nil, err
	}
	return neuteredReaddirFile{f}, nil
}

type neuteredReaddirFile struct {
	http.File
}

func (f neuteredReaddirFile) Readdir(count int) ([]os.FileInfo, error) {
	return nil, nil
}
