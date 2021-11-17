package main_test

import (
	"os"
	"petrisim/helpers"
	"strconv"
	"testing"
	"time"
)

// const localPath = "/home/a846866/Documents/"

const localPath = "/home/freedy/Documents/master/RedesYDistribuidos/MiniproyectoSD/petri-net-sim/dist-sim/"

// const binFile = "; ./petrisim "

const binFile = "; ./petrisim "

// const networkFile = "labnetwork.json"

const networkFile = "network.json"

// Used to launch all processes programmatically
func TestMain(m *testing.M) {
	// crear los 3 procesos
	numProcesses, file := test2subnet()

	processes := helpers.ReadNetConfig(networkFile)
	for i, proc := range processes {
		if i == numProcesses {
			break
		}
		sshConn := helpers.CreateSSHClient(proc.Ip)

		command := "cd " + localPath + binFile + strconv.Itoa(i) + " " + file
		go helpers.RunCommand(command, sshConn)

		defer sshConn.Close()
	}
	time.Sleep(10 * time.Second)
	code := m.Run()
	os.Exit(code)
}

func test2subnet() (int, string) {
	return 2, "2sub"
}
