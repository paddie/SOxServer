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
	"text/template"
	// "strings"
	// "net"
	"strconv"
	// "sort"
)

func newLicense(w http.ResponseWriter, r *http.Request, db mgo.Database, argPos int) {
	app := r.FormValue("app")
	path := r.FormValue("path")

	formData := &license{Name: app,
		Path: path}

	set.Execute(w, "newlicense", formData)
}

func addLicense(w http.ResponseWriter, r *http.Request, db mgo.Database, argPos int) {
	app := r.FormValue("app")
	path := r.FormValue("path")
	val := r.FormValue("count")

	fmt.Printf("app:%v, path:%v, count:%v", app, path, val)

	count, err := strconv.Atoi(val)
	if err != nil {
		formData := &license{Name: app, Path: path}

		t := template.Must(template.New("addlicense").ParseFiles("templates/addlicense.html"))
		t.Execute(w, formData)
		return
	}
	c := db.C("machines")
	actual, err := c.Find(bson.M{"apps.path": path}).Count()

	c = db.C("license")

	c.Upsert(bson.M{"path": path},
		bson.M{"name": app,
			"path":         path,
			"max_count":    count,
			"actual_count": actual})

	http.Redirect(w, r, "/licenselist/", 302)
}

func removelicense(w http.ResponseWriter, r *http.Request, db mgo.Database, argPos int) {
	path := r.FormValue("path")
	fmt.Println("Delete", path)
	c := db.C("license")
	err := c.Remove(bson.M{"path": path})

	if err != nil {
		fmt.Print(err)
	}

	http.Redirect(w, r, "/licenselist/", 302)
	return
}

type license struct {
	Name, Path              string
	Max_count, Actual_count int
	Serials                 []string
}

func (l *license) Valid() bool {
	if l.Actual_count <= l.Max_count {
		return true
	}
	return false
}

func licenselist(w http.ResponseWriter, r *http.Request, db mgo.Database, argPos int) {
	var results []license

	err := db.C("license").Find(nil).Sort(bson.M{"name": 1}).All(&results)

	if err != nil {
		fmt.Println(err)
		http.NotFound(w, r)
		return
	}
	set.Execute(w, "licenselist", results)
}
