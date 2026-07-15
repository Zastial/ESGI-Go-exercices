package main

import (
	"fmt"
	"os"
)

const MaxDisplay = 10

var listArgs = os.Args[1:]

func main() {
	if len(listArgs) == 0 {
		fmt.Println("No arguments provided")
		return
	}

	for i, listArg := range listArgs {
		if i < MaxDisplay && len(listArg) > 4 {
			fmt.Printf("Argument %d: %s\n", i, listArg)
		} else {
			fmt.Printf("Non valid argument at index %d: %s\n", i, listArg)
			os.Exit(1)
		}
	}

}
