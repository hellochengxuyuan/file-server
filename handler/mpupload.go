package handler

import (
	"fileStore_server/static/util"
	"fmt"
	"github.com/gomodule/redigo/redis"
	"math"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	rPool "fileStore_server/cache/redis"
	dblayer "fileStore_server/db"
)

// 分块上传信息
type MultiPartUploadInfo struct {
	FileHash   string
	FileSize   int
	UploadID   string //每次上传都会生成的唯一ID，同一个文件不同时间上传也会有不同ID
	ChunkSize  int
	ChunkCount int
}

// 初始化分块上传
func InitialMultipartUploadHandler(w http.ResponseWriter, r *http.Request) {
	var (
		username string
		filehash string
		filesize int
		err      error
		rConn    redis.Conn
		upInfo   MultiPartUploadInfo
	)

	// 1.解析用户请求信息
	r.ParseForm()
	username = r.Form.Get("username")
	filehash = r.Form.Get("filehash")
	if filesize, err = strconv.Atoi(r.Form.Get("filesize")); err != nil {
		w.Write(util.NewRespMsg(-1,
			"params invalid", nil).JSONBytes())
		return
	}

	// 2.获得redis的一个连接
	rConn = rPool.RedisPool().Get()
	defer rConn.Close()

	// 3.生成分块上传的初始化信息
	upInfo = MultiPartUploadInfo{
		FileHash:   filehash,
		FileSize:   filesize,
		UploadID:   username + fmt.Sprintf("%x", time.Now().UnixNano()),
		ChunkSize:  5 * 1024 * 1024, // 5MB
		ChunkCount: int(math.Ceil(float64(filesize) / (5 * 1024 * 1024))),
	}

	// 4.将初始化信息写入到redis缓存
	rConn.Do("HSET", "MP_"+upInfo.UploadID,
		"chunkcount", upInfo.ChunkCount)
	rConn.Do("HSET", "MP_"+upInfo.UploadID,
		"filehash", upInfo.FileHash)
	rConn.Do("HSET", "MP_"+upInfo.UploadID,
		"filesize", upInfo.FileSize)

	// 5.将响应初始化数据返回给客户端
	w.Write(util.NewRespMsg(0, "OK", upInfo).JSONBytes())
}

// 上传分块文件
func UploadPartHandler(w http.ResponseWriter, r *http.Request) {
	var (
		username   string
		uploadID   string
		chunkIndex string
		rConn      redis.Conn
		file       *os.File
		err        error
		buf        []byte
		n          int
		dir        string
	)

	// 1.解析用户请求信息
	r.ParseForm()
	username = r.Form.Get("username")

	username = username

	uploadID = r.Form.Get("uploadid")
	chunkIndex = r.Form.Get("index")

	// 2.获得redis连接池中的一个连接
	rConn = rPool.RedisPool().Get()
	defer rConn.Close()

	// 3.获得文件句柄，用于存储分块内容
	dir = "/home/haha/" + uploadID + "/" + chunkIndex
	os.MkdirAll(path.Dir(dir), 0744)
	file, err = os.Create(dir)
	defer file.Close()
	if err != nil {
		w.Write(util.NewRespMsg(-1,
			"Upload part failed", nil).JSONBytes())
		return
	}

	buf = make([]byte, 1024*1024)
	for {
		n, err = r.Body.Read(buf)
		file.Write(buf[:n])
		if err != nil {
			break
		}
	}

	// 4.更新redis缓存状态
	rConn.Do("HSET", "MP_"+uploadID, "chkidx_"+chunkIndex, 1)

	// 5.返回处理结果到客户端
	w.Write(util.NewRespMsg(0, "OK", nil).JSONBytes())
}

// 通知上传合并
func CompleteUploadHandler(w http.ResponseWriter, r *http.Request) {
	var (
		username string
		uploadID string
		filehash string
		filesize int
		filename string
		rConn    redis.Conn
		//err        error
		data       []interface{}
		totalCount int
		chunkCount int
	)

	// 1.解析请求参数
	r.ParseForm()
	username = r.Form.Get("username")
	uploadID = r.Form.Get("uploadid")
	filehash = r.Form.Get("filehash")
	filesize, _ = strconv.Atoi(r.Form.Get("filesize"))
	filename = r.Form.Get("filename")

	// 2.获得redis连接池中的一个连接
	rConn = rPool.RedisPool().Get()
	defer rConn.Close()

	// 3.通过uploadid查询redis并判断是否所有分块上传完成
	data, _ = redis.Values(rConn.Do("HGETALL", "MP_"+uploadID))
	//if err != nil {
	//	w.Write(util.NewRespMsg(-1, "complete upload failed", nil).JSONBytes())
	//	return
	//}

	for i := 0; i < len(data); i += 2 {
		k := string(data[i].([]byte))
		v := string(data[i+1].([]byte))
		if k == "chunkcount" {
			totalCount, _ = strconv.Atoi(v)
		} else if strings.HasPrefix(k, "chkidx_") && v == "1" {
			chunkCount++
		}
	}
	if totalCount != chunkCount {
		w.Write(util.NewRespMsg(-2, "invalid request",
			nil).JSONBytes())
		return
	}

	// 4.TODO:合并分块

	// 5.更新唯一文件表及文件表
	dblayer.OnFileUploadFinished(filehash, filename, int64(filesize),
		"")
	dblayer.OnUserFileUploadFinished(username, filehash, filename,
		int64(filesize))

	// 6.响应处理结果
	w.Write(util.NewRespMsg(0, "OK", nil).JSONBytes())
}
