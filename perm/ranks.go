package perms

import (
	"encoding/json"
	"errors"
	"equestriaunleashed.com/eclipsingr/celestibot/tools"
	"fmt"
	"github.com/boltdb/bolt"
	"equestriaunleashed.com/eclipsingr/celestibot/db"
)

var ranks_all map[string]map[string]*Rank = make(map[string]map[string]*Rank)

var gtn bool = false

func GetRankMap(id string) (map[string]*Rank, error) {
	r := ranks_all[id]
	if ranks_all[id] != nil {
		if r != nil {
			return r, nil
		}
		return nil, errors.New("Rank not found.")
	}
	err := LoadRankMap(id)
	if err != nil {
		return nil, err
	}
	if gtn == false {
		gtn = true
		return GetRankMap(id)
	}
	gtn = false
	return nil, errors.New("GetRankMap goroutine attempted infinite recursion, it has therefore been terminated.")
}

func UpdateLegacyMap(id string) error {
	DB, err := db.Open("celestibot")
	if err != nil {
		return err
	}
	defer DB.Close()
	if tools.FSExists("config/" + id + "/ranks.json") {
		s, err := tools.ReadFromFile("config/" + id + "/ranks.json")
		if err != nil {
			return err
		}
		err = DB.Database.Update(func (tx *bolt.Tx) error {
			_, err := tx.CreateBucketIfNotExists([]byte(id))
			if err != nil {
				return err
			}
			b := tx.Bucket([]byte(id))
			return b.Put([]byte("ranks"), s)
		})
		return err
	}
	return errors.New("File does not exist!")
}

func LoadRankMap(id string) error {
	DB, err := db.Open("celestibot")
	if err != nil {
		fmt.Println(err.Error())
		return err
	}
	defer DB.Close()
	err = DB.Database.Update(func (tx *bolt.Tx) error {

		_, err := tx.CreateBucketIfNotExists([]byte(id))
		if err != nil {
			return err
		}

		b := tx.Bucket([]byte(id))

		s := b.Get([]byte("ranks"))
		fmt.Println("Loaded Rankmap("+id+"):\n"+string(s))
		r := SerializeRanks(string(s))
		if r == nil {
			return errors.New("Could not deserialize JSON.")
		}
		ranks_all[id] = r
		return nil
	})
	return err
}

func HasPermission(guildid, permid string, perm int) bool {
	v, err := GetRankMap(guildid)
	if err != nil {
		fmt.Println(err.Error())
		return false
	}
	if _, ok := v[permid]; ok {
		if v[permid].PermissionLevel >= perm {
			return true
		}
	}
	return false
}

func IsPermission(guildid, permid string) bool {
	v, err := GetRankMap(guildid)
	if err != nil {
		return false
	}
	if _, ok := v[permid]; ok {
		return true
	}
	return false
}

func GetLevel(guildid, permid string) int {
	v, err := GetRankMap(guildid)
	if err != nil {
		return 0
	}
	if _, ok := v[permid]; ok {
		return v[permid].PermissionLevel
	}
	return 0
}

func GetRankFromLevel(guildid string, level int) *Rank {
	va, err := GetRankMap(guildid)
	if err != nil {
		return nil
	}
	for _, k := range va {
		if k.PermissionLevel == level {
			return k
		}
	}
	return nil
}

func GetRankFromID(guildid, roleid string) *Rank {
	va, err := GetRankMap(guildid)
	if err != nil {
		return nil
	}
	for _, k := range va {
		if k.RankID == roleid {
			return k
		}
	}
	return nil
}
// Rank represents a rank on the server
type Rank struct {
	//This is just for readability
	RankName string `json:"rank_name"`

	RankID string `json:"rank_id"`
	PermissionLevel int `json:"permission_level"`
}

func DeserializeRanks(ranks map[string]*Rank) string {
	var txt []byte
	txt, err := json.Marshal(ranks)
	if err != nil {
		tools.LogError(err.Error(), "RankDeserializer")
	}
	return string(txt)
}

func SerializeRanks(inputJson string) map[string]*Rank {
	var ranks map[string]*Rank
	err := json.Unmarshal([]byte(inputJson), &ranks)
	if err != nil {
		tools.LogError(err.Error(), "RankSerializer")
	}
	return ranks
}
