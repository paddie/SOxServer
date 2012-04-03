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

/***********************************
view details for each machine
************************************/
func machineView(w http.ResponseWriter, r *http.Request, db mgo.Database, argPos int) {
    key := r.URL.Path[argPos:]
    if len(key) < 11 {
        http.NotFound(w,r)
        return
    }

    c := db.C("machines")

    var mach *machine
    err := c.Find(bson.M{"_id" : key}).
        One(&mach)

    if err != nil {
        fmt.Println(key, err)
        http.NotFound(w,r)
        return
    }
    set.Execute(w,"machine",mach)
}

/***********************************
delete a machine given machine_id
************************************/
func deleteMachine(w http.ResponseWriter, r *http.Request, db mgo.Database, argPos int) {
    machine_id := r.URL.Path[argPos:]
    if len(machine_id) == 0 {
        http.Redirect(w, r, "/", 302)
    }
    fmt.Println("Deleting machine: ", machine_id)
    col_m := db.C("machines")
    
    var m *machine
    err := col_m.Find(bson.M{"_id": machine_id}).One(&m)
    
    if err != nil {
        fmt.Print(err)
    }

    _, err = db.C("old_machines").Upsert(bson.M{"hostname":m.Hostname}, m)

    if err != nil {
        fmt.Print(err)
    }
    
    err = col_m.Remove(bson.M{"_id": machine_id})

    if err != nil {
        fmt.Print(err)
    }

    http.Redirect(w,r, fmt.Sprintf("/machine/%v", machine_id), 302)
    return
}


// TODO: define which fields are shown using the header-file
func machineList(w http.ResponseWriter, r *http.Request, db mgo.Database, argPos int) {
    sortKey := r.FormValue("sortkey")
    if sortKey == "" {
        sortKey = "hostname"
    }

    m := new(machines)
    m.Headers = []header{
        {"#",""},
        {"Hostname","hostname"},
        {"IP","ip"},
        {"System","osx"},
        {"Recon","recon"},
        {"Firewall","firewall"},
        {"Sophos Antivirus",""},
        {"Date","date"},
        {"Model","model"},
        {"Delete",""}}
    
    c := db.C("machines")

    var arr *machine
    i := 1    
    err := c.Find(nil).
        Sort(&map[string]int{sortKey:1}).
        For(&arr, func() os.Error {
            arr.Cnt = i
            i++
            m.Machines = append(m.Machines, *arr)
            return nil
        })
    if err != nil {
        fmt.Println(err)
        http.NotFound(w,r)
        return
    }
    set.Execute(w,"machinelist", m)
}

func oldmachineList(w http.ResponseWriter, r *http.Request, db mgo.Database, argPos int) {
    sortKey := r.FormValue("sortkey")
    if sortKey == "" {
        sortKey = "hostname"
    }
    m := new(machines)
    m.Headers = []header{
        {"#",""},
        {"Hostname","hostname"},
        {"IP","ip"},
        {"System","osx"},
        {"Recon","recon"},
        {"Firewall","firewall"},
        {"Sophos Antivirus",""},
        {"Date","date"},
        {"Model","model"},
        {"Delete",""}}
    
    c := db.C("old_machines")

    var arr *machine
    i := 1    
    err := c.Find(nil).
        Sort(&map[string]int{sortKey:1}).
        For(&arr, func() os.Error {
            arr.Cnt = i
            i++
            m.Machines = append(m.Machines, *arr)
            return nil
        })
    if err != nil {
        fmt.Println(err)
        http.NotFound(w,r)
        return
    }
    set.Execute(w,"machinelist", m)
}