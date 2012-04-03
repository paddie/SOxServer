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