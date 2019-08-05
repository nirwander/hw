package main

// https://vlg-gitlab01.megafon.ru/dwh/cellmetrics.v2

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
)

var data []string

const dcli = `/usr/local/bin/dcli`
const configFile = `grafana.json`
const grafanaSrv = `unix-dashboard.megafon.ru:2103`

// Группа синхронизации - для ожидания получения всех данных
var wg sync.WaitGroup
var mu sync.Mutex

// Структура для хранения конфигурации, получаемой из json файла
type config struct {
	HostGroup    string `json:"hostGroup"`
	MetricDb     string `json:"metricDb"`
	MetricGroup  string `json:"metricGroup"`
	MetricCmd    string `json:"metricCmd"`
	MetricFormat string `json:"metricFormat"`
}

// Config - configuration parameters
var Config []config

// cmd flags
var fdebug bool

func init() {
	const (
		defaultDebug = false
		debugUsage   = "set debug=true to get output metrics in StdOut instead of sending to Grafana"
	)
	flag.BoolVar(&fdebug, "debug", defaultDebug, debugUsage)
}

func main() {

	flag.Parse()
	getConfig()

	conn, err := net.Dial("tcp", grafanaSrv)
	if err != nil {
		fmt.Println("Error establishing connection to grafana", err)
		return
	}
	defer conn.Close()

	for i, arr := range Config {
		//fmt.Printf("%q\n", arr)
		log.Println("Working on metric #", i+1)
		var cmdArgs []string
		cmdArgs = append(cmdArgs, `-g`)
		cmdArgs = append(cmdArgs, arr.HostGroup)
		cmdArgs = append(cmdArgs, `-l`)
		cmdArgs = append(cmdArgs, `root`)
		cmdArgs = append(cmdArgs, `--maxlines=1000000`)
		cmdArgs = append(cmdArgs, arr.MetricCmd)

		data = make([]string, 0, 100)
		if fdebug {
			fmt.Println(cmdArgs)
		}
		go getData(&data, cmdArgs, arr, i+1)

	}
	wg.Wait()
	start = time.Now()
	log.Println("Sending data...")

	for _, tcpString := range data {
		if fdebug {
			fmt.Printf("%s", tcpString)
		} else {
			fmt.Fprintf(conn, tcpString)
		}
	}

	log.Println("Sent in", time.Since(start).String(), len(data), "records")

}

func getData(s *[]string, args []string, metricCfg config, i int) {
	start := time.Now()
	wg.Add(1)
	defer wg.Done()
	log.Printf("#%d Getting data...\n", i)

	fileBytes := execCmd(dcli, args)
	if fdebug {
		log.Printf("#%d Command executed\n", i)
	}
	lines := bytes.Split(fileBytes.Bytes(), []byte("\n"))
	if fdebug {
		log.Printf("#%d Got lines\n", i)
	}
	var hostname string
	var metric string
	var metricTime time.Time
	var metricObj string
	var metricValue float64
	var err error

	//Для синхронизации показателей на графиках выгружаем все данные в одно время. Иначе получаем "лесенку"
	metricTime = time.Now().In(time.Local)

	for _, line := range lines {

		fields := bytes.Fields(line)
		switch metricCfg.MetricFormat {
		case "cellcli":
			if len(fields) > 4 {
				hostname = strings.TrimSuffix(string(fields[0]), ":")
				metricObj = string(fields[3])
				metric = string(fields[1])

				metricValue, err = strconv.ParseFloat(strings.Replace(string(fields[4]), ",", "", -1), 64)
				if err != nil {
					fmt.Printf("#%d Error converting metric value, %s", i, err)
					return
				}

				str := "Oracle.DWH." + metricCfg.MetricDb + "." + metricCfg.MetricGroup + "." + hostname + "." + metric + "." + metricObj + " " + strconv.FormatFloat(metricValue, 'f', -1, 64) + " " + strconv.FormatInt(metricTime.Unix(), 10) + "\r\n"

				mu.Lock()
				*s = append(*s, str)
				mu.Unlock()
			}
		case "ilom":
			if len(fields) > 3 {
				hostname = strings.TrimSuffix(string(fields[0]), ":")
				for i, field := range fields {
					if string(field) == "value" && string(fields[i+1]) == "=" {
						metricValue, err = strconv.ParseFloat(strings.Replace(string(fields[i+2]), ",", "", -1), 64)
						if err != nil {
							fmt.Printf("#%d Error converting metric value, %s", i, err)
							return
						}
						str := "Oracle.DWH." + metricCfg.MetricDb + "." + metricCfg.MetricGroup + "." + hostname + " " + strconv.FormatFloat(metricValue, 'f', -1, 64) + " " + strconv.FormatInt(metricTime.Unix(), 10) + "\r\n"
						mu.Lock()
						*s = append(*s, str)
						mu.Unlock()
						break
					}
				}
			}
		}
	}
	log.Printf("#%d Got in %s\n", i, time.Since(start).String())
}

func execCmd(bin string, args []string) bytes.Buffer {
	var out bytes.Buffer
	var serr bytes.Buffer

	cmd := exec.Command(bin, args...)
	cmd.Stdout = &out
	cmd.Stderr = &serr

	err := cmd.Run()
	// log.Printf("Command executed\n")
	// log.Printf("Got %d bytes\n", len(out.Bytes()))
	// log.Printf("Got %d error bytes\n", len(serr.Bytes()))
	if err != nil {
		// Некритичная ошибка
		if bytes.Contains(serr.Bytes(), []byte("Unable to connect")) {
			log.Printf("%s\n", serr)
		} else {
			log.Printf("%s\n", serr)
		}
	}
	// log.Printf("Returned\n")
	return out
}

func getConfig() {
	ex, err := os.Executable()
	if err != nil {
		log.Fatal(err)
	}
	exPath := filepath.Dir(ex)

	fileBytes, err := ioutil.ReadFile(exPath + "/" + configFile)
	if err != nil {
		log.Fatal("Error reading config file - expecting", exPath+"/"+configFile, err)
	}

	err = json.Unmarshal(fileBytes, &Config)
	if err != nil {
		log.Fatal("Error parsing config", err)
	}

}
