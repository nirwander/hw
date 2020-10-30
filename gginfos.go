package main

import (
	"bytes"
	"database/sql"
	"encoding/gob"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"strconv"

	_ "gopkg.in/goracle.v2"
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
var fdebug bool

func init() {
	const (
		defaultTCPPort = 8000
		tcpPortUsage   = "set port= to start listening on the specified port"
		defaultDebug   = false
		debugUsage     = "set debug=true to get output data in StdOut (and additional info)"
	)
	flag.IntVar(&tcpPort, "port", defaultTCPPort, tcpPortUsage)
	flag.BoolVar(&fdebug, "debug", defaultDebug, debugUsage)
}

func main() {

	flag.Parse()
	// for purpose of verbosity, I will be removing error handling from this
	// sample code

	server, _ := net.Listen("tcp", ":"+strconv.Itoa(tcpPort))
	defer server.Close()

	fmt.Printf("Accept connection on port %s\n", tcpPort)

	for {
		conn, err := server.Accept()
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println("Calling handleConnection")
		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	defer conn.Close()

	tmp, err := ioutil.ReadAll(conn)
	if err != nil {
		log.Printf("Got error reading connection, %s", err)
	}

	// convert bytes into Buffer (which implements io.Reader/io.Writer)
	tmpbuff := bytes.NewBuffer(tmp)

	tmpstruct := new(CsvData)

	// creates a decoder object
	gobobj := gob.NewDecoder(tmpbuff)

	// decodes buffer and unmarshals it into a Message struct
	gobobj.Decode(tmpstruct)

	if fdebug {
		fmt.Printf("Got %s\n\n", tmpstruct)
	}

	saveToDB(tmpstruct)
}

func saveToDB(data *CsvData) {
	db, err := sql.Open("goracle" /*os.Args[1]*/, `inventory/Pi100_let338@msk_dbmon`)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer db.Close()

	tx, err := db.Begin() //db.Begin()
	if err != nil {
		log.Printf("Error starting DB transaction: %s\n", err)
	}

	// #;Date;Host;GG_home;GG_version;Name;Type;Status;Mode;Sourse_DB_Alias;Mining_DB_Alias;Dest_DB_Alias;Src_trail;Dest_host;Dest_trail
	stmt, err := tx.Prepare("insert into gginfo_raw_data values (:b1, :b2, :b3, :b4, :b5, :b6, :b7, :b8, :b9, :b10, :b11, :b12, :b13, :b14, :b15)")
	if err != nil {
		log.Printf("Error preparing statement, %s\n", err)
	}

	var tmpRowsCnt int64
	for _, v := range *data {
		// Пропускаем строку с заголовком
		if v.No == "#" {
			continue
		}
		res, err := stmt.Exec(v.No, v.Date, v.Host, v.GGhome, v.GGversion, v.Name, v.Type, v.Status, v.Mode, v.SourseDBAlias, v.MiningDBAlias, v.DestDBAlias, v.SrcTrail, v.DestHost, v.DestTrail)
		if err != nil {
			log.Println("Error inserting row in gginfo_raw_data: " + err.Error())
		}
		rowsCnt, err := res.RowsAffected()
		if err != nil {
			log.Println("Error getting affected rows: " + err.Error())
		}
		tmpRowsCnt += rowsCnt
	}
	if fdebug {
		log.Println("Inserted " + strconv.FormatInt(tmpRowsCnt, 10) + " rows into tmp_replicated_tables")
	}
	err = stmt.Close()
	if err != nil {
		log.Println("Error closing statement: " + err.Error())
	}

	err = tx.Commit()
	if err != nil {
		log.Println("Error commiting transaction: " + err.Error())
	}
}
