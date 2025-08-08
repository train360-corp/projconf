package main

import (
	"github.com/train360-corp/projconf/cmd"
	"log"
)

func main() {
	err := cmd.Run()
	if err != nil {
		log.Fatal(err)
	}
}
