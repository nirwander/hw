package main

import (
	"bytes"
	"encoding/gob"
	"flag"
	"fmt"
	"net"
	"strconv"
)

type CsvLine struct {
	No            string
	Date          string
	Host          string
	GGhome        string
	GGversion     string
	Name          string
	Type          string
	Status        string
	Mode          string
	SourseDBAlias string
	MiningDBAlias string
	DestDBAlias   string
	SrcTrail      string
	DestHost      string
	DestTrail     string
}

type CsvData []CsvLine

// cmd flags
var tcpPort int

func init() {
	const (
		defaultTcpPort = 8000
		tcpPortUsage   = "set port= to start listening on the specified port"
	)
	flag.IntVar(&tcpPort, "port", defaultTcpPort, tcpPortUsage)
}

func main() {

	flag.Parse()
	// for purpose of verbosity, I will be removing error handling from this
	// sample code

	server, _ := net.Listen("tcp", ":"+strconv.Itoa(tcpPort))
	conn, _ := server.Accept()

	// create a temp buffer
	tmp := make([]byte, 1024)

	// loop through the connection to read incoming connections. If you're doing by
	// directional, you might want to make this into a seperate go routine
	for {

		_, _ = conn.Read(tmp)

		// convert bytes into Buffer (which implements io.Reader/io.Writer)
		tmpbuff := bytes.NewBuffer(tmp)

		tmpstruct := new(CsvData)

		// creates a decoder object
		gobobj := gob.NewDecoder(tmpbuff)

		// decodes buffer and unmarshals it into a Message struct
		gobobj.Decode(tmpstruct)

		// lets print out!
		fmt.Println(tmpstruct) // reflects.TypeOf(tmpstruct) == Message{}

	}

}
