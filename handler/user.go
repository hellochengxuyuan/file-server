package handler

import (
	dblayer "fileStore_server/db"
	"fileStore_server/static/util"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

const (
	pwd_salt = "*#890"
)

// 处理用户注册请求
func SignUpHandler(w http.ResponseWriter, r *http.Request) {
	var (
		bytes     []byte
		err       error
		username  string
		passwd    string
		encPasswd string
	)

	if r.Method == http.MethodGet {
		if bytes, err = ioutil.ReadFile("./static/view/signup.html"); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.Write(bytes)
		return
	}

	r.ParseForm()
	username = r.Form.Get("username")
	passwd = r.Form.Get("password")

	if len(username) < 3 || len(passwd) < 5 {
		w.Write([]byte("Invalid parameter"))
		return
	}

	encPasswd = util.Sha1([]byte(passwd + pwd_salt))
	if dblayer.UserSignUp(username, encPasswd) {
		w.Write([]byte("SUCCESS"))
	} else {
		w.Write([]byte("FAILED"))
	}
}

// 登陆接口
func SignInHandler(w http.ResponseWriter, r *http.Request) {
	var (
		username  string
		password  string
		encPasswd string
		isTrue    bool
		token     string
		bytes     []byte
		err       error
		resp      util.RespMsg
	)

	if r.Method == http.MethodGet {
		if bytes, err = ioutil.ReadFile("./static/view/signin.html"); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.Write(bytes)
		return
	}

	r.ParseForm()
	username = r.Form.Get("username")
	password = r.Form.Get("password")

	encPasswd = util.Sha1([]byte(password + pwd_salt))

	// 1.校验用户密码
	if isTrue = dblayer.UserSignIn(username, encPasswd); !isTrue {
		w.Write([]byte("FAILED"))
		return
	}

	// 2.生成访问凭证(token)
	token = GenToken(username)
	if !dblayer.UpdateToken(username, token) {
		w.Write([]byte("FAILED"))
		return
	}

	// 3.登陆成功后重定向到首页
	//w.Write([]byte("http://" + r.Host + "/static/view/home.html"))
	resp = util.RespMsg{
		Code: 0,
		Msg:  "OK",
		Data: struct {
			Location string
			Username string
			Token    string
		}{
			Location: "http://" + r.Host + "/static/view/home.html",
			Username: username,
			Token:    token,
		},
	}
	w.Write(resp.JSONBytes())
}

// 查询用户信息
func UserInfoHandler(w http.ResponseWriter, r *http.Request) {
	var (
		username string
		//token    string
		err      error
		userInfo dblayer.User
		resp     util.RespMsg
	)

	// 1. 解析请求参数
	r.ParseForm()
	username = r.Form.Get("username")
	//token = r.Form.Get("token")

	// 2. 验证token是否有效
	//if !IsTokenValid(token) {
	//	w.WriteHeader(http.StatusForbidden)
	//	return
	//}

	// 3. 查询用户信息
	if userInfo, err = dblayer.GetUserInfo(username); err != nil {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	// 4. 组装并响应用户数据
	resp = util.RespMsg{
		Code: 0,
		Msg:  "OK",
		Data: userInfo,
	}
	w.Write(resp.JSONBytes())

}

// 验证token是否有效
func IsTokenValid(token string) bool {
	if len(token) != 40 {
		return false
	}
	// TODO: 判断token的时效性，是否过期
	// TODO: 从数据库表tbl_user_token查询username对应的token信息
	// TODO: 对比两个token是否一致
	return true
}

func GenToken(username string) string {
	var (
		ts          string
		tokenPrefix string
	)

	// 40位字符：md5(username + timestamp + +token_salt) + timestamp[:8]
	ts = fmt.Sprintf("%x", time.Now().Unix())
	tokenPrefix = util.MD5([]byte(username + ts + "_tokensalf"))
	return tokenPrefix + ts[:8]
}
