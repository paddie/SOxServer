package main

import (
	// "os"
	// "bytes"
	// "path"
	"fmt"
	"labix.org/v2/mgo"
	"net/http"
	// "reflect"
	// "time"
	// old "old/template"
	// "template"
	// "strings"
	// "net"
	// "strconv"
	// "sort"
)

// Returns a .CSV-file to be opened in Excel (or whatever) containing the important
// SOx information.
func soxlist(w http.ResponseWriter, r *http.Request, db *mgo.Database, argPos int) {
	SortKey := r.URL.Path[argPos:]
	if len(SortKey) == 0 {
		SortKey = "hostname"
	}

	c := db.C("machines")

	// var results []map[string]interface{}
	var results []machine
	err := c.Find(nil). //Sort(&map[string]int{SortKey: 1}).
				All(&results)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", "text/csv; charset=utf-8")

	fmt.Fprintf(w, "#\tHostname\tSerial\tIp\tOS (Build)\tRecon\tFirewall\tDate\tModel\tMHz\tRam\tVirus (Definitions)\tLast Virus Scan\tSophos Issue\tOutdated\tSOxIssues\tComment\n")

	for i, doc := range results {
		fmt.Fprintf(w, "%v\t%v\t%v\t%v\t%v\t%v\t%v\t%v\t%v\t%v\t%v\t%v (%v)\t%v\t%v\t%v\n",
			i+1,
			doc.Hostname,
			doc.Serial, // doc["hostname"], 
			doc.Ip,     //doc["ip"],
			doc.Osx,
			doc.Recon,
			doc.Firewall, //["firewall"],
			doc.Now,      // time.NanosecondsToUTC(int64(doc["date"].(bson.Timestamp))),
			doc.Model,    //strings.Replace(doc.Model, ",", ".", -1),
			doc.Cpu,
			doc.Memory,
			doc.Virus_version,
			doc.Virus_def,
			doc.Virus_last_run,
			doc.IsOld(),
			doc.SoxIssues())
	}
}
