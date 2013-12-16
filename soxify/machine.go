package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
	"net/http"
	// "strconv"
	"html/template"
	"strings"
	"time"
)

var SophVersionArr []int
var SophVersionString string

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

// 'device_names': {
//             'computername':computername.split(".")[0],
//             'hostname':hostname.split(".")[0],
//             'localhostname':localhostname.split(".")[0],
//             'netbiosname':netbios_name.lower(),    
//         },   
type device_names struct {
	Computername  string
	Hostname      string
	Localhostname string
	Netbiosname   string
}

type machine struct {
	Firewall       bool      //"firewall"
	Virus_version  string    //"virus_version"
	Memory         string    //"memory"
	Virus_last_run string    // "virus_last_run"
	Hostname       string    //"hostname"
	Model          string    // "model"
	Recon          bool      //"recon"
	Recon_version  string    // "recon_version"
	Ip             string    //"ip"
	Virus_def      string    //"virus_def"
	Id             string    `json:"_id" bson:"_id"`
	Cpu            string    //"cpu"
	Osx            string    //"osx"
	Apps           []app     //"apps"
	Now            time.Time //"date"
	Device_names   device_names
	// Time           string
	// Datetime       int64
	Users          []string //"users"
	Cnt            int
	Serial         string
	Softwareupdate bool
	Softwareoutput template.HTML
	Script_v       string
	Script_d       string
	// Ignore_firewall bool
}

func MachineListHeaders() []header {
	return []header{
		{"#", ""},
		{"Hostname", "hostname"},
		{"IP", "ip"},
		{"System", "osx"},
		{"Softwareupdates", "softwareupdate"},
		{"Recon", "recon"},
		{"Firewall", "firewall"},
		{"Sophos Antivirus", ""},
		{"Date", "now"},
		{"Model", "model"},
		{"Ram", "memory"},
		{"Script_v", "script_v"},
		{"Delete", ""},
	}
}

func (m *machine) SecurityUpdate() bool {
	// Make sure security updates are registered as critical
	if strings.Contains(string(m.Softwareoutput), "Security Update") {
		return true
	}
	// make sure that Java updates are registered as critical
	if strings.Contains(string(m.Softwareoutput), "Java") {
		return true
	}

	return false
}

// helper function to calculate the days since the last update
// - mongo saves time in milliseconds and time.Time operates in either seconds or nanoseconds. Because of this, we divide m.date (int64) with 1000 to convert it into seconds before initialising the time.Time
func (m *machine) TimeOfUpdate() time.Time {
	return m.Now
}

// func (m *machine) Seconds() int {
// 	return int(m.Datetime)
// }

// calculates the number of days from last machine update to now
func (m *machine) DaysSinceLastUpdate() int64 {
	// if it's been more than 2 weeks since the machine responded
	// seconds in a day: 60^2 * 24 = 86400
	if m.Now.IsZero() {
		return int64(90)
	}

	return int64(time.Now().Sub(m.Now).Hours() / 24)

	// return int64(time.Now().Sub(m.Now).Seconds() / 86400)
}

// returns true if it is more than 14 days since the machine called home
func (m *machine) IsOld() bool {
	if m.DaysSinceLastUpdate() > 14 {
		return true
	}
	return false
}

func (m *machine) IsAncient() bool {
	if m.DaysSinceLastUpdate() > 60 {
		return true
	}
	return false
}

func (m *machine) Date() string {
	return m.Now.Format("01/02/06")
}

// if the machine is a macbook and the firewall is "OFF", we return true
func (m *machine) FirewallIssue() bool {
	if strings.HasPrefix(m.Model, "MacBook") && !m.Firewall {
		return true
	}
	return false
}

// If length of OSX machine name exceeds 15 characters
// the name will be corrupted on: [url](antivirus.yrbrands.com/sox.aspx)
func (m *machine) NameLengthIssue() bool {
	if len(m.Hostname) > 15 {
		return true
	}
	return false
}

func (m *machine) InvalidNetBIOSName() bool {
	if strings.HasPrefix(m.Device_names.Netbiosname, "cph41") ||
		strings.HasPrefix(m.Device_names.Netbiosname, "CPH41") ||
		len(m.Device_names.Netbiosname) == 0 { //ignore if field isn't set
		return false
	}

	return true
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

// abstracted into its owm method, since it could prove usefull later. Helper for method 'updateStatus()'
func (m *machine) SoxIssues() bool {
	if m.IsAncient() ||
		!m.Recon ||
		m.FirewallIssue() ||
		m.Virus_version == "N/A" ||
		m.NameLengthIssue() ||
		m.SecurityUpdate() ||
		m.InvalidNetBIOSName() {
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
		return
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
	m.Headers = MachineListHeaders()

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
		{"Date", "now"},
		{"Model", "model"},
		{"Ram", "memory"}}

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

// this method responds to client updates
// the json data is part of the post request body, we parse
//   it and...
//   1. set the id to the serial
//   2. convert the softwareoutput to html using template.html
//   3. upsert the data into the database
func updateMachine(w http.ResponseWriter, r *http.Request, db *mgo.Database, argPos int) {
	if r.Method != "POST" {
		http.Error(w, "only accepts POST requests", 405)
	}

	body, err := ioutil.ReadAll(r.Body)
	var m machine
	err = json.Unmarshal(body, &m)
	if err != nil {
		fmt.Println(err)
		return
	}

	m.Now = time.Now()
	m.Softwareoutput = template.HTML(strings.Replace(string(m.Softwareoutput), "\n", "<br>", -1))

	fmt.Printf("%v: Connection from %v - ip: %v\n", m.Now, m.Hostname, m.Ip)

	_, err = db.C("machines").UpsertId(m.Id, m)
	if err != nil {
		fmt.Println(err)
	}

	return
}
