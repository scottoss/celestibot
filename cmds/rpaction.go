package cmds

import (
	"equestriaunleashed.com/eclipsingr/celestibot/core"
	"math/rand"
	"strconv"
	"strings"
)

var actionssingle []string = []string{
	"{0} lept off the ground and {1}ed {2}{3}.",
	"{0} snuck up behind {2} and {1}ed them{3}.",
	"{2} was in the middle of doing something, when {0} suddenly {1}ed them{3}.",
	"What's that? Oh, it's {0} {1}ing {2} {3}! Fascinating...",
}

var actionsmult []string = []string{
	"It seems like {0} pulled {2} into a group {1}{3}",
	"{0} wrote {2} down into a list called \"Pones to {1}{3}\"!",
}

func RpActionCommand(a core.CommandArgs, e []string) bool {
	arg0 := "<@" + a.Event.Author.ID + ">"
	arg1 := a.UsedTag
	arg2 := ""
	msg := ""
	withsomething := ""
	lene := len(e)

	if lene > 1 {
		if !strings.HasPrefix(e[lene-1], "<@") {
			lene = lene - 1
			withsomething = e[lene]
		}
	}

	if (withsomething != "" && lene == 1) || lene == 1 {
		arg2 = e[0]
		if strings.HasSuffix(arg1, "g") {
			arg1 += "g"
		}

		if !strings.HasSuffix(arg1, "op") && !strings.HasSuffix(arg1, "mp") && strings.HasSuffix(arg1, "p") {
			arg1 += "p"
		}

		if strings.HasSuffix(arg1, "b") {
			arg1 += "b"
		}

		if strings.HasSuffix(arg1, "e") {
			arg1 = arg1[:len(arg1)-1]
		}

		if !strings.HasPrefix(arg2, "<@") {
			arg2 = strings.Replace(arg2, "_", " ", -1)
		}

		if !strings.HasPrefix(arg2, "<@") {
			arg2 = strings.Replace(arg2, "_", " ", -1)
		}
		if withsomething != "" {
			withsomething = " with a " + withsomething
		}
		msg = HandleTextAction(actionssingle[rand.Int()%len(actionssingle)], arg0, arg1, arg2, withsomething)

	} else if lene > 1 {

		arg2 = core.SliceToHumanListing(e[:lene])
		if withsomething != "" {
			withsomething = " with a " + withsomething
		}
		msg = HandleTextAction(actionsmult[rand.Int()%len(actionsmult)], arg0, arg1, arg2, withsomething)

	} else {
		me, _ := a.Session.User("@me")
		arg2 = arg0
		arg0 = "<@" + me.ID + ">"

		if strings.HasSuffix(arg1, "g") {
			arg1 += "g"
		}

		if !strings.HasSuffix(arg1, "op") && !strings.HasSuffix(arg1, "mp") && strings.HasSuffix(arg1, "p") {
			arg1 += "p"
		}

		if strings.HasSuffix(arg1, "b") {
			arg1 += "b"
		}

		if strings.HasSuffix(arg1, "e") {
			arg1 = arg1[:len(arg1)-1]
		}

		if !strings.HasPrefix(arg2, "<@") {
			arg2 = strings.Replace(arg2, "_", " ", -1)
		}

		if !strings.HasPrefix(arg2, "<@") {
			arg2 = strings.Replace(arg2, "_", " ", -1)
		}

		if withsomething != "" {
			withsomething = " with a " + withsomething
		}

		msg = HandleTextAction(actionssingle[rand.Int()%len(actionssingle)], arg0, arg1, arg2, withsomething)
	}

	a.SendMessage("**_" + msg + "_**")
	return true
}

func HandleTextAction(input string, inputs ...string) string {
	var final string = input
	for i, input := range inputs {
		final = strings.Replace(final, "{"+strconv.Itoa(i)+"}", input, -1)
	}
	return final
}
