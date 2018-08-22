package main

import (
	"bytes"
	"fmt"
	"log"
	"os/exec"
)

const dcli = `/usr/local/bin/dcli`

func main() {

	var cmdArgs []string
	cmdArgs = append(cmdArgs, `-g`)
	cmdArgs = append(cmdArgs, `/root/cell_group`)
	cmdArgs = append(cmdArgs, `-l`)
	cmdArgs = append(cmdArgs, `root`)
	cmdArgs = append(cmdArgs, `--maxlines=1000000`)
	cmdArgs = append(cmdArgs, `ipmitool sunoem cli "show /SYS/T_AMB" | grep value`)

	var out bytes.Buffer
	var serr bytes.Buffer

	cmd := exec.Command(dcli, cmdArgs...)
	cmd.Stdout = &out
	cmd.Stderr = &serr

	err := cmd.Run()
	if err != nil {
		log.Fatal(serr, err)
	}

	fmt.Printf("%s\n", out)
	/* f, err := strconv.ParseFloat(strings.Replace("335,693", ",", "", -1), 64)
	if err != nil {
		fmt.Println("Error converting metric value", err)
		return
	}

	fmt.Println(f) */
}