package main

import (
	"testing"
	"strings"
	// "launchpad.net/gobson/bson"
	// "time"
	// "launchpad.net/mgo"
	"net"
	"fmt"
	"os"
)

// func TestTest(t *testing.T) {
// 	ifs, err := net.Interfaces()
// 	if err != nil {
// 		fmt.Print(err)
// 	}

// 	for _,v := range ifs {
// 		addrs, _ := v.Addrs(); if len(addrs) == 0 {
// 			continue
// 		}
// 		fmt.Println(v.Index, v.Name)
// 		third := strings.SplitN(addrs[1].String(), ".", 4)[2]

// 		if third == "38" || third == "210" {
// 			fmt.Println("Work Network!")
// 			return
// 		}
// 	}

// 	fmt.Println("Local Network!")
// }

func TestSecondWay(t *testing.T) {
	name, err := os.Hostname() 
	if err != nil { 
		fmt.Printf("Oops: %v\n", err)
		return 
	}
	addrs, err := net.LookupHost(name) 
	if err != nil {
		fmt.Printf("Oops: %v\n", err) 
		return 
	}

	for _,v := range addrs {
		if strings.Contains(v, ".") {
			third := strings.SplitN(v, ".", 4)[2]

			if third == "38" || third == "210" {
				fmt.Println("Work Network!")
				return
			}	
		}
	}
	fmt.Println("Local Network!")
}