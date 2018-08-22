package main

import (
	"regexp"
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

var data []string

const dcli = `/usr/local/bin/dcli`
const configFile = `/root/go/src/grafana.json`

// Структура для хранения конфигурации, получаемой из json файла
type config struct {
	HostGroup    string `json:"hostGroup"`
	MetricDb     string `json:"metricDb"`
	MetricGroup  string `json:"metricGroup"`
	MetricCmd    string `json:"metricCmd"`
	MetricFormat sring  `json:"metricFormat"`
}

// Config - configuration parameters
var Config []config

// cmd flags
var fdebug bool

func init() {
	const (
		defaultDebug = false
		debugUsage   = "set debug=true to get output metrics in StdOut"
	)
	flag.BoolVar(&fdebug, "debug", defaultDebug, debugUsage)
}

func main() {

	flag.Parse()
	getConfig()

	conn, err := net.Dial("tcp", "unix-dashboard.megafon.ru:2103")
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
		//cmdArgs = append(cmdArgs, `/root/cell_group`)
		cmdArgs = append(cmdArgs, arr.HostGroup)
		cmdArgs = append(cmdArgs, `-l`)
		cmdArgs = append(cmdArgs, `root`)
		cmdArgs = append(cmdArgs, `--maxlines=1000000`)
		//cmdArgs = append(cmdArgs, `cellcli -e list metriccurrent where name = 'CL_CPUT' attributes name,collectionTime,metricObjectName,metricValue`)
		cmdArgs = append(cmdArgs, arr.MetricCmd)

		//'Oracle.DWH.''||p_db||''.cellmetric_sum.''||p_disk_type||''.''||regexp_replace(substr(regexp_replace(p_metric_name,''[(,),%,/,:]'',''''),1,49),''[*, ]'',''_'')||''.''||p_suffix
		start := time.Now()
		log.Println("Getting data...")
		data = make([]string, 10, 100)
		getData(&data, cmdArgs, arr)
		log.Println("Got in ", time.Since(start).String())

		//fmt.Printf("%q", data)
		start = time.Now()
		log.Println("Sending data...")

		for _, tcpString := range data {
			if fdebug {
				fmt.Printf("%s", tcpString)
			}
			fmt.Fprintf(conn, tcpString)
		}

		log.Println("Sent in ", time.Since(start).String())

	}
}

func getData(s *[]string, args []string, metricCfg config) {
	//fileBytes, _ := ioutil.ReadFile(`C:\Users\ivan.zotov\go\src\github.com\nirwander\hw\cellm.txt`)
	fileBytes := execCmd(dcli, args)

	lines := bytes.Split(fileBytes.Bytes(), []byte("\n"))

	var hostname string
	var metric string
	var metricTime time.Time
	var metricObj string
	var metricValue float64
	var err error
	var re regexp.Regexp
	if metricCfg.MetricFormat != nil {
		re = regexp.MustCompile(metricCfg.MetricFormat)
	}

	for _, line := range lines {

		if metricCfg.MetricFormat == nil {
			fields := bytes.Fields(line)
			if len(fields) > 5 {
				//fmt.Printf("%q\n", fields)
				hostname = strings.TrimSuffix(string(fields[0]), ":")
				metricObj = string(fields[3])
				metric = string(fields[1])
				metricTime, err = time.Parse(time.RFC3339, string(fields[2]))
				if err != nil {
					fmt.Println("Error converting time", err)
					return
				}
				metricValue, err = strconv.ParseFloat(strings.Replace(string(fields[4]), ",", "", -1), 64)
				if err != nil {
					fmt.Println("Error converting metric value", err)
					return
				}

				//mn := "Oracle.DWH.msk_uat.cellmetric_sum.cpu.avg"
				str := "Oracle.DWH." + metricCfg.MetricDb + "." + metricCfg.MetricGroup + "." + hostname + "." + metric + "." + metricObj + " " + strconv.FormatFloat(metricValue, 'f', -1, 64) + " " + strconv.FormatInt(metricTime.Unix(), 10) + "\r\n"

				*s = append(*s, str)
			}
		}
		else {

		}
	}
}

func execCmd(bin string, args []string) bytes.Buffer {
	var out bytes.Buffer
	var serr bytes.Buffer

	cmd := exec.Command(bin, args...)
	cmd.Stdout = &out
	cmd.Stderr = &serr

	err := cmd.Run()
	if err != nil {
		fmt.Printf("%s\n", serr)
		log.Fatal(err)
	}

	return out
}

func getConfig() {
	fileBytes, err := ioutil.ReadFile(configFile)
	if err != nil {
		log.Fatal("Error reading config file ", err)
	}

	err = json.Unmarshal(fileBytes, &Config)
	if err != nil {
		log.Fatal("Error parsing config ", err)
	}

	//fmt.Println(Config)
}
