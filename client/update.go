package main

import (
	"fmt"
	"os"
	"os/exec"
)

var git_path string

func UpdateRepo(path string) ([]byte, error) {

	cmd := exec.Command(git_path, "pull")
	fmt.Println(cmd.Args)
	return cmd.Output()
}

func Version(path string) {
	cmd := exec.Command(git_path, path, "describe")
	v, err := cmd.Output()

	fmt.Printf("version: %s - err: %s\n", v, err)

}

func ExecuteSOxScripts(path string) ([]byte, error) {
	cmd := exec.Command("/usr/bin/python", path)
	return cmd.Output()
}

func main() {
	var err error
	cmd := exec.Command("which", "git")
	git_path, err = cmd.Output()

	fmt.Printf("%s", git_path)

	// wd, err := os.Getwd()
	os.Chdir("/Library/AdPeople/SOxClient")

	out, err := UpdateRepo("/Library/AdPeople/SOxClient/")
	if err != nil {
		fmt.Println("could not pull:", err)
	}
	fmt.Printf("%s\n", out)
	// Version("/Users/patrick/dev/sox")
	// if err != nil {
	// 	fmt.Println("could not pull:", err)
	// }
	// fmt.Printf("%s\n", out)
	out, err = ExecuteSOxScripts("/Library/AdPeople/SOxClient/sox_sophos.py")
	if err != nil {
		fmt.Println(err)
	}
	fmt.Printf("%s\n", out)
}
