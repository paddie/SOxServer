package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
	"net/http"
	// "strconv"
	// "html/template"
	// "strings"
	"time"
)

var adpeople_owned []string
var approved_insecure []string

func init() {

	approved_insecure = []string{
		"harmony_guest",
		"Surftown Public",
		"101",
	}

	adpeople_owned = []string{
		"symphony_corp",
		"testssid",
		"harmony_guest",
		"nycmadap01",
		"AdPeople Mansion",
		"AdPeople Teepee Ranch",
	}
}

type RSSI struct {
	Hostname string `bson:"_id"`
	RSSI     int
}

type Network struct {
	Hostname string
	Ip       string
	Ssid     string
	Rssi     int
	Rssis    []RSSI
	Sec      []string
	LastSeen time.Time
	ID       string `json:"bssid" bson:"_id"`
}

func (n *Network) Secure() bool {
	if len(n.Sec) == 0 {
		return false
	}

	return true
}

// Issue returns true if the network is open, but not in the either of the two lists of approved items:
// - adpeople_owned and approved_insecure
func (n *Network) Issue() bool {
	if !n.Secure() && !n.AdPeopleOwned() && !n.ApprovedInsecure() {
		return true
	}

	return false
}

// ApprovedInsecure returns true if the ssid is in the approved_insecure []string
func (n *Network) ApprovedInsecure() bool {

	for _, ssid := range approved_insecure {
		if n.Ssid == ssid {
			return true
		}
	}
	return false
}

// AdPeopleOwned returns true if the ssid is in the adpeople_owned []string
func (n *Network) AdPeopleOwned() bool {
	for _, ssid := range adpeople_owned {
		if n.Ssid == ssid {
			return true
		}
	}
	return false
}

func (n *Network) AvgRSSI() float64 {
	cnt, sum := 0, 0
	for _, rssi := range n.Rssis {
		sum += rssi.RSSI
		cnt++
	}

	if cnt == 0 {
		return 0.0
	}

	return float64(sum) / float64(cnt)
}

func listWireless(w http.ResponseWriter, r *http.Request, db *mgo.Database, argPos int) {

	c := db.C("wireless")

	var networks []Network

	var network *Network
	err := c.Find(nil).Sort("ssid").For(&network, func() error {
		networks = append(networks, *network)
		return nil
	})
	if err != nil {
		fmt.Println(err)
		http.Error(w, fmt.Sprintf("mgo: error %v", err), 405)
		return
	}

	if err = set.ExecuteTemplate(w, "wireless", networks); err != nil {
		fmt.Println(err)
	}
}

func wirelessScan(w http.ResponseWriter, r *http.Request, db *mgo.Database, argPos int) {
	if r.Method != "POST" {
		http.Error(w, "only accepts POST requests", 405)
		return
	}

	body, err := ioutil.ReadAll(r.Body)
	var ns []Network
	err = json.Unmarshal(body, &ns)
	if err != nil {
		fmt.Println(err)
		http.Error(w, "only accepts POST requests", 405)
		return
	}

	for _, n := range ns {
		n.LastSeen = time.Now()

		rssi := RSSI{n.Hostname, n.Rssi}

		n.Rssis = []RSSI{rssi}

		if _, err := db.C("wireless").Upsert(
			bson.M{
				"_id": n.ID,
			},
			bson.M{
				"$set": bson.M{
					"ssid":     n.Ssid,
					"hostname": n.Hostname,
					"lastseen": n.LastSeen,
					"sec":      n.Sec,
					"rssi":     n.Rssi,
					"ip":       n.Ip,
				},
			}); err != nil {
			fmt.Println(err)
		}

		err := db.C("wireless").Update(
			bson.M{"_id": n.ID},
			bson.M{
				"$addToSet": bson.M{"rssis": rssi},
			})
		if err != nil {
			fmt.Println(err)
		}
	}
	fmt.Printf("%d accesspoints were upserted\n", len(ns))
}
