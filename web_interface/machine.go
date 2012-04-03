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

type app struct {
    Path string //"path"
    Version string //"version"
    Name string "_name"
    // Info string
}

func (m *app) ShortPath() string {
    const max = 80
    const split = max/2
    if len(m.Path) > max {
        diff := len(m.Path) - max
        return m.Path[:split] + "..." + m.Path[split+diff:]
    }
    return m.Path
}

func (m *app) ShortVersion() string {
    const max = 30
    if len(m.Version) > max {
        return m.Version[:max] + "..."
    }
    return m.Version
}

// private type to handle format conversion from mongo's milisecond time-format, 
type mongotime int64

// time is stored in milliseconds in mongo
// - to get a *time.Time we need to convert milli -> seconds..
func (m mongotime) String() string {
    return fmt.Sprint(time.SecondsToUTC(int64(m)/1e3))
}

// helper struct for the machinelist-view
type machines struct {
    Machines []machine
    Headers []header
}

type machine struct {
    Firewall bool //"firewall"
    Virus_version string //"virus_version"
    Memory string //"memory"
    Virus_last_run string // "virus_last_run"
    Hostname string //"hostname"
    Model string // "model"
    Recon bool //"recon"
    Ip string //"ip"
    Virus_def string  //"virus_def"
    Id string "_id"
    Cpu string //"cpu"
    Osx string //"osx"
    Apps []app //"apps"
    Date mongotime //"date"
    Users []string //"users"
    Cnt int
    // Ignore_firewall bool
}

// helper function to calculate the days since the last update
// - mongo saves time in milliseconds and time.Time operates in either seconds or nanoseconds. Because of this, we divide m.date (int64) with 1000 to convert it into seconds before initialising the time.Time
func (m *machine) TimeOfUpdate() *time.Time {
    return time.SecondsToUTC( int64(m.Date) / 1e3 )
}

func (m *machine) Seconds() int64 {
    return int64(m.Date) / 1e3
}

// calculates the number of days from the last update, to the current date.
func (m *machine) DaysSinceLastUpdate() int64 {
    // seconds in a day: 60^2 * 24 = 86400
    return (time.Seconds() - (int64(m.Date)/1e3)) / 86400
}

// returns true if it is more than 14 days since the machine called home
func (m *machine) IsOld() bool {
    if m.DaysSinceLastUpdate() > 14 {
        return true
    }
    return false
}

// if the machine is a macbook and the firewall is "OFF", we return true
func (m *machine) MacbookFirewallCheck() bool {
    if strings.HasPrefix(m.Model, "MacBook") && !m.Firewall {
        return false
    }
    return true
}

// abstracted into its owm method, since it could prove usefull later. Helper for method 'updateStatus()'
func (m *machine) SoxIssues() bool {
    if m.IsOld() {
        return true
    }
    if !m.Recon {
        return true
    }
    if !m.MacbookFirewallCheck() {
        return true    
    }
    return false
}

// temp url to the specific machine in our system
func (m *machine) Url() string {
    return fmt.Sprintf("/machine/%s", m.Id)
}

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

// The 'name' to be shown in machinelist
// The 'key' to be used when sorting
type header struct {
    Name, Key string
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