package handler

import (
	"encoding/json"
	cmn "fileStore_server/common"
	cfg "fileStore_server/config"
	dblayer "fileStore_server/db"
	"fileStore_server/meta"
	"fileStore_server/mq"
	"fileStore_server/static/util"
	"fileStore_server/store/oss"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"os"
	"strconv"
	"time"
)

// 处理文件上传
func UploadHandler(w http.ResponseWriter, r *http.Request) {
	var (
		readResp   []byte
		err        error
		file       multipart.File
		fileHeader *multipart.FileHeader
		newFile    *os.File
		fileMeta   meta.FileMeta
		username   string
	)

	if r.Method == "GET" {
		// 返回上传html页面
		if readResp, err = ioutil.ReadFile("./static/view/index.html"); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.Write(readResp)
	} else if r.Method == "POST" {
		// 接收文件流以及存储到本地目录
		file, fileHeader, err = r.FormFile("file")
		defer file.Close()

		if err != nil {
			fmt.Printf("Failed to get date, err: %s\n", err.Error())
			return
		}

		fileMeta = meta.FileMeta{
			FileName: fileHeader.Filename,
			Location: "/home/alien/wall/" + fileHeader.Filename,
			UploadAt: time.Now().Format("2006-01-02 15:04:05"), //以“”中的格式输出时间，
			// 只有填“2006-01-02 15:04:05”这个时间，这个函数才会正确执行，不然输出的时间会错误
		}

		newFile, err = os.Create(fileMeta.Location)
		defer newFile.Close()
		if err != nil {
			fmt.Printf("Failed to create file, err: %s\n", err.Error())
			return
		}

		// TODO: 2种复制文件的方法，第一种复制后无上传到oss，采用第2种方法才可以
		// TODO: 后面证实第1种方法也可以，但上传前要执行 newFile.Seek(0, 0) 这个函数后才能上传，不然会报错
		if fileMeta.FileSize, err = io.Copy(newFile, file); err != nil {
			fmt.Printf("Failed to save data into file, err: %s\n", err.Error())
			return
		}
		//data, _ := ioutil.ReadAll(file)
		//if _, err = newFile.Write(data); err != nil {
		//	fmt.Printf("Failed to save data into file, err: %s\n", err.Error())
		//	return
		//}

		newFile.Seek(0, 0)
		fileMeta.FileSha1 = util.FileSha1(newFile)
		fmt.Println(fileMeta.FileSha1, "hash")

		chenggong := meta.UpdateFileMetaDB(fileMeta)
		fmt.Println(chenggong)

		newFile.Seek(0, 0)

		ossPath := "oss/" + fileMeta.FileSha1
		//if err = oss.Bucket().PutObject(ossPath, newFile); err != nil {
		//	fmt.Println(err.Error())
		//	w.Write([]byte("Upload failed"))
		//	return
		//}
		//fileMeta.Location = ossPath
		data := mq.TransferData{
			FileHash:      fileMeta.FileSha1,
			CurLocation:   fileMeta.Location,
			DestLocation:  ossPath,
			DestStoreType: cmn.StoreOSS,
		}
		pubData, _ := json.Marshal(data)
		suc := mq.Publish(cfg.TransExchangeName,
			cfg.TransOSSRoutingKey,
			pubData)
		if !suc {
			// TODO： 加入重新发送消息逻辑
		}

		//  更新用户文件表记录
		r.ParseForm()
		username = r.Form.Get("username")
		if dblayer.OnUserFileUploadFinished(username, fileMeta.FileSha1,
			fileMeta.FileName, fileMeta.FileSize) {
			fmt.Printf("%+v\n", fileMeta)
			http.Redirect(w, r, "/static/view/home.html", http.StatusFound)
		} else {
			w.Write([]byte("Upload Failed."))
		}
	}
}

// 上传已完成
func UploadStuctHandler(w http.ResponseWriter, r *http.Request) {
	io.WriteString(w, "Upload finished.")
}

// 获取文件元信息
func GetFileMetaHandler(w http.ResponseWriter, r *http.Request) {
	var (
		err      error
		fileHash string
		fmeta    meta.FileMeta
		bytes    []byte
	)

	r.ParseForm()

	fileHash = r.Form["fileHash"][0]
	//fmeta = meta.GetFileMeta(fileHash)
	if fmeta, err = meta.GetFileMetaDB(fileHash); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if bytes, err = json.Marshal(fmeta); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Write(bytes)
}

