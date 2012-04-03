package main

import (
    "os"
    "bytes"
    // "path"
	"fmt"
	"http"
    "launchpad.net/gobson/bson"
	"launchpad.net/mgo"
    // "reflect"
    // "time"
    // old "old/template"
    "template"
    // "strings"
    // "net"
    // "strconv"
    // "sort"
)

func ignorefw(w http.ResponseWriter, r *http.Request, db mgo.Database, argPos int) {
    id := r.FormValue("id")
    if id == "" {
        fmt.Println("ignorefw: no 'id' argument")
        return
    }
    fmt.Println("ignorefw: ignoring firewall for id: ", id)
    err := db.C("machines").Update(bson.M{"_id" : id}, bson.M{"$set" : bson.M{"ignore_firewall": true}})
    if err != nil {
        fmt.Println(err)
        http.NotFound(w,r)
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
        if err != nil { panic(err) }
        b := new(bytes.Buffer) 
        b.ReadFrom(f) 
        fmt.Fprintf(w, b.String())
}

// type-alias to help with the rest
type myhandler func(http.ResponseWriter, *http.Request, mgo.Database, int)

// builds a new handler that creates a session to mongodb before passing on the function.
func NewHandleFunc(pattern string, fn myhandler) {
    http.HandleFunc(pattern, func(w http.ResponseWriter, r *http.Request) {
        session, err := mgo.Mongo("152.146.38.56")
        if err != nil { 
            panic(err)
        }
        defer session.Close()

        session.SetSyncTimeout(5e9)

        fn(w, r, session.DB("sox"), len(pattern))

    })
}

var set *template.Set

func main() {
    // load template files, add new templates to this list
    // - remember to {{define "unique_template_name"}} <html> {{end}}
    set = template.SetMust(template.ParseSetFiles(
        "templates/base.html", // topbar, top and bottom
        "templates/licenselist.html",
        "templates/newlicense.html",
        "templates/machine.html",
        "templates/searchresults.html",
        "templates/applicationlist.html",
        "templates/blacklist.html",
        "templates/machinelist.html"))
    
    NewHandleFunc("/searchexact/", searchExact)
    NewHandleFunc("/ignorefw/", ignorefw)
	NewHandleFunc("/searchfuzzy/", searchAppSubstring)
    NewHandleFunc("/sox/", soxlist)
    NewHandleFunc("/machine/", machineView)
    NewHandleFunc("/newlicense/",newLicense)
    NewHandleFunc("/licenselist/",licenselist)
    NewHandleFunc("/addlicense/",addLicense)
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
	err := http.ListenAndServe(":8080", nil)
    if err != nil {
        fmt.Println(err)
    }

    
}