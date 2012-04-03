package main

import (
    "os"
    // "bytes"
    // "path"
    "fmt"
    "http"
    "launchpad.net/gobson/bson"
    "launchpad.net/mgo"
    // "reflect"
    // "time"
    // old "old/template"
    // "template"
    // "strings"
    // "net"
    // "strconv"
    // "sort"
)

// Filters apps based on exact name of application
// - includes case
func filter_apps(value string, apps []app, usePath bool) []app {
    tmp := make([]app, 0, 10)
    if usePath {
        for _, v := range apps {
            if v.Path == value {
                tmp = append(tmp, v)
            }
        }
        return tmp
    }

    for _, v := range apps {
        if v.Name == value {
            tmp = append(tmp, v)
        }
    }
    return tmp
}

// returns the subset of applications whose name contains the substring
// - ignores case
func fuzzyFilter_apps(substr string, apps []app) []app {
    tmp := make([]app, 0, 10)
    name := strings.ToLower(substr)
    for _, v := range apps {
        if strings.Contains(strings.ToLower(v.Name), name) {
            tmp = append(tmp, v)
        }
    }
    return tmp
}

type appResult struct {
    Hostname string //"hostname"
    Id string "_id"
    Apps []app //"apps"
}

/*******************************************************
// queries a list of machines that contain the substring
// - filters using fuzzyFilter_apps
********************************************************/
func searchAppSubstring(w http.ResponseWriter, r *http.Request, db mgo.Database, argPos int) {
    app_str := r.FormValue("search")
    // app_str2 := r.FormValue("test2")
    fmt.Println("searching for substring in apps: ", app_str)

    c := db.C("machines")

    context := make([]appResult, 0, 10)
    var res *appResult

    p := "^.*" + app_str + ".*"

    fmt.Println("query: ", p)
    // m := bson.M{}    
    if len(app_str) == 0 {
        http.NotFound(w,r)
        return
    }
    err := c.Find(bson.M{"apps._name" : &bson.RegEx{Pattern:p, Options:"i"}}).
            Select(bson.M{
                "hostname":1,
                "apps":1,
                "_id":1}).
            Sort(bson.M{"hostname":1}).
            For(&res, func() os.Error {
                res.Apps = fuzzyFilter_apps(app_str, res.Apps)
                context = append(context, *res)
                return nil
            })    
    
    if err != nil {
        http.NotFound(w,r)
        return
    }
    set.Execute(w,"searchresults", context)
}

/********************************************************
// queries a list of machines that has exacly the machine
// - filters using filter_apps
*********************************************************/
func searchExact(w http.ResponseWriter, r *http.Request, db mgo.Database, argPos int) {
    key := r.FormValue("key")
    val := r.FormValue("val")

    context := make([]appResult, 0, 10)
    var res *appResult

    c := db.C("machines")
    var usePath bool
    if key == "apps.path" { usePath = true }

    err := c.Find(bson.M{key : val}).
        Select(bson.M{
            "hostname":1,
            "apps":1,
            "_id":1}).
        Sort(bson.M{"hostname":1}).
        For(&res, func() os.Error {
            res.Apps = filter_apps(val, res.Apps, usePath)
            context = append(context, *res)
            return nil
        })
    
    if err != nil {
        fmt.Println(err)
        http.NotFound(w,r)
        return
    }
    set.Execute(w, "searchresults", context)
}