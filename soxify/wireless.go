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
	// "time"
)

type Network struct {
	SSID        string
	Sec         []string
	Noise, RSSI int
	ID          string `json:"bssid" bson:"_id"`
}

func wirelessScan(w http.ResponseWriter, r *http.Request, db *mgo.Database, argPos int) {
	if r.Method != "POST" {
		http.Error(w, "only accepts POST requests", 405)
		return
	}

	fmt.Println("wireless scan received..")

	body, err := ioutil.ReadAll(r.Body)
	var n []Network
	err = json.Unmarshal(body, &n)
	if err != nil {
		fmt.Println(err)
		http.Error(w, "only accepts POST requests", 405)
		return
	}

	for _, network := range n {
		_, err = db.C("wireless").UpsertId()
		if err != nil {
			fmt.Println(err)
		}
	}
}
