// This module contains the communication functions
package helpers

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"net"
)

// Send any data to desired ip
func Send(data interface{}, ip string) error {
	var conn net.Conn
	var err error
	var encoder *gob.Encoder

	conn, err = net.Dial("tcp", ip)

	if err != nil {
		panic("Client connection error")
	}

	binBuffer := new(bytes.Buffer)

	encoder = gob.NewEncoder(binBuffer)
	err = encoder.Encode(data)

	conn.Write(binBuffer.Bytes())
	defer conn.Close()
	return err
}

// listen for incoming messages
func Receive(data interface{}, listener *net.Listener) error {
	var conn net.Conn
	var err error
	tmp := make([]byte, 512)
	tmpbuff := bytes.NewBuffer(tmp)

	conn, err = (*listener).Accept()
	if err != nil {
		panic("Server accept connection error")
	}

	_, err = conn.Read(tmp[0:])
	if err != nil {
		panic(fmt.Sprintf("Server accept connection error: %s", err))
	}

	decoder := gob.NewDecoder(tmpbuff)

	err = decoder.Decode(data)
	defer conn.Close()

	return err
}
