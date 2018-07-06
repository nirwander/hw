package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"strings"
	"time"
)

const curl = `D:\_Soft\curl-7.60.0-win64-mingw\bin\curl.exe`

func main() {
	//Получаем список файлов
	fileBytes, err := ioutil.ReadFile(`E:\list.txt`)
	if err != nil {
		panic(err)
	}

	lines := bytes.Split(fileBytes, []byte("\n"))

	result := make(chan string, 10)
	limit := make(chan int, 3)

	for j, line := range lines {
		fmt.Println(j)
		fmt.Println(time.Now().Format(time.RFC3339))
		limit <- 1
		go runCurl(line, result, limit)
	}

	fmt.Println("Waiting routines")
	for i := 0; i < len(lines); i++ {
		fmt.Println(<-result)
	}
	fmt.Println("Done")
}

//curl --upload-file "Z:\Exadata\SR 3-17840843051  RAC database crash 20180705\ExaWatcher_msk-kb-dbadm04.megafon.ru_2018-07-05_09_00_00_5h00m00s.tar.bz2" --user ivan.zotov@megafon.ru:.Member3 --proxy http://dv-proxy.megafon.ru:3128 https://transport.oracle.com/upload/issue/3-17840843051/

func runCurl(file string, res chan string, limit chan int) {
	file = strings.TrimSpace(file)
	//fmt.Println(args)
	//pos := strings.Index(file, "SR 3-")
	//sr := file[pos+3 : pos+16]
	//fmt.Println(pos)
	//fmt.Println(sr)

	var cmdArgs []string
	cmdArgs = append(cmdArgs, `--upload-file`)
	cmdArgs = append(cmdArgs, file)
	cmdArgs = append(cmdArgs, `--user`)
	cmdArgs = append(cmdArgs, `ivan.zotov@megafon.ru:.Member3`)
	cmdArgs = append(cmdArgs, `--proxy`)
	cmdArgs = append(cmdArgs, `http://dv-proxy.megafon.ru:3128`)
	//cmdArgs = append(cmdArgs, `https://transport.oracle.com/upload/issue/`+sr+`/`)

	//fmt.Println(cmdArgs)
	start := time.Now()
	// Делаем 5 попыток загрузки
	/*for i := 0; i < 5; i++ {
		cmd := exec.Command(curl, cmdArgs...)
		out, err := cmd.CombinedOutput()
		if err == nil {
			break
		}
		fmt.Printf("Try #%d ", i)
		fmt.Println(err)
		fmt.Println(string(out))
	}*/
	time.Sleep(5 * time.Second)
	fmt.Println("Done goroutine")
	//fmt.Println(time.Now())
	res <- fmt.Sprintf("%s : In %s uploaded file %s\n", time.Now().Format(time.RFC3339), time.Since(start), file)
	<-limit
	//fmt.Srintf("%s : In %s uploaded file %s\n", time.Now(), time.Since(start), file)

}
