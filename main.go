package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"encoding/json"
	"io"
	"io/ioutil"
)

// 声明程序启动参数变量（全局）
var (
	help bool
	addr string
	port int
)

// 程序所在路径
var currentDir string

// 初始化参数 & 参数说明
func init() {
	flag.BoolVar(&help, "h", false, "Show help.")
	flag.StringVar(&addr, "a", "", "Specify the server listening address.")
	flag.IntVar(&port, "p", 80, "Specify the server listening port.")
}

// isFileExist 用于判断文件是否存在
func isFileExist(path string) bool {
	_, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}
	return true
}

// checkExtension 检查媒体文件扩展名是否受支持
func checkExtension(filename string) bool {
	ext := path.Ext(filename)
	supported := []string{".mp3", ".ogg", ".m4a", ".aac", ".wav"}
	for _, value := range supported {
		if strings.EqualFold(ext, value) {
			return true
		}
	}
	return false
}

// readMusicDir 获取音乐文件夹指定路径的目录和音频文件列表
func readMusicDir(folder string) (musicList []MusicListElement, subFolderList []string) {
	// 初始化空 slice
	musicList = make([]MusicListElement, 0, 10)
	subFolderList = make([]string, 0, 10)

	// 请求时提交 null 字符串则视为空字符串
	if folder == "null" {
		folder = ""
	}

	// 获取文件列表
	fileList, err := ioutil.ReadDir(path.Join(currentDir, "music", folder))
	if err != nil {
		log.Fatal(err)
	}
	// 分离目录和文件
	for _, file := range fileList {
		if file.IsDir() {
			// 如果是目录，直接添加到 subFolderList
			subFolderList = append(subFolderList, folder + file.Name())
		} else {
			fileFullPath := path.Join(currentDir, "music", folder, file.Name())
			// 如果是文件，先检查扩展名
			if checkExtension(fileFullPath) {
				// 将文件信息添加到 musicList
				fileInfo, _ := os.Stat(fileFullPath)
				musicList = append(musicList, MusicListElement {
					FileName: file.Name(),
					FileSize: fileInfo.Size(),
					ModifiedTime: strconv.FormatInt(fileInfo.ModTime().Unix(), 10),
				})
			}
		}
	}
	return
}

// api 实现 api.php
func api(writer http.ResponseWriter, request *http.Request) {
	if request.Method == http.MethodPost {
		// 解析 POST 表单
		err := request.ParseForm()
		if err != nil {
			log.Fatal(err)
			response, _ := json.Marshal(NewErrorResponse(500, err.Error()))
			io.WriteString(writer, string(response))
			return
		}

		// 识别 do 参数
		switch request.Form.Get("do") {
		case "getfilelist":
			folder := request.Form.Get("folder")
			musicList, subFolderList := readMusicDir(folder)
			response, _ := json.Marshal(NewFileListResponse(musicList, subFolderList))
			io.WriteString(writer, string(response))
		default:
			response, _ := json.Marshal(NewErrorResponse(400, "unsupported operation"))
			io.WriteString(writer, string(response))
		}
	} else {
		// 如果请求方式不是 POST
		response, _ := json.Marshal(NewErrorResponse(400, "illegal request!"))
		io.WriteString(writer, string(response))
	}
}

func main() {
	// 解析程序启动参数
	flag.Parse()
	if help {
		flag.Usage()
		return
	}

	// 获取当前程序所在路径
	currentDir, _ = filepath.Abs(filepath.Dir(os.Args[0]))

	// 处理 HTTP 请求
	http.HandleFunc("/", func(writer http.ResponseWriter, request *http.Request) {
		switch request.URL.Path {
		case "/":
			// 首页
			http.ServeFile(writer, request, path.Join(currentDir, "static", "index.html"))
		case "/api.php":
			// API 接口路径
			api(writer, request)
		default:
			// 禁止在路径中使用 ".."，防止访问任意目录下的文件
			requireFilePath := strings.Replace(request.URL.Path, "..", "", -1)

			// 生成文件路径
			staticFileName := path.Join(currentDir, "static", requireFilePath)
			musicFileName := path.Join(currentDir, "music", requireFilePath)

			if isFileExist(staticFileName) {
				// 如果在 static 目录中能找到此文件，则传输此文件
				http.ServeFile(writer, request, staticFileName)
			} else if isFileExist(musicFileName) && checkExtension(musicFileName) {
				// 如果在 music 目录中能找到此文件，并且扩展名受支持，则传输此文件
				http.ServeFile(writer, request, musicFileName)
			} else {
				// 404
				http.NotFound(writer, request)
			}
		}
	})

	// 开启服务器
	http.ListenAndServe(addr+":"+strconv.Itoa(port), nil)
}
