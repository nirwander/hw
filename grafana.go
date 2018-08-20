package main

import (
	"fmt"
	"net"
	"strconv"
	"time"
)

func main() {
	start := time.Now()
	loc, _ := time.LoadLocation("Europe/Moscow")

	conn, err := net.Dial("tcp", "unix-dashboard.megafon.ru:2103")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer conn.Close()
	//'Oracle.DWH.''||p_db||''.cellmetric_sum.''||p_disk_type||''.''||regexp_replace(substr(regexp_replace(p_metric_name,''[(,),%,/,:]'',''''),1,49),''[*, ]'',''_'')||''.''||p_suffix
	mn := "Oracle.DWH.msk_uat.cellmetric_sum.cpu.avg"
	mv := 58
	mt := time.Now().In(loc).Unix()

	tcpString := mn + " " + strconv.Itoa(mv) + " " + strconv.FormatInt(mt, 10) + "\r\n"
	fmt.Print(tcpString)
	fmt.Fprintf(conn, tcpString)

	fmt.Println(time.Since(start).String())
}
