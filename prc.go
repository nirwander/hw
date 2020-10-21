package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	ps "github.com/mitchellh/go-ps"
)

func main() {

	fmt.Println(string("..."))
	dir, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(dir)

	ex, err := os.Executable()
	if err != nil {
		panic(err)
	}
	exPath := filepath.Dir(ex)
	fmt.Println(exPath)

	prc, _ := ps.Processes()

	// fmt.Printf("%s", prc)

	for x := range prc {
		var process ps.Process
		process = prc[x]

		if strings.Contains(process.Executable(), "smon") {
			log.Printf("%d\t%s\n", process.Pid(), process.Executable())
			// do os.* stuff on the pid
			line := "/proc/" + strconv.Itoa(process.Pid()) + "/environ"
			fmt.Println(line)
			if _, err := os.Stat(line); err != nil {
				if os.IsNotExist(err) {
					fmt.Printf("File %s does not exists\n", line)
				} else {
					fmt.Printf("\n%s\n", err.Error())

				}
			} else {
				// read environment variables
				fileBytes, _ := ioutil.ReadFile(line)
				lines := bytes.Split(fileBytes, []byte("\x00"))
				for _, env := range lines {
					if bytes.Contains(env, []byte("ORA")) || bytes.Contains(env, []byte("LD")) {
						fmt.Printf("%s\n", env)
					}
				}
			}
		}
	}
}
