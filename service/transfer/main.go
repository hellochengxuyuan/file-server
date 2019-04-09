package main

import (
	"encoding/json"
	"fileStore_server/config"
	dblyer "fileStore_server/db"
	"fileStore_server/mq"
	"fileStore_server/store/oss"
	"fmt"
	"log"
	"os"
)

// 处理文件转移的真正逻辑
func ProcessTransfer(msg []byte) bool {
	var (
		pubData mq.TransferData
		err     error
		file    *os.File
	)

	log.Println(string(msg), "OK")

	// 1.解析msg
	pubData = mq.TransferData{}
	if err = json.Unmarshal(msg, &pubData); err != nil {
		log.Println(err.Error())
		return false
	}
	fmt.Println("ok1: ", pubData.DestLocation)

	// 2. 根据临时存储文件路径，创建文件句柄
	if file, err = os.Open(pubData.CurLocation); err != nil {
		log.Println(err.Error())
		return false
	}

	fmt.Println("ok2: ", pubData.DestLocation)

	// 3. 通过文件句柄将文件内容读出来并且上传到oss
	if err = oss.Bucket().PutObject(pubData.DestLocation,
		file); err != nil {
		log.Println("上传到oss失败")
		log.Println(err.Error())
		return false
	}

	fmt.Println("ok3: ", pubData.DestLocation)

	// 4. 更新文件的存储路径到文件表
	_ = dblyer.UpdateFileLocation(pubData.FileHash, pubData.DestLocation)

	return true
}

func main() {
	if !config.AsyncTransferEnable {
		log.Println("异步转移文件功能目前被禁用，请检查相关配置")
	}
	log.Println("开始监听转移任务队列")
	mq.StartConsume(
		config.TransOSSQueueName,
		"transfer_oss",
		ProcessTransfer)
}
