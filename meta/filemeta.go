package meta

import (
	mydb "fileStore_server/db"
	"fmt"
	"sort"
)

// 文件元信息结构
type FileMeta struct {
	FileSha1 string // 文件的唯一标识
	FileName string
	FileSize int64
	Location string // 存在本地的文件路径
	UploadAt string // 时间戳
}

var (
	fileMetas map[string]FileMeta
)

func init() {
	fileMetas = make(map[string]FileMeta)
}

// 新增/更新文件元信息
func UpdateFileMeta(fmeta FileMeta) {
	fileMetas[fmeta.FileSha1] = fmeta
}

// 新增/更新文件元信息到mysql中
func UpdateFileMetaDB(fmeta FileMeta) bool {
	fmt.Println(fmeta.FileSha1, fmeta.FileName, fmeta.FileSize, fmeta.Location)
	return mydb.OnFileUploadFinished(fmeta.FileSha1, fmeta.FileName,
		fmeta.FileSize, fmeta.Location)
}

// 通过sha1获取文件元信息
func GetFileMeta(fileSha1 string) FileMeta {
	return fileMetas[fileSha1]
}

// 从mysql获取文件元信息
func GetFileMetaDB(filesha1 string) (fmeta FileMeta, err error) {
	var (
		tfile mydb.TableFile
	)

	fmeta = FileMeta{}

	if tfile, err = mydb.GetFileMeta(filesha1); err != nil {
		return
	}

	fmeta = FileMeta{
		FileSha1: tfile.FileHash,
		FileName: tfile.FileName.String,
		FileSize: tfile.FileSize.Int64,
		Location: tfile.FileAddr.String,
	}
	return
}

// 获取count数量的文件元信息列表
func GetLastFileMetas(count int) []FileMeta {
	var (
		fMetaArray []FileMeta
		fmeta      FileMeta
	)

	fMetaArray = make([]FileMeta, len(fileMetas))
	for _, fmeta = range fileMetas {
		fMetaArray = append(fMetaArray, fmeta)
	}

	sort.Sort(ByUploadTime(fMetaArray))
	return fMetaArray[0:count]
}

// 删除元信息
func RemoveFileMeta(fileSha1 string) {
	delete(fileMetas, fileSha1)
}
