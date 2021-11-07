package main

import (
	"fmt"
	"os"
	"petrisim/helpers"
	"petrisim/process"
	"strconv"
)

func main() {
	// Obtain process data
	args := os.Args[1:]
	subNetId := args[0]
	networkFile := args[1]

	index, err := strconv.Atoi(subNetId)
	if err != nil {
		panic("Invalid argument when creating process")
	}

	network := helpers.ReadNetConfig(networkFile)

	// create LP
	petriFile := "filename"
	lp := process.CreateLogicProcess(index, network, petriFile)
	fmt.Println(lp)
}
