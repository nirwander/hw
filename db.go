package main

import (
	"database/sql"
	"fmt"
	"os"

	_ "github.com/mattn/go-oci8"
)

func main() {

	os.Setenv("NLS_LANG", "")

	db, err := sql.Open("oci8" /*os.Args[1]*/, `sys/sys@orcl?as=sysdba`)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer db.Close()

	var cnt int64
	err = db.QueryRow("select count(*) from hr.large_table").Scan(&cnt)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Printf("Successful 'as sysdba' connection. Large table records count: %v\n", cnt)

}
