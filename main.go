package main

import (
	"os"
	"os/signal"
	"strings"
	"syscall"
	"math/rand"
	"github.com/bwmarrin/discordgo"
	"equestriaunleashed.com/eclipsingr/celestibot/core"
	"equestriaunleashed.com/eclipsingr/celestibot/db"
	"equestriaunleashed.com/eclipsingr/celestibot/tools"
	"equestriaunleashed.com/eclipsingr/celestibot/audio"
	"equestriaunleashed.com/eclipsingr/celestibot/cmds"
	"time"
	"equestriaunleashed.com/eclipsingr/celestibot/smart"
	"equestriaunleashed.com/eclipsingr/celestibot/perm"
	"io/ioutil"
	"fmt"
	"equestriaunleashed.com/eclipsingr/celestibot/rp"
)

var dg *discordgo.Session
var err error

var V = &core.Vendor{}

func main() {
	DB, err := db.Open("celestibot")
	if len(os.Args[1:]) > 0 {
		args := os.Args[1:]
		if args[0] == "--registertoken" {
			err := RegisterToken(DB, args[1])
			if err != nil {
				core.LogFatal("Could not write token to databse, reason: " + err.Error(), "DATABASE_TOKEN_SET", 1)
			}
		}

	}
	if err != nil {
		core.LogFatal("Could not open database \"celestibot\", reason: " + err.Error(), "DATABASE_LOAD", 2)
		return
	}

	token, err := Token(DB)
	if err != nil {
		core.LogFatal("Token could not be loaded, reason: "+err.Error(), "DATABASE_TOKEN_GET", 4)
		return
	}
	DB.Close()
	dg, err = discordgo.New("Bot " + token)
	if err != nil {
		core.LogFatal("Discord could not connect, reason: "+err.Error(), "DISCORD_LOAD", 5)
		return
	}
	// Make sure to clear the token from memory.
	token = ""

	core.LogInfo("Seeding rand...", "GOLANG_RAND")
	rand.Seed(time.Now().Unix())

	// Add message handler.
	dg.AddHandler(onMessageRecieve)
	dg.AddHandler(onMessageQue)
	dg.AddHandler(V.Handle)

	err = dg.Open()
	if len(os.Args[1:]) > 0 {
		args := os.Args[1:]
		if args[0] == "--portjson" {
			core.LogInfo("Attempting to port json", "PORT_JSON")
			dirs, _ := ioutil.ReadDir("config/")
			for _, g := range dirs {
				err := perms.UpdateLegacyMap(g.Name())
				core.LogInfo("Porting server " + g.Name() + "...", "PORT_JSON")
				if err != nil {
					core.LogWarning("Failed porting server " + g.Name() + "!: "+err.Error(), "PORT_JSON")
				}
			}
			if err != nil {
				core.LogFatal("Could not port json!: "+err.Error(), "PORT_JSON", 1)
			}
		}
	}
	if err != nil {
		core.LogFatal("Discord could not connect, reason: "+err.Error(), "DISCORD_WS_LOAD", 6)
		return
	}
	AddCommands()
	rp.ApplySession(dg)

	core.LogInfoG("Celestibot Connected! Ctrl-C to exit.", "DISCORD_LOAD")
	rp.StartWebServer()
	// Wait till Ctrl + C is pressed, then close the connection and exit.
	c := make(chan os.Signal, 2)
	signal.Notify(c, os.Interrupt, syscall.SIGINT|syscall.SIGHUP)
	go func() {
		<-c
		onExit()
		os.Exit(0)
	}()
	<-make(chan struct{})
	onExit()
}


