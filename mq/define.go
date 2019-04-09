package mq

import cmn "fileStore_server/common"

// 转移队列种消息载体的结构格式
type TransferData struct {
	FileHash      string
	CurLocation   string
	DestLocation  string
	DestStoreType cmn.StoreType
}
