package db

import (
	"database/sql"
	mydb "fileStore_server/db/mysql"
	"fmt"
	"time"
)

// 用户文件表结构体
type UserFile struct {
	UserName    string
	FileHash    string
	FileName    string
	FileSize    int64
	UploadAt    string
	LastUpdated string
}

// 更新用户文件表
func OnUserFileUploadFinished(username string, filehash string,
	filename string, filesize int64) bool {
	var (
		stmt *sql.Stmt
		err  error
	)

	stmt, err = mydb.DBConn().Prepare("insert ignore into tbl_user_file " +
		"(`user_name`,`file_sha1`,`file_name`,`file_size`,`upload_at`) values " +
		"(?,?,?,?,?)")
	defer stmt.Close()

	if err != nil {
		return false
	}

	if _, err = stmt.Exec(username, filehash, filename, filesize, time.Now()); err != nil {
		return false
	}
	return true
}

// 批量获取用户文件信息
func QueryUserFileMetas(username string, limit int) (
	userFile []UserFile, err error) {
	var (
		stmt  *sql.Stmt
		rows  *sql.Rows
		ufile UserFile
	)

	stmt, err = mydb.DBConn().Prepare("select file_sha1,file_name," +
		"file_size,upload_at,last_update from tbl_user_file " +
		"where user_name=? limit ?")
	defer stmt.Close()

	if err != nil {
		return nil, err
	}

	if rows, err = stmt.Query(username, limit); err != nil {
		return nil, err
	}

	for rows.Next() {
		ufile = UserFile{}
		if err = rows.Scan(&ufile.FileHash, &ufile.FileName, &ufile.FileSize,
			&ufile.UploadAt, &ufile.LastUpdated); err != nil {
			fmt.Println(err.Error())
			break
		}
		userFile = append(userFile, ufile)
	}
	return userFile, nil
}
