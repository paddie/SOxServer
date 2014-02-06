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

type RSSI struct {
	Hostname string
	RSSI     int
}

type ReceiveScan struct {
	Hostname string
	Ip       string
	Ssid     string
	Rssi     int
	Sec      []string
	LastSeen time.Time
	ID       string `json:"bssid" bson:"_id"`
}

type ViewScan struct {
	Hostname string
	Ip       string
	Ssid     string
	Rssi     int
	Rssis    []RSSI
	Sec      []string
	LastSeen time.Time
	ID       string `json:"bssid" bson:"_id"`
}

func (n *ViewScan) AvgRSSI() float64 {
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

	var networks []ViewScan

	var network *ViewScan
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
	var ns []ReceiveScan
	err = json.Unmarshal(body, &ns)
	if err != nil {
		fmt.Println(err)
		http.Error(w, "only accepts POST requests", 405)
		return
	}

	for _, n := range ns {
		n.LastSeen = time.Now()

		if _, err = db.C("wireless").UpsertId(n.ID, bson.M{"$set": n}); err != nil {
			fmt.Println(err)
		}
		//  db.wireless.update(
		//		{ "_id" : "0:24:51:5:e2:93"},
		//		{$addToSet: {"rssis":{ "hostname" : "cph41smidtht", "rssi" : -79 }}} )
		_, err = db.C("wireless").Upsert(
			bson.M{"_id": n.ID, "rssis.hostname": n.Hostname},
			bson.M{"$addToSet": bson.M{
				"rssis": RSSI{n.Hostname, n.Rssi}},
			})
		if err != nil {
			fmt.Println(err)
		}
	}
	fmt.Printf("%d accesspoints were upserted\n", len(ns))
}
