package main

import (
	"bytes"
	"os"
	// "path"
	"fmt"
	"launchpad.net/mgo"
	"launchpad.net/mgo/bson"
	"net/http"
	// "reflect"
	// "time"
	// old "old/template"
	"html/template"
	"path/filepath"
	// "strings"
	// "net"
	// "strconv"
	// "sort"
)

func ignorefw(w http.ResponseWriter, r *http.Request, db *mgo.Database, argPos int) {
	id := r.FormValue("id")
	if id == "" {
		fmt.Println("ignorefw: no 'id' argument")
		return
	}
	fmt.Println("ignorefw: ignoring firewall for id: ", id)
	err := db.C("machines").Update(bson.M{"_id": id}, bson.M{"$set": bson.M{"ignore_firewall": true}})
	if err != nil {
		fmt.Println(err)
		http.NotFound(w, r)
		return
	}
}

// Serve files for CSS and JS purposes
// TODO: use http.ServeFiles..
func sourceHandler(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if err := recover(); err != nil {
			fmt.Fprintf(w, "%v", err)
		}
	}()
	fmt.Println("load source:", r.URL.Path[1:])
	f, err := os.OpenFile(r.URL.Path[1:], os.O_RDONLY, 0644)
	defer f.Close()
	if err != nil {
		panic(err)
	}
	b := new(bytes.Buffer)
	b.ReadFrom(f)
	fmt.Fprintf(w, b.String())
}

// type-alias to help with the rest
type myhandler func(http.ResponseWriter, *http.Request, *mgo.Database, int)

// builds a new handler that creates a session to mongodb before passing on the function.
func NewHandleFunc(pattern string, fn myhandler) {
	http.HandleFunc(pattern, func(w http.ResponseWriter, r *http.Request) {
		s := session.Copy()
		defer s.Close()

		fn(w, r, s.DB("sox"), len(pattern))

	})
}

var set *template.Template
var session *mgo.Session

func main() {
	// load template files, add new templates to this list
	// - remember to {{define "unique_template_name"}} <html> {{end}}
	wd, err := os.Getwd()
	pattern := filepath.Join(wd, "templates", "*.html")
	fmt.Println("loading templates matching regex: ", pattern)
	set = template.Must(template.ParseGlob(pattern))

	session, err = mgo.Dial("152.146.38.56")
	// session, err = mgo.Dial("127.0.0.1")

	NewHandleFunc("/searchexact/", searchExact)
	NewHandleFunc("/ignorefw/", ignorefw)
	NewHandleFunc("/searchfuzzy/", searchAppSubstring)
	NewHandleFunc("/sox/", soxlist)
	NewHandleFunc("/machine/", machineView)
	NewHandleFunc("/newlicense/", newLicense)
	NewHandleFunc("/licenselist/", licenselist)
	NewHandleFunc("/addlicense/", addLicense)
	NewHandleFunc("/removelicense/", removelicense)
	NewHandleFunc("/del/", deleteMachine)
	http.HandleFunc("/js/", sourceHandler)
	http.HandleFunc("/bootstrap/", sourceHandler)
	NewHandleFunc("/", machineList)
	NewHandleFunc("/allapps", applications)
	NewHandleFunc("/oldmachines/", oldmachineList)
	NewHandleFunc("/blacklist/", blacklist)
	NewHandleFunc("/addblacklist/", addBlacklist)
	NewHandleFunc("/removeblacklist/", removeBlacklist)
	err = http.ListenAndServe(":6060", nil)
	if err != nil {
		fmt.Println(err)
	}
}
