package main

import (
	"bytes"
	"fmt"
	"log"
	"os/exec"
	"time"

	"golang.org/x/crypto/ssh"
)

func main() {
	fmt.Println("Hello World!")

	out, err := exec.Command("cmd", "/k", "dir", "C:\\Users").Output()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Output:\n%s\n", out)
	fmt.Println("Now Running SSH")
	start := time.Now()
	sshConfig := &ssh.ClientConfig{
		User:            "root",
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Auth: []ssh.AuthMethod{
			ssh.Password("root"),
		},
	}

	connection, err := ssh.Dial("tcp", "192.168.56.102:22", sshConfig)
	if err != nil {
		panic("Failed to dial: " + err.Error())
	}

	session, err := connection.NewSession()
	if err != nil {
		panic("Failed to create session: " + err.Error())
	}
	defer session.Close()

	var b bytes.Buffer
	session.Stdout = &b
	if err := session.Run("df -h"); err != nil {
		panic("Failed to run: " + err.Error())
	}
	fmt.Println(time.Since(start).String())
	fmt.Println(b.String())
}
