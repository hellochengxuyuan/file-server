package db

import (
	"database/sql"
	mydb "fileStore_server/db/mysql"
	"fmt"
)

// 文件表对应的一些字段
type TableFile struct {
	FileHash string
	FileName sql.NullString
	FileSize sql.NullInt64
	FileAddr sql.NullString
}

// 文件上传完成，保存meta
func OnFileUploadFinished(filehash string, filename string,
	filesize int64, fileaddr string) bool {
	var (
		stmt     *sql.Stmt
		err      error
		execResp sql.Result
		rf       int64
	)
	stmt, err = mydb.DBConn().Prepare("insert ignore into " +
		"tbl_file(`file_sha1`,`file_name`,`file_size`,`file_addr`," +
		"`status`) values (?,?,?,?,1)")
	defer stmt.Close()

	if err != nil {
		fmt.Println("Failed to prepare statement,err: ", err.Error())
		return false
	}

	if execResp, err = stmt.Exec(filehash, filename, filesize, fileaddr); err != nil {
		fmt.Println(err.Error())
		return false
	}
	fmt.Println("exec执行成功 ", err)

	// 返回受插入，更新或删除影响的行数
	if rf, err = execResp.RowsAffected(); nil == err {
		fmt.Println("rf: ", rf)
		if rf <= 0 {
			fmt.Printf("File with hash: %s "+
				"has been upload before\n", filehash)
			return false
		}
		return true
	}
	return false
}

// 从mysql获取文件元信息
func GetFileMeta(filesha1 string) (tfile TableFile, err error) {
	var (
		stmt *sql.Stmt
	)

	stmt, err = mydb.DBConn().Prepare("select file_sha1,file_name," +
		"file_size,file_addr from tbl_file where file_sha1=? and status=1 limit 1")
	defer stmt.Close()

	// 初始化，这样不会指向空
	tfile = TableFile{}

	if err != nil {
		fmt.Println(err.Error())
		return
	}

	if err = stmt.QueryRow(filesha1).Scan(&tfile.FileHash, &tfile.FileName,
		&tfile.FileSize, &tfile.FileAddr); err != nil {
		fmt.Println(err.Error())
		return
	}
	return
}

// UpdateFileLocation : 更新文件的存储地址(如文件被转移了)
func UpdateFileLocation(filehash string, fileaddr string) bool {
	stmt, err := mydb.DBConn().Prepare(
		"update tbl_file set `file_addr`=? where  `file_sha1`=? limit 1")
	if err != nil {
		fmt.Println("预编译sql失败, err:" + err.Error())
		return false
	}
	defer stmt.Close()

	ret, err := stmt.Exec(fileaddr, filehash)
	if err != nil {
		fmt.Println(err.Error())
		return false
	}
	if rf, err := ret.RowsAffected(); nil == err {
		fmt.Println("rf= ", rf)
		if rf <= 0 {
			fmt.Printf("更新文件location失败, filehash:%s", filehash)
			return false
		}
		return true
	}
	return false
}
