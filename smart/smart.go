package smart

import (
	"encoding/json"
	"fmt"
	"strings"
)

var DefTokenConfig = `{
	"is": [
		["what", "WHAT"],
		["it", "IT"],
		["it's", "IT+IS"], ["its", "IS"], ["is", "IS"],
		["cold", "MOD_COLD"], ["hot", "MOD_WARM"], ["warm", "MOD_WARM"],
		["sister", "MOD_RELATIONS"],
		["*", "MOD_WORD"]
	],
	"begins_with": [
		["t'", "IT"],
	],
	"ends_with": [
		["'s", "IS"]
	]
}`

type TokenMap struct {
	Is []struct {
		Match string `json:0`
		Token string `json:"1"`
	} `json:"is"`
	BeginsWith []struct {
		Match string `json:"0"`
		Token string `json:"1"`
	} `json:"begins_with"`
	EndsWith []struct {
		Match string `json:"0"`
		Token string `json:"1"`
	} `json:"ends_with"`
}

type SmartToken struct {
	Type  string
	Value string
}

var tmap TokenMap

func PrepareTokenMap() {
	json.Unmarshal([]byte(DefTokenConfig), &tmap)
	fmt.Printf("%+v\n", tmap)
}

func getToken(tokenstr string) SmartToken {
	for _, e := range tmap.Is {
		if tokenstr == e.Match {
			return SmartToken{Type: e.Token, Value: e.Match}
		}
	}
	for _, e := range tmap.BeginsWith {
		if strings.HasPrefix(tokenstr, e.Match) {
			return SmartToken{Type: e.Token, Value: e.Match}
		}
	}
	for _, e := range tmap.EndsWith {
		if strings.HasSuffix(tokenstr, e.Match) {
			return SmartToken{Type: e.Token, Value: e.Match}
		}
	}
	return SmartToken{Type: "WORD?", Value: tokenstr}
}

func TokenizeString(input string) []SmartToken {
	tokens := make([]SmartToken, 0)
	sents := strings.Split(input, ".")
	if len(sents) == 0 {
		sents = append(sents, input)
	}

	for _, sent := range sents {
		words := strings.Split(sent, " ")
		for _, word := range words {
			lword := strings.ToLower(word)
			tokens = append(tokens, getToken(lword))
		}
	}
	return tokens
}