func onMessageQue(s *discordgo.Session, event *discordgo.MessageCreate) {
	if strings.ToLower(event.Content) == "woof" || strings.ToLower(event.Content) == "bark" || strings.ToLower(event.Content) == "bork" {
		s.ChannelTyping(event.ChannelID)
		s.ChannelMessageSend(event.ChannelID, "Oh no, i'm allergic to wolfs!")
		s.ChannelTyping(event.ChannelID)
		time.Sleep(time.Second*2)
		s.ChannelMessageSend(event.ChannelID, "**_Sneezes_**")
	}
	u, err := s.User("@me")
	if err != nil {
		s.ChannelMessageSend(event.ChannelID, "Error, i apparently can't find my self.")
		return
	}
	if strings.HasPrefix(strings.ToLower(event.Content), "hello ") ||
		strings.HasPrefix(strings.ToLower(event.Content), "hi ") ||
		strings.HasPrefix(strings.ToLower(event.Content), "heya ") ||
		strings.HasPrefix(strings.ToLower(event.Content), "howdy ") ||
		strings.HasPrefix(strings.ToLower(event.Content), "hey ") ||
		strings.HasPrefix(strings.ToLower(event.Content), "good morning ") ||
		strings.HasPrefix(strings.ToLower(event.Content), "good evening ") ||
		strings.HasPrefix(strings.ToLower(event.Content), "hai ") {

		sel := []string{"Hello", "Hey", "Howdy", "Greetings", "Salutations", "Welcome"}

		if strings.Contains(strings.ToLower(event.Content), "<@" + u.ID + ">") ||
			strings.Contains(strings.ToLower(event.Content), "everypone")  ||
			strings.Contains(strings.ToLower(event.Content), "everybody") ||
			strings.Contains(strings.ToLower(event.Content), "everypony") ||
			strings.Contains(strings.ToLower(event.Content), "everyponi") ||
			strings.Contains(strings.ToLower(event.Content), "all"){
			s.ChannelMessageSend(event.ChannelID, sel[rand.Int()%len(sel)] + ", <@" + event.Author.ID + ">!")
		}

	}
	if strings.HasPrefix(strings.ToLower(event.Content), "im ") ||
		strings.HasPrefix(strings.ToLower(event.Content), "i'm ") ||
		strings.HasPrefix(strings.ToLower(event.Content), "i am ") ||
		strings.HasPrefix(strings.ToLower(event.Content), "am ") ||
		strings.HasPrefix(strings.ToLower(event.Content), "is ") ||
		strings.HasPrefix(strings.ToLower(event.Content),"hug me <@" + u.ID + ">") ||
		strings.HasPrefix(strings.ToLower(event.Content), "hug meh <@" + u.ID + ">") ||
		strings.ToLower(event.Content) == "i need a hug" {
		if (strings.HasPrefix(strings.ToLower(event.Content),"hug me <@" + u.ID + ">") || strings.HasPrefix(strings.ToLower(event.Content),"hug meh <@" + u.ID + ">") || strings.ToLower(event.Content) == "i need a hug") || strings.Contains(strings.ToLower(event.Content), " lonely") {
			var cmd, err = core.GetCommandByTag("hug")
			if err != nil {
				s.ChannelMessageSend(event.ChannelID, "Error, I don't know how to hug.")
				return
			}
			//Hacky code
			us := event.Author
			event.Author = u
			s.ChannelMessageSend(event.ChannelID, "Aww, here!")
			cmd.Callback(core.CommandArgs{Session: s, Event: event, UsedTag: "hug"}, []string{"<@" + us.ID + ">"})
			return
		}

	}

	if strings.ToLower(event.Content) == "can i have a cookie?" {
		if rand.Int()%100 < 25 {
			if rand.Float32() < 0.5 {
				s.ChannelMessageSend(event.ChannelID, "Know what <@"+event.Author.ID+">? You've served equestria well.")
				s.ChannelMessageSend(event.ChannelID, "I hereby bestow opun you this whole cookie.")
				s.ChannelMessageSend(event.ChannelID, "**_Gives <@"+event.Author.ID+"> a cookie_**")

			} else {
				s.ChannelMessageSend(event.ChannelID, "Know what <@"+event.Author.ID+">? I feel generous.")
				s.ChannelMessageSend(event.ChannelID, "**_Gives <@"+event.Author.ID+"> half a cookie_**")
			}

		} else {
			s.ChannelMessageSend(event.ChannelID, "Ahem, <@"+event.Author.ID+"> **NO THEY ARE MINE, SO IS THE CAKE! >:C**")
			s.ChannelMessageSend(event.ChannelID, "**_Aggressivley munches cookies_**")
		}
	}

	if strings.ToLower(event.Content) == "can i have some cake?" {
		if rand.Int()%100 < 45 {
			if rand.Float32() < 0.5 {
				s.ChannelMessageSend(event.ChannelID, "Know what <@"+event.Author.ID+">? You've served equestria well.")
				s.ChannelMessageSend(event.ChannelID, "I hereby bestow upon you this half cake.")
				s.ChannelMessageSend(event.ChannelID, "**_Gives <@"+event.Author.ID+"> a cake_**")

			} else {
				s.ChannelMessageSend(event.ChannelID, "Know what <@"+event.Author.ID+">? I feel generous.")
				s.ChannelMessageSend(event.ChannelID, "**_Cuts out a miniscule slice and gives it to <@"+event.Author.ID+">_**")
			}

		} else {
			s.ChannelMessageSend(event.ChannelID, "**_Looks at <@"+event.Author.ID+"> with a stare that yells NO_**")
		}
	}

	if strings.Trim(strings.ToLower(event.Content), " ") == "🍪" {
		s.ChannelMessageSend(event.ChannelID, "🍪 indeed.")
	}

	if strings.Trim(strings.ToLower(event.Content), " ") == "🍰" {
		s.ChannelMessageSend(event.ChannelID, "That 🍰 is MINE! **_Steals it and noms it_**")
	}
}

