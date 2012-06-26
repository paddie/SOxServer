package main

import (
	"os"
	"fmt"
	"launchpad.net/mgo"
	"launchpad.net/mgo/bson"
	"net/http"
	"html/template"
	"path/filepath"
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
	source := filepath.Join(wd, "bootstrap/")
	pattern := filepath.Join(wd, "templates", "*.html")
	set = template.Must(template.ParseGlob(pattern))
	session, err = mgo.Dial("152.146.38.56")
	if err != nil {
		fmt.Println(err)
		fmt.Println("Trying to connect to localhost")
		session, err = mgo.Dial("localhost")
		if err != nil {
			panic(err)
		}
	}

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
	NewHandleFunc("/", machineList)
	NewHandleFunc("/allapps", applications)
	NewHandleFunc("/oldmachines/", oldmachineList)
	NewHandleFunc("/blacklist/", blacklist)
	NewHandleFunc("/addblacklist/", addBlacklist)
	NewHandleFunc("/removeblacklist/", removeBlacklist)
	http.Handle("/assets/", http.StripPrefix("/assets/", http.FileServer(http.Dir(source))))

	err = http.ListenAndServe(":6060", nil)
	if err != nil {
		fmt.Println(err)
	}

}
