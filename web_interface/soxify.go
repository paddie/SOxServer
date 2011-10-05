package main

import (
    "os"
	"fmt"
	"http"
    // "launchpad.net/gobson"
	"launchpad.net/mgo"
    // "reflect"
    "time"
    "template"
    "strings"
)

type apps [][]string

func (a apps) String() string {
    str := ""
    for _,app := range a {
        if len(app) < 3{
            continue
        }
        str += fmt.Sprintf("</h4>%s: %s (%s)</h4><br>", app[0], app[1], app[2])
    }
    return str
}

type machine struct {
    Firewall string "Firewall"
    Virus_version string "Virus_version"
    Memory string "Memory"
    Virus_last_run string "Virus_last_run"
    Hostname string "Hostname"
    Model_id string "Model_id"
    Recon string "Recon"
    Ip string "Ip"
    Virus_def string "Virus_def"
    Id string "_id"
    Cpu string "Cpu"
    Osx string "Osx"
    Apps apps "Apps"
    Date mongotime "Date"
    Users []string "Users"
    Issue bool
    Cnt int
}

// helper function to calculate the days since the last update
// -    the only complicated bit here, is that mongo saves time in milliseconds
//      and everything else operates in seconds or nanoseconds.
//      ´because of this, we devide m.date (int64) with 1000 to convert it into seconds
func (m *machine) DaysSinceLastUpdate() int64 {
    return (time.Seconds() - (int64(m.Date)/1e3)) / 86400
}

// if it is more than 14 days since the machine called home, we return true
func (m *machine) IsOld() bool {
    if m.DaysSinceLastUpdate() > 14 { 
        return true
    }
    return false
}

// if the machine is a macbook and the firewall is "OFF", we return true
func (m *machine) macbookFirewallCheck() bool {
    if strings.HasPrefix(m.Model_id, "MacBook") && m.Firewall != "ON" {
        return true
    }
    return false
}

// returns a time.Time object, calculated from millisecond top seconds
func (m *machine) TimeOfUpdate() *time.Time {
    return time.SecondsToUTC(int64(m.Date)/1e3)
}



// TODO: temp url to the specific machine in our system
func (m *machine) url() string {
    return fmt.Sprintf("/machine/%s", m.Id)
}

// status: if any of the sox parameters are not met, we return true
func (m *machine) updateStatus() {
    if m.hasSoxIssues() {
        m.Issue = true
        return
    }
    m.Issue = false
}

func (m *machine) hasSoxIssues() bool {
    if m.IsOld() {
        return true
    }
    if m.Recon != "Installed" {
        return true
    }
    if m.macbookFirewallCheck() {
        return true    
    }
    return false
}

// private type to handle format conversion,, simple wrapper
type mongotime int64

// time is stored in milliseconds in mongo
// - to get a *time.Time we need to convert milli -> seconds..
func (m mongotime) String() string {
    return fmt.Sprint(time.SecondsToUTC(int64(m)/1e3))
}


func machineView(w http.ResponseWriter, r *http.Request, c *mgo.Collection, argPos int) {
    key := r.URL.Path[argPos:]
    if len(key) < 11 {
        http.NotFound(w,r)
        return
    }
    var mach *machine
    err := c.Find(map[string]string{"_id" : key}).
        One(&mach)
    if err != nil {
        http.NotFound(w,r)
        return
    }
    t, err := template.ParseFile("templates/machine.html",nil)
    if err != nil {
        http.NotFound(w,r)
        return
    }
    mach.updateStatus()
    t.Execute(w,mach)
}

type machines struct {
    Machines []machine
    Headers []string
}

type tableItem struct {
    header string
    value string
}

// TODO: make table-view generic - map[string] string {header:value}
func machineList(w http.ResponseWriter, r *http.Request, c *mgo.Collection, argPos int) {
    SortKey := r.URL.Path[argPos:]
    if len(SortKey) == 0 {
        SortKey = "Hostname"
    }
    m := new(machines)
    // m.Headers = []string{"#","Hostname", "IP", "System", "Recon", "Firewall", "Sophos Antivirus", "Date", "Model"}
    m.Headers = []string{"#","Hostname", "IP", "System", "Firewall", "Sophos Antivirus", "Date", "Model"}   
    var arr *machine
    i := 1    
    err := c.Find(nil).
        Sort(&map[string]int{SortKey:1}).
        For(&arr, func() os.Error {
            arr.updateStatus()
            arr.Cnt = i
            i++
            m.Machines = append(m.Machines, *arr)
            //t.Write(w,arr)
            return nil
        })
    if err != nil {
        http.NotFound(w,r)
        return
    }
    t, err := template.ParseFile("templates/machinelist_new.html", nil)
    if err != nil {
        panic(err)
    }
    t.Execute(w,m)
}

func NewHandleFunc(pattern string,
    fn func(http.ResponseWriter, *http.Request, *mgo.Collection, int)) {
    
    http.HandleFunc(pattern, func(w http.ResponseWriter, r *http.Request) {
            session, err := mgo.Mongo("152.146.38.56")
            if err != nil { 
                panic(err)
            }
            defer session.Close()
            c := session.DB("sox").C("test_script")
            fn(w, r, &c, len(pattern))
        })
}

func writeFixtures(w http.ResponseWriter, r *http.Request, c *mgo.Collection, argPos int) {
    SortKey := r.URL.Path[argPos:]
    if len(SortKey) == 0 {
        SortKey = "Hostname"
    }
    m := new(machines)
    var arr *machine
    i := 1    
    err := c.Find(nil).
        Sort(&map[string]int{SortKey:1}).
        For(&arr, func() os.Error {
            arr.updateStatus()
            arr.Cnt = i
            i++
            m.Machines = append(m.Machines, *arr)
            //t.Write(w,arr)
            return nil
            })
    if err != nil {
        http.NotFound(w,r)
        return
    }
}

func main() {
	NewHandleFunc("/", machineList)
	NewHandleFunc("/machine/", machineView)
	http.ListenAndServe(":8080", nil)
}