// 查询批量的文件元信息
func FileQueryHandler(w http.ResponseWriter, r *http.Request) {
	var (
		username string
		limitCnt int
		userFile []dblayer.UserFile
		bytes    []byte
		err      error
	)

	r.ParseForm()
	username = r.Form.Get("username")
	limitCnt, _ = strconv.Atoi(r.Form.Get("limit"))
	//fileMetas = meta.GetLastFileMetas(limitCnt)
	if userFile, err = dblayer.QueryUserFileMetas(username, limitCnt); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if bytes, err = json.Marshal(userFile); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Write(bytes)
}

// 下载文件
func DownloadHandler(w http.ResponseWriter, r *http.Request) {
	var (
		err      error
		fsha1    string
		fileMeta meta.FileMeta
		file     *os.File
		bytes    []byte
	)

	r.ParseForm()
	fsha1 = r.Form.Get("fileHash")
	fileMeta = meta.GetFileMeta(fsha1)

	file, err = os.Open(fileMeta.Location)
	defer file.Close()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if bytes, err = ioutil.ReadAll(file); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/octect-stream")
	w.Header().Set("Content-Descrption",
		"attachment;filename=\""+fileMeta.FileName+"\"")
	w.Write(bytes)
}

// 修改文件元信息（以重命名为例）
func FileUpdateMetaHandler(w http.ResponseWriter, r *http.Request) {
	var (
		fileSha1    string
		newFileName string
		fileMeta    meta.FileMeta
		opType      string
		bytes       []byte
		err         error
	)

	r.ParseForm()

	opType = r.Form.Get("op")
	fileSha1 = r.Form.Get("fileHash")
	newFileName = r.Form.Get(("filename"))

	if opType != "0" {
		w.WriteHeader(http.StatusForbidden)
	}

	if r.Method != "POST" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	fileMeta = meta.GetFileMeta(fileSha1)
	fileMeta.FileName = newFileName
	meta.UpdateFileMeta(fileMeta)

	if bytes, err = json.Marshal(fileMeta); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(bytes)
}

// 删除文件及信息
func FileDelHandler(w http.ResponseWriter, r *http.Request) {
	var (
		fileSha1 string
		fileMeta meta.FileMeta
	)

	r.ParseForm()
	fileSha1 = r.Form.Get("fileHash")
	fileMeta = meta.GetFileMeta(fileSha1)
	os.Remove(fileMeta.Location)

	meta.RemoveFileMeta(fileSha1)

	w.WriteHeader(http.StatusOK)
}

// 尝试秒传接口
func TryFastUploadHandler(w http.ResponseWriter, r *http.Request) {
	var (
		username string
		filehash string
		filename string
		filesize int
		fmeta    meta.FileMeta
		err      error
		resp     util.RespMsg
	)

	// 1.解析请求参数
	r.ParseForm()
	username = r.Form.Get("username")
	filehash = r.Form.Get("filehash")
	filename = r.Form.Get("filename")
	filesize, _ = strconv.Atoi(r.Form.Get("filesize"))

	// 2.从文件表中查询相同hash的文件记录
	fmeta, err = meta.GetFileMetaDB(filehash)
	if err != nil {
		fmt.Println(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		if fmeta.FileSha1 == "" {
			resp = util.RespMsg{
				Code: -1,
				Msg:  "秒传失败，请访问普通上传借口",
			}
			w.Write(resp.JSONBytes())
		}
		return
	}

	//// 3.查不到记录则返回秒传失败
	if fmeta.FileSha1 == "" {
		resp = util.RespMsg{
			Code: -1,
			Msg:  "秒传失败，请访问普通上传借口",
		}
		w.Write(resp.JSONBytes())
		return
	}

	// 4.上传过则讲文件信息写入用户文件表，返回成功
	if dblayer.OnUserFileUploadFinished(username, filehash,
		filename, int64(filesize)) {
		resp = util.RespMsg{
			Code: 0,
			Msg:  "秒传成功",
		}
		w.Write(resp.JSONBytes())
		return
	}
	resp = util.RespMsg{
		Code: -2,
		Msg:  "秒传失败，请稍后重试",
	}
	w.Write(resp.JSONBytes())
	return
}

func DownloadURLHandler(w http.ResponseWriter, r *http.Request) {
	var (
		filehash string
		tfile    dblayer.TableFile
		//err      error
		signedURL string
	)

	r.ParseForm()
	filehash = r.Form.Get("filehash")

	// 从文件表查找记录
	tfile, _ = dblayer.GetFileMeta(filehash)

	// TODO:判断文件存在oss还是在ceph

	signedURL = oss.DownloadURL(tfile.FileAddr.String)
	w.Write([]byte(signedURL))
}
