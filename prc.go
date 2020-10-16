package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
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

		if strings.Contains(process.Executable(), "cmd.exe") {
			log.Printf("%d\t%s\n", process.Pid(), process.Executable())
		}

		// do os.* stuff on the pid
		os.Get
	}
}
