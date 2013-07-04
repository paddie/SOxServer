package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
	"net/http"
	// "strconv"
	"strings"
	"time"
)

var SophVersionArr []int
var SophVersionString string

// func compareVersionString(version string) []int {
// 	v_arr := strings.Split(version, ".")

// 	var max int

// 	if len(SophVersionArr) == len(v_arr) {
// 		max = len(SophVersionArr)
// 	} else {
// 		if len(SophVersionArr) > len(v_arr) {
// 			max = len(v_arr)
// 		} else {
// 			max = len(SophVersionArr)
// 		}
// 	}

// 	tmp_arr := make([]int, 0, len(v_arr))
// 	for i := 0; i < max; i++ {
// 		n, err := strconv.Atoi(v_arr[i])
// 		if err != nil {
// 			return nil
// 		}

// 	}

// }

type app struct {
	Path    string //"path"
	Version string //"version"
	Name    string `json:"_name"`
}

func (m *app) ShortPath() string {
	const max = 80
	const split = max / 2
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

// helper struct for the machinelist-view
type machines struct {
	Machines []machine
	Headers  []header
}

type machine struct {
	Firewall       bool   //"firewall"
	Virus_version  string //"virus_version"
	Memory         string //"memory"
	Virus_last_run string // "virus_last_run"
	Hostname       string //"hostname"
	Model          string // "model"
	Recon          bool   //"recon"
	Ip             string //"ip"
	Virus_def      string //"virus_def"
	Id             string "_id"
	Cpu            string //"cpu"
	Osx            string //"osx"
	Apps           []app  //"apps"
	Date           string //"date"
	Time           string
	Datetime       int64
	Users          []string //"users"
	Cnt            int
	Serial         string
	Softwareupdate bool
	Softwareoutput string
	// Ignore_firewall bool
}

func (m *machine) SecurityUpdate() bool {
	if strings.Contains(m.Softwareoutput, "Security Update") {
		return true
	}
	return false
}

// helper function to calculate the days since the last update
// - mongo saves time in milliseconds and time.Time operates in either seconds or nanoseconds. Because of this, we divide m.date (int64) with 1000 to convert it into seconds before initialising the time.Time
func (m *machine) TimeOfUpdate() string {
	return m.Date
}

// func (m *machine) Seconds() int {
// 	return int(m.Datetime)
// }

// calculates the number of days from last machine update to now
func (m *machine) DaysSinceLastUpdate() int64 {
	// if it's been more than 2 weeks since the machine responded
	// seconds in a day: 60^2 * 24 = 86400
	return (time.Now().Unix() - m.Datetime) / 86400
}

// returns true if it is more than 14 days since the machine called home
func (m *machine) IsOld() bool {
	// if m.DaysSinceLastUpdate() > 14 {
	// 	return true
	// }
	return false
}

func (m *machine) IsAncient() bool {
	if m.DaysSinceLastUpdate() > 90 {
		return true
	}
	return false
}

// if the machine is a macbook and the firewall is "OFF", we return true
func (m *machine) FirewallIssue() bool {
	if strings.HasPrefix(m.Model, "MacBook") && !m.Firewall {
		return true
	}
	return false
}

// abstracted into its owm method, since it could prove usefull later. Helper for method 'updateStatus()'
func (m *machine) SoxIssues() bool {
	if m.IsAncient() || !m.Recon || m.FirewallIssue() || m.Virus_version == "N/A" {
		return true
	}
	return false
}

func (m *machine) AntivirusIssue() bool {
	if m.Virus_version == "N/A" {
		return true
	}
	return false
}

func (m *machine) SoxWarning() bool {
	if m.IsOld() || m.Softwareupdate {
		return true
	}
	return false

}

// temp url to the specific machine in our system
func (m *machine) Url() string {
	return fmt.Sprintf("/machine/%s", m.Serial)
}

func (m *machine) OldUrl() string {
	return fmt.Sprintf("/oldmachine/%s", m.Serial)
}

/***********************************
view details for each machine
************************************/
func machineView(w http.ResponseWriter, r *http.Request, db *mgo.Database, argPos int) {
	key := r.URL.Path[argPos:]
	if len(key) < 11 {
		http.NotFound(w, r)
		return
	}

	c := db.C("machines")

	var mach *machine
	err := c.Find(bson.M{"_id": key}).
		One(&mach)

	mach.DaysSinceLastUpdate()

	if err != nil {
		fmt.Println(key, err)
		http.NotFound(w, r)
		return
	}
	set.ExecuteTemplate(w, "machine", mach)
}
func oldMachineView(w http.ResponseWriter, r *http.Request, db *mgo.Database, argPos int) {
	key := r.URL.Path[argPos:]
	if len(key) < 11 {
		http.NotFound(w, r)
		return
	}

	c := db.C("old_machines")

	var mach *machine
	err := c.Find(bson.M{"_id": key}).
		One(&mach)

	if err != nil {
		fmt.Println(key, err)
		http.NotFound(w, r)
		return
	}
	set.ExecuteTemplate(w, "machine", mach)
}

/***********************************
delete a machine given machine_id
************************************/
func deleteMachine(w http.ResponseWriter, r *http.Request, db *mgo.Database, argPos int) {
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

	_, err = db.C("old_machines").Upsert(bson.M{"hostname": m.Hostname}, m)

	if err != nil {
		fmt.Print(err)
	}

	err = col_m.Remove(bson.M{"_id": machine_id})

	if err != nil {
		fmt.Print(err)
	}

	http.Redirect(w, r, "/", 302)
	return
}

// The 'name' to be shown in machinelist
// The 'key' to be used when sorting
type header struct {
	Name, Key string
}

// TODO: define which fields are shown using the header-file
func machineList(w http.ResponseWriter, r *http.Request, db *mgo.Database, argPos int) {
	sortKey := r.FormValue("sortkey")
	if sortKey == "" {
		sortKey = "hostname"
	}

	m := new(machines)
	m.Headers = []header{
		{"#", ""},
		{"Hostname", "hostname"},
		{"IP", "ip"},
		{"System", "osx"},
		{"Softwareupdates", "softwareupdate"},
		{"Recon", "recon"},
		{"Firewall", "firewall"},
		{"Sophos Antivirus", ""},
		{"Date", "date"},
		{"Model", "model"},
		{"Memory", "memory"},
		{"Delete", ""}}

	c := db.C("machines")

	var arr *machine
	i := 1
	err := c.Find(nil).Sort(sortKey).
		For(&arr, func() error {
		arr.Cnt = i
		i++
		m.Machines = append(m.Machines, *arr)
		return nil
	})
	if err != nil {
		fmt.Println(err)
		http.NotFound(w, r)
		return
	}
	set.ExecuteTemplate(w, "machinelist", m)
}

func oldmachineList(w http.ResponseWriter, r *http.Request, db *mgo.Database, argPos int) {
	sortKey := r.FormValue("sortkey")
	if sortKey == "" {
		sortKey = "hostname"
	}
	m := new(machines)
	m.Headers = []header{
		{"#", ""},
		{"Hostname", "hostname"},
		{"IP", "ip"},
		{"System", "osx"},
		{"Recon", "recon"},
		{"Firewall", "firewall"},
		{"Sophos Antivirus", ""},
		{"Date", "date"},
		{"Model", "model"},
		{"Memory", "memory"}}

	c := db.C("old_machines")

	var arr *machine
	i := 1
	err := c.Find(nil).
		Sort(sortKey).
		For(&arr, func() error {
		arr.Cnt = i
		i++
		m.Machines = append(m.Machines, *arr)
		return nil
	})
	if err != nil {
		fmt.Println(err)
		http.NotFound(w, r)
		return
	}
	set.ExecuteTemplate(w, "machinelist_old", m)
}

func updateMachine(w http.ResponseWriter, r *http.Request, db *mgo.Database, argPos int) {
	if r.Method != "POST" {
		http.Error(w, "onle accepts POST requests", 405)
	}
	body, err := ioutil.ReadAll(r.Body)
	var m machine
	err = json.Unmarshal(body, &m)
	if err != nil {
		fmt.Println(err)
	}
	m.Id = m.Serial

	m.Softwareoutput = strings.Replace(m.Softwareoutput, "\n", "<br>", -1)

	fmt.Printf("%v %v: Connection from %v - ip: %v\n", m.Date, m.Time, m.Hostname, m.Ip)

	_, err = db.C("machines").UpsertId(m.Id, m)
	if err != nil {
		fmt.Println(err)
	}

	return
}
