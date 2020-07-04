package main

import (
	"fmt"
	"io/ioutil"
)

func main() {
	bytes, err := ioutil.ReadFile("c8games/PONG")
	if err != nil {
		fmt.Println("Whoops, couln't read rom")
	}

	fmt.Println(string(bytes[:100]))
}
