package main

import (
	"database/sql"
	"fmt"
	"os"
	"time"

	_ "github.com/mattn/go-oci8"
)

func main() {

	os.Setenv("NLS_LANG", "")

	db, err := sql.Open("oci8" /*os.Args[1]*/, `sys/sys@orcl?as=sysdba&prefetch_rows=1000`)
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

	objs := make([]int, 0, 100000)

	start = time.Now()
	rows, err := db.Query("select a_object_id from hr.large_table where rownum<=2000000")
	defer rows.Close()
	for rows.Next() {
		var obj int
		err := rows.Scan(&obj)
		if err != nil {
			fmt.Println(err)
			return
		}
		objs = append(objs, obj)
	}

	fmt.Println(len(objs))
	fmt.Printf("Records fetched in %s\n", time.Since(start))

}
