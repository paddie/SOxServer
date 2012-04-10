package main

import (
	// "os"
	// "bytes"
	// "path"
	"fmt"
	"launchpad.net/mgo/bson"
	"launchpad.net/mgo"
	"net/http"
	// "reflect"
	// "time"
	// old "old/template"
	// "template"
	"strings"
	// "net"
	// "strconv"
	"sort"
)

func applications(w http.ResponseWriter, r *http.Request, db mgo.Database, argPos int) {
	var context []string
	// var res *appResult
	c := db.C("machines")
	err := c.Find(nil).Distinct("apps.path", &context)

	if err != nil {
		fmt.Println(err)
		http.NotFound(w, r)
		return
	}
	sort.Strings(context)
	set.Execute(w, "applicationlist", &context)
}

type black struct {
	Path, Name, Key, Val string
	Count                int
}

// *****************************************
// BLACKLISTING APPLICATIONS
// *****************************************
func addBlacklist(w http.ResponseWriter, r *http.Request, db mgo.Database, argPos int) {
	// name example: key="apps._name", val="Dropbox"
	// path example: key="apps.path", val="/Applications/Xinet Software/Uploader Manager.app"
	path := r.FormValue("path")
	name := r.FormValue("name")
	app := &black{
		Path: path,
		Name: name}

	if app.Name == "" {
		tmp := strings.Split(path, "/")
		app.Name = strings.Split(tmp[len(tmp)-1], ".")[0]
	}

	if strings.Split(path, "/")[1] == "Users" {
		fmt.Println("blacklisting by name: ", path, "\n\tname: ", name)
		// if application is located in a users folder
		// we must match on name instead of complete path
		app.Key = "apps._name"
		app.Val = app.Name
	} else {
		fmt.Println("blacklisting by path: ", path)
		app.Key = "apps.path"
		app.Val = path
	}
	// doesn't nessesarily need to match on both key AND val..
	db.C("blacklist").Upsert(bson.M{"key": app.Key, "val": app.Val}, app)

	http.Redirect(w, r, "/blacklist/", 302)
}

func removeBlacklist(w http.ResponseWriter, r *http.Request, db mgo.Database, argPos int) {
	key := r.FormValue("key")
	val := r.FormValue("val")

	err := db.C("blacklist").Remove(bson.M{"key": key, "val": val})

	if err != nil {
		fmt.Print(err)
	}

	http.Redirect(w, r, "/blacklist/", 302)
	return
}

func blacklist(w http.ResponseWriter, r *http.Request, db mgo.Database, argPos int) {
	var bl []black
	err := db.C("blacklist").Find(nil).All(&bl)

	if err != nil {
		fmt.Println(err)
		return
	}
	set.Execute(w, "blacklist", bl)
}
