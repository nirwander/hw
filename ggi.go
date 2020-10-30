package main

import (
	"bytes"
	"encoding/csv"
	"encoding/gob"
	"flag"
	"io/ioutil"
	"log"
	"net"
	"os"
	"os/exec"
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

// CsvData - структура для хранения csv файла с информацией о процессах GoldenGate
type CsvData []CsvLine

// cmd flags
var fdebug bool
var fServer string

func init() {
	const (
		defaultDebug  = false
		debugUsage    = "set debug=true to get output data in StdOut (and additional info)"
		defaultServer = "10.99.76.130:11000"
		ServerUsage   = "set server parameter in format ip:port to send data to"
	)
	flag.StringVar(&fServer, "server", defaultServer, ServerUsage)
	flag.BoolVar(&fdebug, "debug", defaultDebug, debugUsage)
}

func main() {

	log.Println("Program started")
	flag.Parse()
	dir, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Current working dir %s\n", dir)

	// ex, err := os.Executable()
	// if err != nil {
	// 	panic(err)
	// }
	// exPath := filepath.Dir(ex)
	// log.Println(exPath)

	prc, _ := ps.Processes()

	// log.Printf("%s", prc)
	var llp string
	var pwd string
	var mgr bool

	for x := range prc {
		var process ps.Process
		process = prc[x]

		if process.Executable() == "mgr" {
			if fdebug {
				log.Printf("%d\t%s\n", process.Pid(), process.Executable())
			}
			// do os.* stuff on the pid
			// /proc/53209/cmdline
			// ./mgrPARAMFILE/data/ggate/ggs/dirprm/mgr.prmREPORTFILE/data/ggate/ggs/dirrpt/MGR.rptPROCESSIDMGR

			line := "/proc/" + strconv.Itoa(process.Pid()) + "/cmdline"
			if _, err := os.Stat(line); err != nil {
				if os.IsNotExist(err) {
					log.Printf("File %s does not exists\n", line)
				} else {
					log.Printf("\n%s\n", err.Error())

				}
			} else {
				// read cmdline params
				fileBytes, _ := ioutil.ReadFile(line)
				lines := bytes.Split(fileBytes, []byte("\x00"))

				// fmt.Printf("Cmdline: %s\n\n", lines)

				for i, cm := range lines {
					// fmt.Printf("cm: %s\n", cm)

					if bytes.Equal(cm, []byte("PARAMFILE")) {
						pwd = string(bytes.TrimSuffix(lines[i+1], []byte("/dirprm/mgr.prm")))
					}
					if bytes.Equal(cm, []byte("PROCESSID")) {
						if bytes.Equal(lines[i+1], []byte("MGR")) {
							mgr = true
						}
					}
				}

				log.Printf("pwd: %s\n", pwd)

				// panic(0)

				if mgr {

					line = "/proc/" + strconv.Itoa(process.Pid()) + "/environ"
					// log.Println(line)
					if _, err := os.Stat(line); err != nil {
						if os.IsNotExist(err) {
							log.Printf("File %s does not exists\n", line)
						} else {
							log.Printf("\n%s\n", err.Error())

						}
					} else {
						// read environment variables

						fileBytes, _ := ioutil.ReadFile(line)
						lines := bytes.Split(fileBytes, []byte("\x00"))
						for _, env := range lines {
							if bytes.HasPrefix(env, []byte("ORA")) || bytes.HasPrefix(env, []byte("LD")) || bytes.HasPrefix(env, []byte("PWD")) {
								if fdebug {
									log.Printf("%s\n", env)
								}
							}

							// if bytes.HasPrefix(env, []byte("PWD")) {
							// 	pwd = string(bytes.TrimLeft(env, "PWD="))
							// }
							if bytes.HasPrefix(env, []byte("LD_LIBRARY_PATH")) {
								llp = string(env)
							}
						}

						cmd := exec.Command("/bin/bash", "ggproc.sh", pwd, "./")
						cmd.Env = os.Environ()
						cmd.Env = append(cmd.Env, llp)
						if fdebug {
							log.Printf("Running ggproc.sh for gghome %s\n", pwd)
						}
						// log.Printf("\nEnv: %s\n", cmd.Env)
						if err := cmd.Run(); err != nil {
							log.Fatal(err)
						}

					}
				}
			}
		}
	}

	log.Printf("Searching csv files\n")

	files := readCurrentDir()
	var CSVfiles []string

	for _, name := range files {
		if strings.HasSuffix(name, ".csv") {
			CSVfiles = append(CSVfiles, name)
		}
	}

	log.Printf("Found csv files: %s\n", CSVfiles)

	log.Printf("Processing files\n")

	var csvData = make(CsvData, 0, 20)

	for _, name := range CSVfiles {

		log.Printf("Parsing file: %s\n", name)
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
		if fdebug {
			log.Printf("Got %s\n\n", csvData)
		}

		log.Printf("Sending data...\n")
		sendData(csvData)
		if err != nil {
			log.Fatalf("Error sending data. %s", err)
		}

		log.Printf("Clearing file %s\n", name)
		os.Remove(name)
	}
	log.Printf("Done\n")

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
	conn, err := net.Dial("tcp", fServer)
	if err != nil {
		log.Fatalln(err)
	}
	defer conn.Close()

	// create a encoder object
	gobobj := gob.NewEncoder(binBuf)

	// encode buffer and marshal it into a gob object
	gobobj.Encode(msg)

	i, err := conn.Write(binBuf.Bytes())

	if fdebug {
		log.Printf("Sent %d bytes to server", i)
	}

	return err
}
