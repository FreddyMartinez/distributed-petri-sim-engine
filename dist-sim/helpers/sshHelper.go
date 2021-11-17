// This module is intended to contain ssh functions,
// for automation and deployment tasks.
// Inspired in https://blog.tarkalabs.com/ssh-recipes-in-go-part-one-5f5a44417282
// And http://networkbit.ch/golang-ssh-client/
package helpers

import (
	"fmt"
	"io/ioutil"
	"os"

	"golang.org/x/crypto/ssh"
)

const SshPort = "22"
const PrivateKeyPath = "/home/freedy/.ssh/id_rsa"

const user = "freedy"

// const user = "a846866"

func CreateSSHClient(host string) *ssh.Client {

	config := &ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{
			getPublicKey(),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		HostKeyAlgorithms: []string{
			ssh.KeyAlgoRSA,
			ssh.KeyAlgoDSA,
			ssh.KeyAlgoECDSA256,
			ssh.KeyAlgoECDSA384,
			ssh.KeyAlgoECDSA521,
			ssh.KeyAlgoED25519,
		},
	}

	client, err := ssh.Dial("tcp", host+":"+SshPort, config)
	if err != nil {
		panic(err)
	}
	return client
}

// Taken as is from the first link
func getPublicKey() ssh.AuthMethod {
	key, err := ioutil.ReadFile(PrivateKeyPath)
	if err != nil {
		panic(err)
	}
	signer, err := ssh.ParsePrivateKey(key)
	if err != nil {
		panic(err)
	}
	return ssh.PublicKeys(signer)
}

// Taken as is from the first link
func RunCommand(cmd string, conn *ssh.Client) {
	sess, err := conn.NewSession()
	if err != nil {
		panic(err)
	}
	defer sess.Close()

	// setup standard out and error
	sess.Stdout = os.Stdout
	sess.Stderr = os.Stderr
	//fmt.Println(cmd)
	err = sess.Run(cmd)
	if err != nil {
		fmt.Println(cmd)
		panic(err)
	}
}
