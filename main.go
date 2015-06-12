package main

import (
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"syscall"
	"time"
)

var (
	dir  string
	port string
)

var (
	portExp = regexp.MustCompile(`([1-9]\d{0,3})|([1-5]\d{4})|(6[0-4]\d{3})|(65[0-4]\d{2})|(655[0-2]\d)|(6553[0-5])`)
)

const (
	FileNotFound = `<h1 style="text-align:center;font-family:cursive;">404</h1><p style="text-align:center;font-family: cursive;">:(</><p style="text-align:center;font-family: monospace;">无法找到您所请求的内容，文件可能已过期或系统正在维护中...</p>`
	UriNotAllow  = `<h1 style="text-align:center;font-family:cursive;">401</h1><p style="text-align:center;font-family: cursive;">:(</><p style="text-align:center;font-family: monospace;">当前系统设置为不允许查看目录！</p>`
)

func main() {
	if len(os.Args) >= 2 {
		dir = os.Args[1]
	}
	if port == "" {
		wlog("请将端口号(1~65535)写于执行文件的文件名中，如\"static-881-fortest.exe\"")
	} else {
		http.Handle("/js/", http.FileServer(http.Dir(dir)))
		http.Handle("/css/", http.FileServer(http.Dir(dir)))
		http.Handle("/media/", http.FileServer(http.Dir(dir)))
		http.Handle("/media.html", http.FileServer(http.Dir(dir)))

		for _, disk := range GetLogicalDrives() {
			// if disk == "c:" {
			// 	continue
			// }
			wlog("Handle " + disk)
			http.HandleFunc("/"+disk+"/", StaticServerMaker(disk))
		}

		http.HandleFunc("/", StaticServer)

		wlog("Listening " + port + "...")
		if err := http.ListenAndServe(":"+port, nil); err != nil {
			wlog(err.Error())
		}
	}
}

func StaticServer(w http.ResponseWriter, req *http.Request) {
	fileInfo, err := os.Stat(dir + req.RequestURI)
	if err != nil {
		w.WriteHeader(404)
		w.Write([]byte(FileNotFound))
	} else {
		if fileInfo.IsDir() {
			w.WriteHeader(401)
			w.Write([]byte(UriNotAllow))
		} else {
			w.Header().Set("Content-Disposition", "attachment; filename:\""+fileInfo.Name()+"\"")
			w.Header().Set("Content-Type", "application/force-download")
			w.Header().Set("Content-Length", strconv.FormatInt(fileInfo.Size(), 10))
			staticHandler := http.FileServer(http.Dir(dir))
			staticHandler.ServeHTTP(w, req)
		}
	}
}

func StaticServerMaker(disk string) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		fileInfo, err := os.Stat(disk + req.RequestURI[3:])
		if err != nil {
			w.WriteHeader(404)
			w.Write([]byte(FileNotFound))
		} else {
			if fileInfo.IsDir() {
				w.WriteHeader(401)
				w.Write([]byte(UriNotAllow))
			} else {
				w.Header().Set("Content-Disposition", "attachment; filename:\""+fileInfo.Name()+"\"")
				w.Header().Set("Content-Type", "application/force-download")
				w.Header().Set("Content-Length", strconv.FormatInt(fileInfo.Size(), 10))
				http.StripPrefix("/"+disk+"/", http.FileServer(http.Dir(disk))).ServeHTTP(w, req)
			}
		}
	}
}

func wlog(content string) {
	logPath := dir + "/" + time.Now().Format("2006-01-02") + ".log"
	logfile, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE, 0)
	if err != nil {
		log.Printf("Create log file failed: %s\r\n", err.Error())
		return
	}
	defer logfile.Close()
	log.New(logfile, "", log.Ldate|log.Ltime).Printf(content + "\r\n")
}

func GetLogicalDrives() []string {
	kernel32 := syscall.MustLoadDLL("kernel32.dll")
	GetLogicalDrives := kernel32.MustFindProc("GetLogicalDrives")
	n, _, _ := GetLogicalDrives.Call()
	s := strconv.FormatInt(int64(n), 2)

	var drives_all = []string{"a:", "b:", "c:", "d:", "e:", "f:", "g:", "h:", "i:", "j:", "k:", "l:", "m:", "n:", "o:", "p:", "q:", "r:", "s:", "t:", "u:", "v:", "w:", "x:", "y:", "z:"}
	temp := drives_all[0:len(s)]

	var d []string
	for i, v := range s {
		if v == 49 {
			l := len(s) - i - 1
			d = append(d, temp[l])
		}
	}

	var drives []string
	for i, v := range d {
		drives = append(drives[i:], append([]string{v}, drives[:i]...)...)
	}
	return drives

}

func init() {
	file, _ := exec.LookPath(os.Args[0])
	path, _ := filepath.Abs(file)
	index := strings.LastIndex(path, "\\")
	dotIndex := strings.LastIndex(path, ".")

	match := portExp.FindStringSubmatch(path[index+1 : dotIndex])
	if match != nil {
		port = match[1]
	}

	dir = strings.Replace(path[:index], "\\", "/", -1)
}
