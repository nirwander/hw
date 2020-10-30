package main

import (
	"bytes"
	"encoding/csv"
	"encoding/gob"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	ps "github.com/mitchellh/go-ps"
)

// CsvLine - структура для хранения строки с информацией о процессах GoldenGate
// #;Date;Host;GG_home;GG_version;Name;Type;Status;Mode;Sourse_DB_Alias;Mining_DB_Alias;Dest_DB_Alias;Src_trail;Dest_host;Dest_trail
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
	var llp string
	var pwd string

	for x := range prc {
		var process ps.Process
		process = prc[x]

		if process.Executable() == "mgr" {
			// TODO: усилить проверку через наличие специфических параметров для GG в cmdline
			log.Printf("%d\t%s\n", process.Pid(), process.Executable())
			// do os.* stuff on the pid
			line := "/proc/" + strconv.Itoa(process.Pid()) + "/environ"
			// fmt.Println(line)
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
					if bytes.HasPrefix(env, []byte("ORA")) || bytes.HasPrefix(env, []byte("LD")) || bytes.HasPrefix(env, []byte("PWD")) {
						fmt.Printf("%s\n", env)
					}

					if bytes.HasPrefix(env, []byte("PWD")) {
						pwd = string(bytes.TrimLeft(env, "PWD="))
					}
					if bytes.HasPrefix(env, []byte("LD_LIBRARY_PATH")) {
						llp = string(env)
					}
				}

				cmd := exec.Command("/bin/bash", "/home/oracle/go/src/github.com/nirwander/hw/ggproc.sh", pwd, "/home/oracle/go/src/github.com/nirwander/hw")
				cmd.Env = os.Environ()
				cmd.Env = append(cmd.Env, llp)
				fmt.Printf("\nRunning ggproc.sh for gghome %s\n", pwd)
				// fmt.Printf("\nEnv: %s\n", cmd.Env)
				if err := cmd.Run(); err != nil {
					log.Fatal(err)
				}

			}
		}
	}

	fmt.Printf("\nSearching csv files\n")

	files := readCurrentDir()
	var CSVfiles []string

	for _, name := range files {
		if strings.HasSuffix(name, ".csv") {
			CSVfiles = append(CSVfiles, name)
		}
	}

	fmt.Printf("Found csv files: %s\n", CSVfiles)

	fmt.Printf("\nProcessing files\n")

	var csvData = make(CsvData, 0, 20)

	for _, name := range CSVfiles {

		fmt.Printf("Parsing file: %s\n", name)
		lines, err := readCsv(name)
		if err != nil {
			log.Fatalf("Error reading CSV file %s: %s", name, err)
		}

		// Loop through lines & turn into object
		for _, line := range lines {
			data := CsvLine{
				No:            line[0],
				Date:          line[1],
				Host:          line[2],
				GGhome:        line[3],
				GGversion:     line[4],
				Name:          line[5],
				Type:          line[6],
				Status:        line[7],
				Mode:          line[8],
				SourseDBAlias: line[9],
				MiningDBAlias: line[10],
				DestDBAlias:   line[11],
				SrcTrail:      line[12],
				DestHost:      line[13],
				DestTrail:     line[14],
			}
			csvData = append(csvData, data)

		}
		fmt.Printf("Got %s\n\n", csvData)

		fmt.Printf("\nSending data...\n")
		sendData(csvData)
		if err != nil {
			log.Fatalf("Error sending data. %s", err)
		}

		fmt.Printf("\nClearing file %s\n", name)
		os.Remove(name)
	}
	fmt.Printf("\nDone\n")

}

func readCsv(filename string) ([][]string, error) {

	// Open CSV file
	f, err := os.Open(filename)
	if err != nil {
		return [][]string{}, err
	}
	defer f.Close()

	// Read File into a Variable
	reader := csv.NewReader(f)
	reader.Comma = ';'
	reader.FieldsPerRecord = 0

	lines, err := reader.ReadAll()

	if err != nil {
		return [][]string{}, err
	}

	return lines, nil
}

func readCurrentDir() []string {
	file, err := os.Open(".")
	if err != nil {
		log.Fatalf("Failed opening current directory: %s\n", err)
	}
	defer file.Close()

	list, _ := file.Readdirnames(0) // 0 to read all files and folders

	return list
}

func sendData(msg CsvData) error {

	binBuf := new(bytes.Buffer)

	// error handling still truncated
	conn, err := net.Dial("tcp", "0.0.0.0:8000")
	if err != nil {
		log.Fatalln(err)
	}
	defer conn.Close()

	// create a encoder object
	gobobj := gob.NewEncoder(binBuf)

	// encode buffer and marshal it into a gob object
	gobobj.Encode(msg)

	_, err = conn.Write(binBuf.Bytes())

	return err
}
