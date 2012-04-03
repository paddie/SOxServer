package main

import (
    // "os"
    // "bytes"
    // "path"
	"fmt"
	// "http"
	// "launchpad.net/gobson/bson"
	// "launchpad.net/mgo"
    // "reflect"
    "time"
    // old "old/template"
    // "template"
    "strings"
    // "net"
    // "strconv"
    // "sort"	
)

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
    // Ignore_firewall bool
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