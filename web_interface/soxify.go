package main

import (
    "os"
    "bytes"
    "path"
	"fmt"
	"http"
    "launchpad.net/gobson/bson"
	"launchpad.net/mgo"
    // "reflect"
    "time"
    old "old/template"
    newTemplate "template"
    "strings"
    "net"
)

// private type to handle format conversion from mongo's milisecond time-format, 
type mongotime int64

// time is stored in milliseconds in mongo
// - to get a *time.Time we need to convert milli -> seconds..
func (m mongotime) String() string {
    return fmt.Sprint(time.SecondsToUTC(int64(m)/1e3))
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
    Issue bool
    Cnt int
}

type app struct {
    Path string //"path"
    Version string //"version"
    Name string "_name"
}

// helper function to calculate the days since the last update
// - mongo saves time in milliseconds and time.Time operates in either seconds or nanoseconds. Because of this, we divide m.date (int64) with 1000 to convert it into seconds before initialising the time.Time
func (m *machine) TimeOfUpdate() *time.Time {
    return time.SecondsToUTC( int64(m.Date) / 1e3 )
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

// If any of the sox parameters definned in method 'SoxIssues()' are not met, the m.Issue field of the *machine is updated to true, if not it is set to false
func (m *machine) updateStatus() {
    if m.SoxIssues() {
        m.Issue = true
        return
    }
    m.Issue = false
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

    t := newTemplate.Must(newTemplate.New("machineview").ParseFile("templates/machine.html"))
    if err != nil {
        http.NotFound(w,r)
        return
    }
    mach.updateStatus()
    t.Execute(w,mach)
}

func deleteMachine(w http.ResponseWriter, r *http.Request, c *mgo.Collection, argPos int) {
    machine_id := r.URL.Path[argPos:]
    if len(machine_id) == 0 {
        http.Redirect(w, r, "/", 302)
    }
    fmt.Println("Deleting machine: ", machine_id)
    err := c.Remove(map[string]string{"_id": machine_id})
    if err != nil {
        fmt.Print(err)
    }
    http.Redirect(w,r, "/", 302)
    return
}

type appResult struct {
    Hostname string //"hostname"
    Id string "_id"
    Apps []app //"apps"
}

func filter_apps(substr string, apps []app) []app {
    tmp := make([]app, 0, 10)
    for _, v := range apps {
        if strings.Contains(v.Name, substr) {
            tmp = append(tmp, v)
        }
    }
    return tmp
}

func fuzzyFilter_apps(substr string, apps []app) []app {
    tmp := make([]app, 0, 10)
    for _, v := range apps {
        if strings.Contains(strings.ToLower(v.Name), substr) {
            tmp = append(tmp, v)
        }
    }
    return tmp
}

func searchAppSubstring(w http.ResponseWriter, r *http.Request, c *mgo.Collection, argPos int) {
    app_str := r.FormValue("search")
    // app_str2 := r.FormValue("test2")
    fmt.Println("searching for substring in apps: ", r.Form)

    // context := new(resultList)
    context := make([]appResult, 0, 10)
    var res *appResult

    p := "^.*" + app_str + ".*"

    fmt.Println("query: ", p)
    // m := bson.M{}    
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
    // t := newTemplate.Must(newTemplate.New("results").ParseFile("templates/machine.html"))

    t := newTemplate.Must(newTemplate.New("searchresults").ParseFile("templates/searchresults.html"))

    t.Execute(w,context)
}

type header struct {
    Name, Key string
}

func soxlist(w http.ResponseWriter, r *http.Request, c *mgo.Collection, argPos int) {
    SortKey := r.URL.Path[argPos:]
    if len(SortKey) == 0 {
        SortKey = "hostname"
    }
    m := new(machines)
    m.Headers = []header{{"#",""},
        {"Hostname","hostname"},
        {"System","system"},
        {"Firewall","firewall"},
        {"Recon","recon"},
        {"Sophos Antivirus",""}}
    // m.Headers = []string{"#","Hostname", "IP", "System", "Firewall", "Sophos Antivirus", "Date", "Model"}   
    var arr *machine
    i := 1    
    err := c.Find(nil).
        Sort(&map[string]int{SortKey:1}).
        For(&arr, func() os.Error {
            arr.updateStatus()
            arr.Cnt = i
            i++
            m.Machines = append(m.Machines, *arr)
            return nil
        })
    if err != nil {
        http.NotFound(w,r)
        return
    }

    wd, err := os.Getwd()
    if err != nil {
        panic(err)
    }

    t, err := old.ParseFile(path.Join(wd, "/templates/soxlist.html"), nil)
    if err != nil {
        panic(err)
    }
    t.Execute(w,m)
}


func searchAppExact(w http.ResponseWriter, r *http.Request, c *mgo.Collection, argPos int) {
    app_str := r.URL.Path[argPos:]
    fmt.Println("searching for substring in apps: ", app_str)

    context := make([]appResult, 0, 10)
    var res *appResult

    p := `^`+ app_str
    // o := "i"

    fmt.Println("query: ", p)
    // m := bson.M{}    
    err := c.Find(bson.M{"apps._name" : &bson.RegEx{Pattern:p, Options:"i"}}).
        Select(bson.M{
            "hostname":1,
            "apps":1,
            "_id":1}).
        Sort(bson.M{"hostname":1}).
        For(&res, func() os.Error {
            res.Apps = filter_apps(app_str, res.Apps)
            // res.Apps = tmp
            context = append(context, *res)
            return nil
        })
    
    if err != nil {
        http.NotFound(w,r)
        return
    }

    t := newTemplate.Must(newTemplate.New("searchresults").ParseFile("templates/searchresults.html"))
    t.Execute(w,context)
}

type machines struct {
    Machines []machine
    Headers []header
}

// TODO: make table-view generic - map[string] string {header:value}
func machineList(w http.ResponseWriter, r *http.Request, c *mgo.Collection, argPos int) {
    sortKey := r.URL.Path[argPos:]
    // fmt.Printf("SortKey: %v\n", sortKey)
    if len(sortKey) == 0 {
        sortKey = "hostname"
    }

    // fmt.Printf("SortKey: %v", SortKey)

    m := new(machines)
    m.Headers = []header{{"#",""},{"Hostname","hostname"},{"IP","ip"},{"System","system"},{"Recon","recon"},{"Firewall","firewall"},{"Sophos Antivirus",""}, {"Date","date"}, {"Model","model"}, {"Delete",""}}
    // m.Headers = []string{"#","Hostname", "IP", "System", "Recon", "Firewall", "Sophos Antivirus", "Date", "Model", "Delete"}
    // m.Headers = []string{"#","Hostname", "IP", "System", "Firewall", "Sophos Antivirus", "Date", "Model"}   
    var arr *machine
    i := 1    
    err := c.Find(nil).
        Sort(&map[string]int{sortKey:1}).
        For(&arr, func() os.Error {
            arr.updateStatus()
            arr.Cnt = i
            i++
            m.Machines = append(m.Machines, *arr)
            return nil
        })
    if err != nil {
        http.NotFound(w,r)
        return
    }
    t := newTemplate.Must(newTemplate.New("machinelistt").ParseFile("templates/machinelist.html"))
    // t, err := old.ParseFile(path.Join(wd, "/templates/machinelist.html"), nil)
    t.Execute(w,m)
}

// Backup - not fully implemented yet
func writeFixtures(w http.ResponseWriter, r *http.Request, c *mgo.Collection, argPos int) {
    SortKey := r.URL.Path[argPos:]
    if len(SortKey) == 0 {
        SortKey = "hostname"
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

// Serve files for CSS and JS purposes
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

type myhandler func(http.ResponseWriter, *http.Request, *mgo.Collection, int)

// creates a new handler which creates a session to mongodb
func NewHandleFunc(pattern string, fn myhandler) {

    http.HandleFunc(pattern, func(w http.ResponseWriter, r *http.Request) {
        var ip string
        if isTestLocation() {
            ip = "localhost"
        } else {
            ip = "152.146.38.56"   
        }
        // session, err := mgo.Mongo("152.146.38.56")
        session, err := mgo.Mongo(ip)
        if err != nil { 
            panic(err)
        }
        defer session.Close()
        c := session.DB("sox").C("machines")
        fn(w, r, &c, len(pattern))

    })
}

func isTestLocation() bool {
    name, err := os.Hostname() 
    if err != nil { 
        fmt.Printf("Oops: %v\n", err)
        return true
    }
    // fmt.Println(name)
    addrs, _ := net.LookupHost(name) 
    if err != nil {
        fmt.Printf("Oops: %v\n", err) 
        return true
    }

    for _,v := range addrs {
        if strings.Contains(v, ".") {
            third := strings.SplitN(v, ".", 4)[2]

            if third == "38" || third == "210" {
                fmt.Println("Work Network! Connecting to mongodb on ip: 152.146.38.56")
                return false
            }   
        }
    }
    fmt.Println("Local Network! Connecting to localhost..")

    return true
}


func main() {
    NewHandleFunc("/listapps/", searchAppExact)
	NewHandleFunc("/search/", searchAppSubstring)
    NewHandleFunc("/sox/", soxlist)
    NewHandleFunc("/machine/", machineView)
    NewHandleFunc("/del/", deleteMachine)
    http.HandleFunc("/js/", sourceHandler)
    http.HandleFunc("/bootstrap/", sourceHandler)
    NewHandleFunc("/", machineList)

	http.ListenAndServe(":8080", nil)
}