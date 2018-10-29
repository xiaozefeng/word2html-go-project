package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"strings"
)

const (
	postUrl     = "https://s10.aconvert.com/convert/convert-batch-win.php" // 文件上传api
	savePath    = "/Users/xiaozefeng/Desktop/result/"                      // 保存html文件目录
	downloadUrl = "https://s10.aconvert.com/convert/p3r68-cdx67/"          // 下载转换好的htmlapi
	dir         = "/Users/xiaozefeng/Desktop/resumes/"                     // doc 目录
)

func main() {
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		log.Fatal(err)
	}
	// 结束信号
	signal := make(chan int)
	count := 0
	for i, file := range files {
		filename := file.Name()
		if !(strings.HasSuffix(filename, "doc") || strings.HasSuffix(filename, "docx")) {
			fmt.Println("continue", filename)
			continue
		}
		count++
		go func(filename string, no int) {
			wh := &Word2Html{}
			err := wh.process(dir + filename)
			if err != nil {
				fmt.Printf("err = %+v\n", err)
			}
			signal <- no
		}(filename, i)
	}

	for i := 0; i < count; i++ {
		fmt.Printf("no:%d finished\n", <-signal)
	}
	fmt.Println("done")
}

type Word2Html struct{}

//filename 文件名
func (wh *Word2Html) process(filepath string) error {
	ur, err := wh.PostFile(filepath)
	if err != nil {
		return err
	}
	fmt.Printf("ur = %+v\n", ur)
	url := fmt.Sprintf("%s%s", downloadUrl, ur.Filename)
	return wh.Download(url, ur.Filename)
}

// 上传文件结构体
type UploadResult struct {
	Filename string `json:"filename"`
	Ext      string `json:"ext"`
	Server   string `json:"server"`
	State    string `json:"state"`
}

// 上传文件
func (wh *Word2Html) PostFile(filename string) (*UploadResult, error) {
	bodyBuf := &bytes.Buffer{}
	writer := multipart.NewWriter(bodyBuf)

	fileWriter, err := writer.CreateFormFile("file", filename)
	if err != nil {
		panic("error writing to buffer")
	}

	fmt.Printf("filename=%s\n", filename)

	fh, err := os.Open(filename)
	if err != nil {
		panic("打开文件失败")
	}

	_, err = io.Copy(fileWriter, fh)
	if err != nil {
		panic("拷贝文件失败")
	}

	writer.WriteField("targetformat", "html")
	writer.WriteField("code", "86000")
	writer.WriteField("filelocation", "local")

	contentType := writer.FormDataContentType()
	defer writer.Close()
	resp, err := http.Post(postUrl, contentType, bodyBuf)
	if err != nil {
		panic("上传文件失败")
	}
	defer resp.Body.Close()
	content, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}
	fmt.Printf("%s\n", string(content))

	ur := &UploadResult{}
	err = json.Unmarshal(content, ur)
	if err != nil {
		panic(err)
	}
	return ur, nil
}

func (wh *Word2Html) Download(url, filename string) error {
	fmt.Printf("downloadUrl=%s\n", url)
	resp, err := http.Get(url)
	if err != nil {
		return errors.New("download error")
	}
	defer resp.Body.Close()

	content, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return errors.New("read error")
	}
	err = ioutil.WriteFile(savePath+filename, content, 0644)
	if err != nil {
		return errors.New("write error")
	}
	return nil
}
