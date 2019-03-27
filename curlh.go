package main

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"sync"
	"syscall"
	"time"

	"golang.org/x/crypto/ssh/terminal"
)

//const curl = `D:\_Soft\curl-7.60.0-win64-mingw\bin\curl.exe`
const oracletransport = `https://transport.oracle.com/upload/issue/`

const curl = `curl.exe`

// Группа синхронизации - для ожидания завершения всех загрузок
var wg sync.WaitGroup

// cmd flags
var fMOSAccount string
var fProxy string
var fProxyAccount string
var fFilesList string

// УЗ для доступа к MOS и корпоративному proxy, используются в качетве аргументов cURL
var mosUser, proxyUser string

func init() {
	const (
		defaultMos      = "ivan.zotov@megafon.ru"
		MosUsage        = "Set MOS account in user[:pass] format. When no password is set it shall be asked separately"
		defaultProxyAcc = "ivan.zotov"
		ProxyAccUsage   = "Set Proxy account in user[:pass] format. When no password is set it shall be asked separately"
		defaultFilesL   = "list.txt"
		FilesListUsage  = "Set filename with the list of files to upload"
		defaultProxy    = "http://msk-proxy.megafon.ru:3128"
		ProxyUsage      = "Set http proxy for internet access"
	)
	flag.StringVar(&fMOSAccount, "mos_acc", defaultMos, MosUsage)
	flag.StringVar(&fProxy, "proxy", defaultProxy, ProxyUsage)
	flag.StringVar(&fProxyAccount, "proxy_acc", defaultProxyAcc, ProxyAccUsage)
	flag.StringVar(&fFilesList, "files_list", defaultFilesL, FilesListUsage)
}

func main() {
	//Разворачиваем аргументы
	flag.Parse()
	mosAccFlag := strings.Split(fMOSAccount, ":")
	if len(mosAccFlag[1:]) == 0 {
		//пароль не задан в cmd
		MosUserPwd := getPwd("Enter MOS Account password:")
		mosUser = mosAccFlag[0] + ":" + MosUserPwd
	} else {
		mosUser = mosAccFlag[0] + ":" + strings.Join(mosAccFlag[1:], ":")
	}

	proxyAccFlag := strings.Split(fProxyAccount, ":")
	if len(proxyAccFlag[1:]) == 0 {
		//пароль не задан в cmd
		ProxyUserPwd := getPwd("Enter Proxy Account password:")
		proxyUser = proxyAccFlag[0] + ":" + ProxyUserPwd
	} else {
		proxyUser = proxyAccFlag[0] + ":" + strings.Join(proxyAccFlag[1:], ":")
	}

	//Получаем список файлов
	fileBytes, err := ioutil.ReadFile(fFilesList)
	if err != nil {
		panic(err)
	}

	lines := bytes.Split(fileBytes, []byte("\n"))
	limit := make(chan int, 3)
	fmt.Println("Starting upload...")
	for _, line := range lines {
		fname := strings.TrimSpace(string(line))
		//fmt.Printf("Processing %s\n", fname)
		if _, err := os.Stat(fname); err != nil {
			if os.IsNotExist(err) {
				fmt.Printf("File \"%s\" does not exists\n", fname)
			} else {
				fmt.Println(err.Error())
			}
			continue
		}
		//fmt.Println(time.Now().Format(time.RFC3339))
		limit <- 1
		go runCurl(fname, limit)
	}

	//fmt.Println("Waiting routines")
	wg.Wait()
	fmt.Println("Done")
}

//curl --upload-file "Z:\Exadata\SR 3-17840843051  RAC database crash 20180705\ExaWatcher_msk-kb-dbadm04.megafon.ru_2018-07-05_09_00_00_5h00m00s.tar.bz2" --user ivan.zotov@megafon.ru:.Member3 --proxy http://dv-proxy.megafon.ru:3128 https://transport.oracle.com/upload/issue/3-17840843051/

func runCurl(file string, limit chan int) {
	wg.Add(1)
	defer wg.Done()
	//fmt.Println(args)
	// Service Request numeric format
	re := regexp.MustCompile(`SR ([0-9]-[0-9]{11})`)
	matches := re.FindStringSubmatch(file)
	//fmt.Println(matches)
	if len(matches[1]) == 0 {
		panic("Filename or path doesn't contain SR identifier")
	}
	sr := string(matches[1])
	//pos := strings.Index(file, "SR 3-")
	//sr := file[pos+3 : pos+16]
	//fmt.Println(pos)
	//fmt.Println(sr)
	//panic("Got to here")

	var cmdArgs []string
	cmdArgs = append(cmdArgs, `--upload-file`)
	cmdArgs = append(cmdArgs, file)
	cmdArgs = append(cmdArgs, `--user`)
	cmdArgs = append(cmdArgs, mosUser)
	cmdArgs = append(cmdArgs, `--proxy`)
	cmdArgs = append(cmdArgs, fProxy)
	cmdArgs = append(cmdArgs, `--proxy-user`)
	cmdArgs = append(cmdArgs, proxyUser)
	cmdArgs = append(cmdArgs, oracletransport+sr+`/`)

	//fmt.Println(cmdArgs)
	start := time.Now()
	// Делаем 3 попытки загрузки
	var uploaded bool
	//uploaded = true
	for i := 0; i < 3; i++ {
		cmd := exec.Command(curl, cmdArgs...)
		err := cmd.Run()
		if err == nil {
			uploaded = true
			break
		}
		fmt.Printf("Try No. %d\n", i)
		fmt.Println(err)
		//fmt.Println(cmdArgs)
	}
	//time.Sleep(5 * time.Second)
	//fmt.Println("Done goroutine")
	if uploaded {
		fmt.Printf("%s : In %s uploaded file %s\n", time.Now().Format(time.RFC3339), time.Since(start), file)
	} else {
		fmt.Printf("%s : failed to upload file %s\n", time.Now().Format(time.RFC3339), file)
	}
	<-limit
}

func getPwd(hello string) string {
	fmt.Println(hello)
	bytePassword, err := terminal.ReadPassword(int(syscall.Stdin))
	if err != nil {
		fmt.Println("Can't get password")
		panic(err)
	}
	return strings.TrimSpace(string(bytePassword))
}
