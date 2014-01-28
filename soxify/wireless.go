package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"labix.org/v2/mgo"
	// "labix.org/v2/mgo/bson"
	"net/http"
	// "strconv"
	// "html/template"
	// "strings"
	"time"
)

type Network struct {
	SSID     string
	Sec      []string
	LastSeen time.Time
	ID       string `json:"bssid" bson:"_id"`
}

func listWireless(w http.ResponseWriter, r *http.Request, db *mgo.Database, argPos int) {

	c := db.C("wireless")

	var networks []Network

	var network *Network
	err := c.Find(nil).For(network, func() error {
		networks = append(networks, *network)
		return nil
	})
	if err != nil {
		fmt.Println(err)
		http.Error(w, fmt.Sprintf("mgo: error %v", err), 405)
		return
	}

	set.ExecuteTemplate(w, "wireless", networks)
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

	fmt.Printf("Received information on %d accesspoints. Inserting..", len(ns))

	for _, n := range ns {
		n.LastSeen = time.Now()
		_, err = db.C("wireless").UpsertId(n.ID, n)
		if err != nil {
			fmt.Println(err)
		}
	}

	fmt.Println("Done")

}