func onMessageRecieve(s *discordgo.Session, event *discordgo.MessageCreate) {

	go func(s *discordgo.Session, event *discordgo.MessageCreate) {
		defer func () {
			if r := recover(); r != nil {
				s.ChannelMessageSend(event.ChannelID, "An unhandled error occurred while processing your request.")
				fmt.Println("A fatal error happened and has been recovered from, info:\n", r)
			}
		}()
		if strings.HasPrefix(event.Content, "<@"+s.State.User.ID+">") {
			tokens := smart.TokenizeString(event.Message.Content)
			tkstring := ""
			for _, token := range tokens {
				tkstring += token.Type + " "
			}
			s.ChannelMessageSend(event.ChannelID, tkstring)
		} else {
			if core.GetCommandsLength() > 0 {
				if strings.HasPrefix(event.Content, "/") {
					s.ChannelMessageDelete(event.ChannelID, event.Message.ID)
					var substr = strings.Split(event.Content, " ")
					var dargs = substr[1:]
					var cmdtag = substr[0][1:]
					var cmd, err = core.GetCommandByTag(cmdtag)
					if err != nil {
						s.ChannelMessageSend(event.ChannelID, "`[`<@"+event.Author.ID+">`] Command not found!`")
						return
					}
					tools.LogInfo("<" + event.Author.Username + "> /"+cmdtag + " " + core.SliceToString(dargs), event.Author.Username)
					cmd.Callback(core.NewCMDArgs(s,event, cmdtag, cmd.PermissionLevel), dargs)

				}
			} else {
				tools.LogError("No commands have been implemented.", "CommandHandler")
			}
		}
	}(s, event)
}

func onExit() int {
	audio.KillAll()
	dg.Logout()
	dg.Close()
	return 0
}

// AddCommands adds commands.
func AddCommands() {


	core.AddCommand("admingen", cmds.GenConfCommand, "genconfig", nil, 9000)
	core.AddCommand("adminsetrank", cmds.SetRankCommand, "op", nil, cmds.RANK_ROYAL_GUARD_LEVEL)


	core.AddCommand("say", cmds.SayCommand, "say", nil, cmds.RANK_GUARD_LEVEL)
	core.AddCommand("rpaction", cmds.RpActionCommand, "hug", []string{"cuddle", "boop", "poke", "winghug", "glomp", "nuzzle"}, 0)


	core.AddCommand("dj", cmds.PlayCommand, "join", []string{"dj"}, 0)
	core.AddCommand("color", cmds.ColorCommand, "color", []string{"colour"}, 0)
	core.AddCommand("search-derpi", cmds.SearchCommand, "derpi", []string{"search"}, 0)
	core.AddCommand("dice-roll", cmds.DiceCommand, "dice", []string{"roll"}, 0)

}

func ExitCommand(a core.CommandArgs, v []string) bool {
	onExit()
	os.Exit(0)
	return true
}
