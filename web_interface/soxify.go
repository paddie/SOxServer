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
    "time"
    // old "old/template"
    "template"
    "strings"
    // "net"
    "strconv"
    "sort"
)


type appResult struct {
    Hostname string //"hostname"
    Id string "_id"
    Apps []app //"apps"
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
}

type app struct {
    Path string //"path"
    Version string //"version"
    Name string "_name"
    // Info string
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

// Filters apps based on exact name of application
// - includes case
func filter_apps(name string, apps []app) []app {
    tmp := make([]app, 0, 10)
    for _, v := range apps {
        if v.Name == name {
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
func searchAppExact(w http.ResponseWriter, r *http.Request, db mgo.Database, argPos int) {
    key := r.FormValue("key")
    val := r.FormValue("val")

    context := make([]appResult, 0, 10)
    var res *appResult

    c := db.C("machines")

    err := c.Find(bson.M{key : val}).
        Select(bson.M{
            "hostname":1,
            "apps":1,
            "_id":1}).
        Sort(bson.M{"hostname":1}).
        For(&res, func() os.Error {
            res.Apps = filter_apps(app_str, res.Apps)
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

func applications(w http.ResponseWriter, r *http.Request, db mgo.Database, argPos int) {
    var context []string
    // var res *appResult

    c := db.C("machines")

    err := c.Find(nil).Distinct("apps.path", &context)

    fmt.Println(context)

    if err != nil {
        fmt.Println(err)
        http.NotFound(w,r)
        return
    }

    sort.Strings(context)

    set.Execute(w, "applicationlist", &context)
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

    http.Redirect(w,r, "/", 302)
    return
}

// helper struct for the machinelist-view
type machines struct {
    Machines []machine
    Headers []header
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

type black struct {
    Path, Name, Key, Val string
    Count int
}

// BLACKLISTING APPLICATIONS
func blacklist(w http.ResponseWriter, r *http.Request, db mgo.Database, argPos int) {
    // name example: key="apps._name", val="Dropbox"
    // path example: key="apps.path", val="/Applications/Xinet Software/Uploader Manager.app"
    path := r.FormValue("path")
    name := r.FormValue("name")

    app := &black{
        Path: path,
        Name: name
    }
    if strings.Split(val, "/")[1] != "Users" {
        // if application is located in a users folder..
        // we must match on name instead of complete path
        app.Key = "apps._name"
        app.Val = name
    } else {
        app.Key = "apps.path"
        app.Val = path
    }

    db.C("blacklisted").Upsert(bson.M{"key":app.key, "val":app.val}, app)
    
    http.Redirect(w,r, "/blacklisted/", 302)
}

func blacklistView(w http.ResponseWriter, r *http.Request, db mgo.Database, argPos int) {
    
    var bl black[]
    err := db.C("blacklisted").Find(nil).All(&bl)
    if err != nil {
        fmt.Println(err)
        return
    }

    set.Execute(w, )

}

func newLicense(w http.ResponseWriter, r *http.Request, db mgo.Database, argPos int) {
    app := r.FormValue("app")
    path := r.FormValue("path")

    formData := &license{Name:app,
        Path:path}
    
    set.Execute(w,"newlicense", formData)
}

func addLicense(w http.ResponseWriter, r *http.Request, db mgo.Database, argPos int) {
    app := r.FormValue("app")
    path := r.FormValue("path")
    val := r.FormValue("count")

    fmt.Printf("app:%v, path:%v, count:%v", app,path,val)

    count, err := strconv.Atoi(val)
    if err != nil {
        formData := &license{Name:app,Path:path}

        t := template.Must(template.New("addlicense").ParseFile("templates/addlicense.html"))
        t.Execute(w,formData)
        return
    }
    c := db.C("machines")
    actual, err := c.Find(bson.M{"apps.path":path}).Count()

    c = db.C("license")

    c.Upsert(bson.M{"path":path}, 
        bson.M{"name":app,
            "path":path,
            "max_count":count,
            "actual_count":actual})
    
    http.Redirect(w,r, "/licenselist/", 302)
}

func removelicense(w http.ResponseWriter, r *http.Request, db mgo.Database, argPos int) {
    path := r.FormValue("path")
    fmt.Println("Delete", path)
    c := db.C("license")
    err := c.Remove(bson.M{"path":path})
    
    if err != nil {
        fmt.Print(err)
    }

    http.Redirect(w,r, "/licenselist/", 302)
    return
}

type license struct {
    Name, Path string
    Max_count, Actual_count int
    Serials []string
}

func (l *license) Valid() bool{
    if l.Actual_count <= l.Max_count {
        return true
    }
    return false
}

func licenselist(w http.ResponseWriter, r *http.Request, db mgo.Database, argPos int) {
    var results []license

    err := db.C("license").Find(nil).Sort(bson.M{"name":1}).All(&results)
    
    if err != nil {
        fmt.Println(err)
        http.NotFound(w,r)
        return
    }
    set.Execute(w,"licenselist",results)
}

// Returns a .CSV-file to be opened in Excel (or whatever) containing the important
// SOx information.
func soxlist(w http.ResponseWriter, r *http.Request, db mgo.Database, argPos int) {
    SortKey := r.URL.Path[argPos:]
    if len(SortKey) == 0 {
        SortKey = "hostname"
    }

    c := db.C("machines")

    // var results []map[string]interface{}
    var results []machine
    err := c.Find(nil).
        Sort(&map[string]int{SortKey:1}).
        All(&results)
    if err != nil {
        http.NotFound(w,r)
        return
    }
    w.Header().Set("Content-Type","text/csv; charset=utf-8")
        
    fmt.Fprintf(w, "%v,%v,%v,%v,%v,%v,%v,%v,%v\n ",
        "#",
        "Hostname",
        "Ip",
        "OS (Build)",
        "Recon",
        "Firewall",
        "Date",
        "Model",
        "Virus (Definitions)")

    for i,doc := range results {
        fmt.Fprintf(w, "%v,%v (%v),%v,%v,%v,%v,%v,%v,%v (%v)\n",
            i+1,
            doc.Hostname, doc.Id,// doc["hostname"], 
            doc.Ip,//doc["ip"],
            doc.Osx,
            doc.Recon, 
            doc.Firewall, //["firewall"],
            doc.Date, // time.NanosecondsToUTC(int64(doc["date"].(bson.Timestamp))),
            strings.Replace(doc.Model, ",", ".", -1),
            doc.Virus_version, doc.Virus_def)
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
        "templates/machinelist.html"))
    
    NewHandleFunc("/listapps/", searchAppExact)
	NewHandleFunc("/search/", searchAppSubstring)
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

	err := http.ListenAndServe(":8080", nil)
    if err != nil {
        fmt.Println(err)
    }
}