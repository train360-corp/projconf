package main

import (
	"fmt"
	"github.com/train360-corp/projconf/cmd"
	"os"
)

func main() {
	err := cmd.Run()
	if err != nil {
		if _, err := fmt.Fprintf(os.Stderr, "\n%v\n", err); err != nil {
			panic(err)
		}
		os.Exit(1)
	}
}
