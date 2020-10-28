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

	data = make([]string, 0, 10000)
	// Собираем не более 5 метрик одновременно
	limit := make(chan int, 5)
	for i, arr := range Config {
		log.Println("Working on metric #", i+1)
		var cmdArgs []string
		cmdArgs = append(cmdArgs, `-g`, arr.HostGroup, `-l`, `root`, `--maxlines=1000000`, arr.MetricCmd)

		if fdebug {
			fmt.Println(cmdArgs)
		}
		limit <- 1
		wg.Add(1)
		go getData(&data, cmdArgs, arr, i+1, limit)

	}
	wg.Wait()
	start := time.Now()
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

func getData(s *[]string, args []string, metricCfg config, i int, limit chan int) {
	start := time.Now()
	defer wg.Done()
	log.Printf("#%d Getting data...\n", i)

	fileBytes := execCmd(dcli, args, i)
	if fdebug {
		log.Printf("#%d Command executed\n", i)
	}
	lines := bytes.Split(fileBytes.Bytes(), []byte("\n"))

	var hostname string
	var metric string
	var metricTime time.Time
	var metricObj string
	var metricValue float64
	var err error
	var res []string
	res = make([]string, 0, 100)

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
					fmt.Printf("\t#%d Error converting metric value, %s\n", i, err)
					fmt.Printf("\t#%d Dumping value: %s\n", i, fields[4])
					return
				}

				str := "Oracle.DWH." + metricCfg.MetricDb + "." + metricCfg.MetricGroup + "." + hostname + "." + metric + "." + metricObj + " " + strconv.FormatFloat(metricValue, 'f', -1, 64) + " " + strconv.FormatInt(metricTime.Unix(), 10) + "\r\n"

				res = append(res, str)
			}
		case "ilom":
			if len(fields) > 3 {
				hostname = strings.TrimSuffix(string(fields[0]), ":")
				for i, field := range fields {
					if string(field) == "value" && string(fields[i+1]) == "=" {
						metricValue, err = strconv.ParseFloat(strings.Replace(string(fields[i+2]), ",", "", -1), 64)
						if err != nil {
							fmt.Printf("\t#%d Error converting metric value, %s\n", i, err)
							fmt.Printf("\t#%d Dumping value: %s\n", i, fields[i+2])
							return
						}
						str := "Oracle.DWH." + metricCfg.MetricDb + "." + metricCfg.MetricGroup + "." + hostname + " " + strconv.FormatFloat(metricValue, 'f', -1, 64) + " " + strconv.FormatInt(metricTime.Unix(), 10) + "\r\n"
						res = append(res, str)
						break
					}
				}
			}
		}
	}
	mu.Lock()
	*s = append(*s, res...)
	mu.Unlock()
	log.Printf("#%d Got %d lines in %s, total %d lines of data\n", i, len(res), time.Since(start).String(), len(*s))
	<-limit
}

func execCmd(bin string, args []string, i int) bytes.Buffer {
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
		// if bytes.Contains(serr.Bytes(), []byte("Unable to connect")) {
		// 	log.Printf("%s\n", serr)
		// } else {
		log.Printf("#%d Error executing command %s; %s; %s \n", i, args, serr, err)
		// }
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
