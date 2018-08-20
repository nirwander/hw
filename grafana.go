package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net"
	"strconv"
	"time"
)

var data []string

func main() {
	start := time.Now()
	//loc, _ := time.LoadLocation("Europe/Moscow")

	conn, err := net.Dial("tcp", "unix-dashboard.megafon.ru:2103")
	if err != nil {
		fmt.Println("Error establishing connection to grafana", err)
		return
	}
	defer conn.Close()
	//'Oracle.DWH.''||p_db||''.cellmetric_sum.''||p_disk_type||''.''||regexp_replace(substr(regexp_replace(p_metric_name,''[(,),%,/,:]'',''''),1,49),''[*, ]'',''_'')||''.''||p_suffix
	/* mn := "Oracle.DWH.msk_uat.cellmetric_sum.cpu.avg"
	mv := 58
	mt := time.Now().In(loc).Unix()

	tcpString := mn + " " + strconv.Itoa(mv) + " " + strconv.FormatInt(mt, 10) + "\r\n"
	fmt.Print(tcpString)
	fmt.Fprintf(conn, tcpString) */

	//fmt.Println(len(data))
	getData(&data)
	//fmt.Println(len(data))

	//fmt.Printf("%q", data)

	for _, tcpString1 := range data {
		fmt.Printf("%s", tcpString1)
		fmt.Fprintf(conn, tcpString1)
	}

	fmt.Println(time.Since(start).String())
}

func getData(s *[]string) {
	fileBytes, _ := ioutil.ReadFile(`C:\Users\ivan.zotov\go\src\github.com\nirwander\hw\cellm.txt`)

	lines := bytes.Split(fileBytes, []byte("\n"))
	//re := regexp.MustCompile(`(?i)([[:alnum:]-]+):[[:space:]]+([[:alnum:]_]+)[[:space:]]{2,}([^[:space:]]+)[[:space:]]{2,}([^[:space:]]+)[[:space:]]{2,}(.+)`)
	var hostname string
	var metric string
	var metricTime time.Time
	var metricValue float64
	var err error

	for _, line := range lines {
		//msk-dev-celadm01: CL_CPUT        2018-08-20T12:55:07+03:00       msk_dev_celadm01        8.3 %
		//matches := re.FindSubmatch(bytes.TrimSpace(line))
		//fmt.Printf("%s\n", line)
		//if len(matches) > 0 {
		//	fmt.Printf("%q\n", matches)
		//}

		//fmt.Println("")

		fields := bytes.Fields(line)
		if len(fields) > 5 {
			//fmt.Printf("%q\n", fields)
			hostname = string(fields[3])
			metric = string(fields[1])
			metricTime, err = time.Parse(time.RFC3339, string(fields[2]))
			if err != nil {
				fmt.Println("Error converting time", err)
				return
			}
			metricValue, err = strconv.ParseFloat(string(fields[4]), 64)
			if err != nil {
				fmt.Println("Error converting metric value", err)
				return
			}

			/* fmt.Println(hostname)
			fmt.Println(metric)
			fmt.Println(metricTime)
			fmt.Println(metricTime.Unix())
			fmt.Println(metricValue) */

			//mn := "Oracle.DWH.msk_uat.cellmetric_sum.cpu.avg"
			str := "Oracle.DWH.msk_uat.cellmetric." + hostname + "." + metric + " " + strconv.FormatFloat(metricValue, 'f', -1, 64) + " " + strconv.FormatInt(metricTime.Unix(), 10) + "\r\n"

			//fmt.Print(str)
			*s = append(*s, str)
			//fmt.Println(len(*s))
		}
	}
	//fmt.Println("Done")
}
