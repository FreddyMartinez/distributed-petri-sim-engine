package main

import (
	"os"
	"petrisim/helpers"
	"petrisim/process"
	"strconv"
	"time"
)

/*
This is where the distributed simulation begins, creating the Logic Process
*/
func main() {
	// Obtain process data
	args := os.Args[1:]
	subNetId := args[0]
	networkFile := "network.json"
	filePrefix := args[1]
	transitionsFile := "tests/" + filePrefix + ".transitions.json"
	petriFile := "tests/" + filePrefix + ".subred" + subNetId + ".json"

	index, err := strconv.Atoi(subNetId)
	if err != nil {
		panic("Invalid argument when creating process")
	}

	network := helpers.ReadNetConfig(networkFile)
	transitionsMap := helpers.ReadNetTransitions(transitionsFile)

	// create LP
	lp := process.CreateLogicProcess(index, network, petriFile, transitionsMap)
	time.Sleep(1 * time.Second) // Espera a que los otros procesos sean creados
	lp.RunSimulation(15)
}
