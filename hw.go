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

func bzipCompress(d []byte) ([]byte, error) {
	var out bytes.Buffer
	// -c : compress
	// -9 : select the highest level of compresion
	cmd := exec.Command("bzip2", "-c", "-9")
	cmd.Stdin = bytes.NewBuffer(d)
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		return nil, err
	}
	return out.Bytes(), nil
}
