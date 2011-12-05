package main



import (
    "os"
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
)

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
    if strings.HasPrefix(m.Model_id, "MacBook") && !m.Firewall {
        return false
    }
    return true
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
    if !m.Recon {
        return true
    }
    if !m.macbookFirewallCheck() {
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

type app struct {
    Path string "Path"
    Version string "Version"
    Name string "Name"
}

type machine struct {
    Firewall bool "Firewall"
    Virus_version string "Virus_version"
    Memory string "Memory"
    Virus_last_run string "Virus_last_run"
    Hostname string "Hostname"
    Model_id string "Model_id"
    Recon bool "Recon"
    Ip string "Ip"
    Virus_def string "Virus_def"
    Id string "_id"
    Cpu string "Cpu"
    Osx string "Osx"
    Apps []app "Apps"
    Date mongotime "Date"
    Users []string "Users"
    Issue bool
    Cnt int
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
    Hostname string "Hostname"
    Id string "_id"
    Apps []app "Apps"
}

type resultList struct {
    Res []appResult
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
    err := c.Find(bson.M{"Apps.Name" : &bson.RegEx{Pattern:p, Options:"i"}}).
        Select(bson.M{
            "Hostname":1,
            "Apps":1,
            "_id":1}).
        Sort(bson.M{"Hostname":1}).
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

func soxlist(w http.ResponseWriter, r *http.Request, c *mgo.Collection, argPos int) {
    SortKey := r.URL.Path[argPos:]
    if len(SortKey) == 0 {
        SortKey = "Hostname"
    }
    m := new(machines)
    m.Headers = []string{"Hostname","System","Firewall", "Recon","Sophos Antivirus"}
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

    context := new(resultList)
    var res *appResult

    p := `^`+ app_str
    // o := "i"

    fmt.Println("query: ", p)
    // m := bson.M{}    
    err := c.Find(bson.M{"Apps.Name" : &bson.RegEx{Pattern:p, Options:"i"}}).
        Select(bson.M{
            "Hostname":1,
            "Apps":1,
            "_id":1}).
        Sort(bson.M{"Hostname":1}).
        For(&res, func() os.Error {
            // tmp := make([]app, 0, 10)
            // for _, v := range res.Apps {
            //     if v.Name == app_str {
            //         tmp = append(tmp, v)
            //     }
            // }
            res.Apps = filter_apps(app_str, res.Apps)
            // res.Apps = tmp
            context.Res = append(context.Res, *res)
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
    t, err := old.ParseFile(path.Join(wd, "/templates/searchresults.html"), nil)
    if err != nil {
        panic(err)
    }
    t.Execute(w,context)
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
    m.Headers = []string{"#","Hostname", "IP", "System", "Recon", "Firewall", "Sophos Antivirus", "Date", "Model"}
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

    t, err := old.ParseFile(path.Join(wd, "/templates/machinelist.html"), nil)
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
            c := session.DB("sox").C("dict_scripts")
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
    NewHandleFunc("/listapps/", searchAppExact)
	NewHandleFunc("/search/", searchAppSubstring)
    NewHandleFunc("/sox/", soxlist)
    NewHandleFunc("/machine/", machineView)
    NewHandleFunc("/del/", deleteMachine)
    NewHandleFunc("/", machineList)

	http.ListenAndServe(":8080", nil)
}