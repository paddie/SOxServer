package main

import (
	"flag"
	"fmt"
	"html/template"
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
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

// var port = flag.String("port", "", "Port on the mongodb server")

func eq(args ...interface{}) bool {
	if len(args) == 0 {
		return false
	}
	x := args[0]
	switch x := x.(type) {
	case string, int, int64, byte, float32, float64:
		for _, y := range args[1:] {
			if x == y {
				return true
			}
		}
		return false
	}

	for _, y := range args[1:] {
		if reflect.DeepEqual(x, y) {
			return true
		}
	}
	return false
}

func main() {
	// load template files, add new templates to this list
	// - remember to {{define "unique_template_name"}} <html> {{end}}
	wd, err := os.Getwd()

	source := filepath.Join(wd, "bootstrap")
	pattern := filepath.Join(wd, "templates", "*.html")
	set = template.Must(template.ParseGlob(pattern))

	set.Funcs(template.FuncMap{"eq": eq})

	// server: 152.146.38.56
	// var ip string
	ip := *flag.String("ip", "localhost", "IP for the MongoDB database eg. 'localhost'")
	fmt.Println("Trying to connect to ", ip)
	session, err = mgo.Dial(ip)
	defer session.Close()
	if err != nil {
		panic(err)
		// fmt.Printf("Failed to connect to MongoDB on '%v'\n", ip)
	}
	fmt.Printf("Connected to MongoDB on '%v'\n", ip)

	err = session.DB("sox").C("machine").EnsureIndexKey("hostname", "time", "date")
	if err != nil {
		panic(err)
	}

	// NewHandleFunc("/searchexact/", searchExact)
	// NewHandleFunc("/ignorefw/", ignorefw)
	// NewHandleFunc("/searchfuzzy/", searchAppSubstring)
	NewHandleFunc("/reportWirelessScan/", wirelessScan)
	NewHandleFunc("/sox/", soxlist)
	NewHandleFunc("/machine/", machineView)
	// NewHandleFunc("/newlicense/", newLicense)
	// NewHandleFunc("/licenselist/", licenselist)
	// NewHandleFunc("/addlicense/", addLicense)
	// NewHandleFunc("/removelicense/", removelicense)
	NewHandleFunc("/del/", deleteMachine)
	NewHandleFunc("/", machineList)
	// NewHandleFunc("/allapps", applications)
	NewHandleFunc("/oldmachines/", oldmachineList)
	NewHandleFunc("/oldmachine/", oldMachineView)
	// NewHandleFunc("/blacklist/", blacklist)
	// NewHandleFunc("/addblacklist/", addBlacklist)
	// NewHandleFunc("/removeblacklist/", removeBlacklist)
	NewHandleFunc("/updateMachine/", updateMachine)

	http.Handle("/assets/", http.StripPrefix("/assets/", http.FileServer(http.Dir(source))))

	err = http.ListenAndServe(":6060", nil)
	if err != nil {
		fmt.Println(err)
	}

}
