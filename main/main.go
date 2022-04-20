package main

import (
	"fmt"
	"os/exec"
	"time"
)

func testin() {
	cmd := exec.Command("lsusb")
	stdout, err := cmd.Output()
	if err != nil {
		fmt.Println(err.Error())
	}
	// Print the output
	fmt.Println(string(stdout))
}
func cleanBash() {
	cmd := exec.Command("clear")
	stdout, err := cmd.Output()
	if err != nil {
		fmt.Println(err.Error())
	}
	fmt.Println(string(stdout))
}

func checkinNetwork() {
	cmd := exec.Command("ping", "-c 3", "1.1.1.1")
	_, err := cmd.Output()
	if err != nil {
		fmt.Println("no ping")

	}
	//fmt.Println(string(stdout))
	fmt.Println("ok")
}

func main() {
	for {
		time.Sleep(1 * time.Second)
		checkinNetwork()
		//time.Sleep(1 * time.Second)

	}

}
