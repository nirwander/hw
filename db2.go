package main

import (
	"database/sql"
	"fmt"
	"os"
	"time"

	"gopkg.in/goracle.v2"
)

func main() {

	//fmt.Println(2 << 19)
	//panic("Stop")
	os.Setenv("NLS_LANG", "")

	db, err := sql.Open("goracle" /*os.Args[1]*/, `sys/sys@orcl as sysdba`)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer db.Close()

	start := time.Now()
	var cnt int64
	err = db.QueryRow("select /*+ parallel(4) */ count(*) from hr.large_table").Scan(&cnt)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Printf("Successful 'as sysdba' connection. Large table records count: %v\n", cnt)
	fmt.Printf("Count fetched in %s\n\n", time.Since(start))

	type rec struct {
		ObjID   int
		ObjName string
	}

	//objs := make([]int, 0, 2<<19)
	objs := make([]rec, 0, 2<<19)

	start = time.Now()
	rows, err := db.Query("select a_object_id, a_object_name from hr.large_table where rownum<=2000000", goracle.FetchRowCount(10000))
	defer rows.Close()
	for rows.Next() {
		var obj int
		var n string
		err := rows.Scan(&obj, &n)
		if err != nil {
			fmt.Println(err)
			return
		}
		objs = append(objs, rec{obj, n})
	}

	fmt.Println(len(objs))
	fmt.Printf("Records fetched in %s\n", time.Since(start))
	fmt.Println(objs[0])

}
