package db

import (
	"database/sql"
	mydb "fileStore_server/db/mysql"
	"fmt"
)

// 通过用户名和密码完成user表的注册操作
func UserSignUp(username string, passwd string) bool {
	var (
		stmt     *sql.Stmt
		err      error
		execResp sql.Result
		rf       int64
	)

	stmt, err = mydb.DBConn().Prepare("insert ignore into tbl_user (`user_name`," +
		"`user_pwd`) values (?,?)")
	defer stmt.Close()

	if err != nil {
		fmt.Println("Failed to insert,err: " + err.Error())
		return false
	}

	if execResp, err = stmt.Exec(username, passwd); err != nil {
		fmt.Println("Failed to insert,err: " + err.Error())
		return false
	}

	if rf, err = execResp.RowsAffected(); err == nil && rf > 0 {
		return true
	}
	return false
}

// 判断密码是否一致
func UserSignIn(username string, encpwd string) bool {
	var (
		stmt  *sql.Stmt
		err   error
		rows  *sql.Rows
		pRows []map[string]interface{}
	)

	stmt, err = mydb.DBConn().Prepare("select * from tbl_user where " +
		"user_name=? limit 1")
	defer stmt.Close()

	if err != nil {
		fmt.Println(err.Error())
		return false
	}

	if rows, err = stmt.Query(username); err != nil {
		fmt.Println(err.Error())
		return false
	} else if rows == nil {
		fmt.Println("username not found: " + username)
		return false
	}

	pRows = mydb.ParseRows(rows)
	if len(pRows) > 0 && string(pRows[0]["user_pwd"].([]byte)) == encpwd {
		return true
	}
	return false
}

// 刷新用户登录的token
func UpdateToken(username string, token string) bool {
	var (
		stmt *sql.Stmt
		err  error
	)

	stmt, err = mydb.DBConn().Prepare("replace into tbl_user_token " +
		"(`user_name`,`user_token`) values (?,?)")
	defer stmt.Close()

	if err != nil {
		fmt.Println(err.Error())
		return false
	}

	if _, err = stmt.Exec(username, token); err != nil {
		fmt.Println(err.Error())
		return false
	}
	return true
}

// 用户信息结构体
type User struct {
	Username     string
	Email        string
	Phone        string
	SignupAt     string
	LastActiveAt string
	Status       int
}

// 查询用户信息
func GetUserInfo(username string) (user User, err error) {
	var (
		stmt *sql.Stmt
	)

	user = User{}

	stmt, err = mydb.DBConn().Prepare("select user_name,signup_at " +
		"from tbl_user where user_name=? limit 1")
	defer stmt.Close()
	if err != nil {
		fmt.Println(err.Error())
		return user, err
	}

	// 执行查询的操作
	if err = stmt.QueryRow(username).Scan(&user.Username, &user.SignupAt); err != nil {
		return user, err
	}
	return user, nil
}
