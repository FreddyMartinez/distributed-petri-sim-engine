// This module contains the communication functions
package helpers

import (
	"centralsim"
	"fmt"
	"net"

	"github.com/DistributedClocks/GoVector/govec"
)

// Send any data to desired ip
func Send(data interface{}, ip string, log *centralsim.Logger) error {
	var conn net.Conn
	var err error

	conn, err = net.Dial("tcp", ip)

	if err != nil {
		panic("Client connection error")
	}

	binBuffer := log.GoVec.PrepareSend("Send", data, govec.GetDefaultLogOptions())

	conn.Write(binBuffer)
	defer conn.Close()
	return err
}

// listen for incoming messages
func Receive(data interface{}, listener *net.Listener, log *centralsim.Logger) error {
	var conn net.Conn
	var err error
	tmp := make([]byte, 512)

	conn, err = (*listener).Accept()
	if err != nil {
		panic("Server accept connection error")
	}

	_, err = conn.Read(tmp[0:])
	if err != nil {
		panic(fmt.Sprintf("Server accept connection error: %s", err))
	}

	log.GoVec.UnpackReceive("Receive", tmp, data, govec.GetDefaultLogOptions())

	defer conn.Close()

	return err
}
