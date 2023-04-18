package service

import (
	"net/http"
	"strconv"

	"crypto/sha256"
	"fmt"
	"net/url"
	"time"

	"strings"

	"github.com/gin-gonic/gin"
	database "titable.go/db"
	"titable.go/service/datainit"
	"titable.go/service/ical"
	"titable.go/service/mail"
)

// 助かりマックス
var layout = "2006-01-02 15:04:05.000"

//ログインおよびアカウント登録画面のサーバ側の関数がある

//ログイン画面のルーティングで呼ばれる
// /main にリダレクトする
func Login(ctx *gin.Context) {
	// Get DB connection
	db, err := database.GetConnection()
	if err != nil {
		ctx.String(http.StatusInternalServerError, err.Error())
		return
	}

	name, _ := ctx.GetPostForm("user")
	pw, _ := ctx.GetPostForm("pw")

	//ログインチェック
	var user database.User
	hpw := fmt.Sprintf("%x", sha256.Sum256([]byte(pw)))
	err = db.Get(&user, "SELECT * FROM users WHERE name=? AND password=?", name, hpw)
	if err != nil {
		ctx.HTML(http.StatusOK, "login.html", gin.H{"Title": "LOGIN", "Info": "ユーザネームまたはパスワードが誤っています"})
		return
	}
	uid := strconv.FormatUint(user.ID, 10)

	//zoom　URLを取得しDBへ格納
	// Get classes in DB
	var classes []database.Class
	var query string
	query = "SELECT * FROM classes WHERE uid=" + uid
	err = db.Select(&classes, query)
	if err != nil {
		ctx.String(http.StatusInternalServerError, err.Error())
		return
	}

	//Get class ZOOM url and insert DB
	//classテーブルから重複を許さず授業名をとってくる
	var class_names []string
	for _, class := range classes {
		if NotContain(class_names, class.Class) {
			class_names = append(class_names, class.Class)
		}
	}
	var d_classes []database.Class
	for _, s := range class_names {
		var class database.Class
		class.Class = s
		d_classes = append(d_classes, class)
	}
	var zoomurls []string
	for _, class := range d_classes {
		zoomurls, err = mail.GetURL(class, user)
		if err != nil {
			ctx.String(http.StatusInternalServerError, err.Error())
			return
		}
		fmt.Println("Login():授業", class.Class)
		fmt.Println("Login():url_num", len(zoomurls))
		if len(zoomurls) != 0 {
			nowtime := time.Now().Format(layout)
			for j := len(zoomurls) - 1; j >= 0; j-- {
				data := map[string]interface{}{"url": zoomurls[j], "uid": uid, "class": class.Class, "now": nowtime}
				//data := map[string]interface{}{"url": zoomurls[j], "uid": uid, "class": class.Class}
				if j == len(zoomurls)-1 {
					_, err := db.NamedExec("UPDATE classes SET url=(:url) WHERE uid=(:uid) AND start IN (SELECT MIN(start) FROM (SELECT start FROM classes WHERE class=(:class) AND uid=(:uid) AND start>(:now))tmp)", data)
					if err != nil {
						ctx.String(http.StatusInternalServerError, err.Error())
						return
					}
				} else {
					_, err := db.NamedExec("UPDATE classes SET url=(:url) WHERE uid=(:uid) AND start IN (SELECT MAX(start) FROM (SELECT start FROM classes WHERE class=(:class) AND uid=(:uid) AND start<(:now))tmp)", data)
					if err != nil {
						ctx.String(http.StatusInternalServerError, err.Error())
						return
					}
				}
				var lastupdate_class database.Class
				query = "SELECT * FROM classes WHERE url= '" + zoomurls[j] + "'"
				err = db.Get(&lastupdate_class, query)
				if err == nil {
					nowtime = lastupdate_class.Start.Format(layout)
				}
			}
		}
	}

	//Update ログイン時間を更新
	now := time.Now().Format(layout)
	data := map[string]interface{}{"uid": uid, "now": now}
	_, err = db.NamedExec("UPDATE users SET last_time=(:now) WHERE id=(:uid)", data)
	if err != nil {
		ctx.String(http.StatusInternalServerError, err.Error())
		return
	}

	ctx.SetCookie("id", uid, 1000, "/", "localhost", false, true)
	ctx.Redirect(http.StatusSeeOther, "/main")
}

func Signup(ctx *gin.Context) {
	// Get DB connection
	db, err := database.GetConnection()
	if err != nil {
		ctx.String(http.StatusInternalServerError, err.Error())
		return
	}

	name, _ := ctx.GetPostForm("user")
	email_id, _ := ctx.GetPostForm("mail")
	user_auth, _ := ctx.GetPostForm("user_auth")
	pw, _ := ctx.GetPostForm("pw")

	//avoid same email_id
	var user database.User
	err = db.Get(&user, "SELECT * FROM users WHERE email_id=?", email_id)
	if err == nil {
		ctx.HTML(http.StatusOK, "signup.html", gin.H{"Title": "SIGNUP", "InfoMail": "すでに使用されているメールアドレスです"})
		return
	}

	//avoid same name
	err = db.Get(&user, "SELECT * FROM users WHERE name=?", name)
	if err == nil {
		ctx.HTML(http.StatusOK, "signup.html", gin.H{"Title": "SIGNUP", "InfoUser": "すでに使用されているアカウント名です"})
		return
	}

	//１ヶ月前の時刻を取得
	now := time.Now()
	past := now.AddDate(0, 0, -30).Format(layout)

	//データベース: usersテーブルに追加
	hpw := fmt.Sprintf("%x", sha256.Sum256([]byte(pw)))
	data := map[string]interface{}{"name": name, "email_id": email_id, "user_auth": user_auth, "hpw": hpw, "last_time": past}
	res, err := db.NamedExec("INSERT INTO users (name, email_id, user_auth, password, last_time) VALUES (:name, :email_id, :user_auth, :hpw, :last_time)", data)
	if err != nil {
		ctx.String(http.StatusInternalServerError, err.Error())
		return
	}
	// id(AUTO_INCREMENT)を取得
	tmp, _ := res.LastInsertId()
	uid := strconv.FormatInt(tmp, 10)

	//iカレンダーを受け取る
	icalURL, _ := ctx.GetPostForm("iurl")
	classes, err := ical.GetCalData(icalURL)
	if err != nil {
		delete_user(uid)
		ctx.String(http.StatusInternalServerError, err.Error())
		return
	}
	// 授業の時限を設定
	datainit.BuildClassDatas(classes)

	//データベース: classesテーブルに追加
	//下は授業の数だけfor文になる予定です
	for _, s := range classes {
		true_name := true_title(s.Class)
		data = map[string]interface{}{"class_name": true_name, "start": s.Start, "end": s.End, "x": s.X, "y": s.Y, "length": s.Length, "uid": uid}
		_, err := db.NamedExec("INSERT INTO classes (class, start, end, x, y, length, uid) VALUES (:class_name, :start, :end, :x, :y, :length, :uid)", data)
		if err != nil {
			delete_user(uid)
			ctx.String(http.StatusInternalServerError, err.Error())
			return
		}
	}

	//Get user in DB
	query := "SELECT * FROM users WHERE id=" + uid
	err = db.Get(&user, query)
	if err != nil {
		delete_user(uid)
		ctx.String(http.StatusInternalServerError, err.Error())
		return
	}

	//Get class ZOOM url and insert DB
	//classテーブルから重複を許さず授業名をとってくる
	var class_names []string
	for _, class := range classes {
		if NotContain(class_names, class.Class) {
			class_names = append(class_names, class.Class)
		}
	}
	var d_classes []database.Class
	for _, s := range class_names {
		var class database.Class
		s = true_title(s)
		class.Class = s
		d_classes = append(d_classes, class)
	}
	var zoomurls []string
	for _, class := range d_classes {
		zoomurls, err = mail.GetURL(class, user)
		if err != nil {
			delete_user(uid)
			ctx.String(http.StatusInternalServerError, err.Error())
			return
		}
		fmt.Println("Signup():授業", class.Class)
		fmt.Println("Signup():url_num", len(zoomurls))
		if len(zoomurls) != 0 {
			nowtime := time.Now().Format(layout)
			for j := len(zoomurls) - 1; j >= 0; j-- {
				data := map[string]interface{}{"url": zoomurls[j], "uid": uid, "class": class.Class, "now": nowtime}
				if j == len(zoomurls)-1 {
					_, err := db.NamedExec("UPDATE classes SET url=(:url) WHERE uid=(:uid) AND start IN (SELECT MIN(start) FROM (SELECT start FROM classes WHERE class=(:class) AND uid=(:uid) AND start>(:now))tmp)", data)
					if err != nil {
						delete_user(uid)
						ctx.String(http.StatusInternalServerError, err.Error())
						return
					}
				} else {
					_, err := db.NamedExec("UPDATE classes SET url=(:url) WHERE uid=(:uid) AND start IN (SELECT MAX(start) FROM (SELECT start FROM classes WHERE class=(:class) AND uid=(:uid) AND start<(:now))tmp)", data)
					if err != nil {
						delete_user(uid)
						ctx.String(http.StatusInternalServerError, err.Error())
						return
					}
				}
				var lastupdate_class database.Class
				query = "SELECT * FROM classes WHERE url= '" + zoomurls[j] + "'"
				err = db.Get(&lastupdate_class, query)
				if err == nil {
					nowtime = lastupdate_class.Start.Format(layout)
				}
			}
		}
	}

	//Update ログイン時間を更新
	nowtime := time.Now().Format(layout)
	data = map[string]interface{}{"uid": uid, "now": nowtime}
	_, err = db.NamedExec("UPDATE users SET last_time=(:now) WHERE id=(:uid)", data)
	if err != nil {
		delete_user(uid)
		ctx.String(http.StatusInternalServerError, err.Error())
		return
	}

	ctx.SetCookie("id", uid, 1000, "/", "localhost", false, true)
	ctx.Redirect(http.StatusSeeOther, "/main")
}

func AccountEditPage(ctx *gin.Context){
	// Get DB connection
	db, err := database.GetConnection()
	if err != nil {
		ctx.String(http.StatusInternalServerError, err.Error())
		return
	}
	//クッキーからユーザIDを得る
	cookie, err := ctx.Request.Cookie("id")
	uid, _ := url.QueryUnescape(cookie.Value)
	if err != nil {
		ctx.String(http.StatusOK, "cookie is nil")
		return
	}
	var user database.User
	query := "SELECT * FROM users WHERE id=" + uid
	err = db.Get(&user, query)
	if err != nil {
		ctx.String(http.StatusInternalServerError, err.Error())
		return
	}
	ctx.HTML(http.StatusOK, "account.html", gin.H{"Title": "ACCOUNT", "User": user})
}

func AccountEdit(ctx *gin.Context) {
	// Get DB connection
	db, err := database.GetConnection()
	if err != nil {
		ctx.String(http.StatusInternalServerError, err.Error())
		return
	}

	//クッキーからユーザIDを得る
	cookie, err := ctx.Request.Cookie("id")
	uid, _ := url.QueryUnescape(cookie.Value)
	if err != nil {
		ctx.String(http.StatusOK, "cookie is nil")
		return
	}

	//user name change
	name, _ := ctx.GetPostForm("user")
	if name != "" {
		data_name := map[string]interface{}{"name": name, "uid": uid}
		_, err := db.NamedExec("UPDATE users SET name=(:name) WHERE id=(:uid)", data_name)
		if err != nil {
			ctx.String(http.StatusInternalServerError, err.Error())
			return
		}
	}

	//email_id change
	email_id, _ := ctx.GetPostForm("email_id")
	if email_id != "" {
		data_email := map[string]interface{}{"email_id": email_id, "uid": uid}
		_, err := db.NamedExec("UPDATE users SET email_id=(:email_id) WHERE id=(:uid)", data_email)
		if err != nil {
			ctx.String(http.StatusInternalServerError, err.Error())
			return
		}
	}

	//acount pw change
	new_pw, _ := ctx.GetPostForm("pw")
	if new_pw != "" {
		new_hpw := fmt.Sprintf("%x", sha256.Sum256([]byte(new_pw)))
		data_pw := map[string]interface{}{"new_hpw": new_hpw, "uid": uid}
		_, err := db.NamedExec("UPDATE users SET password=(:new_hpw) WHERE id=(:uid)", data_pw)
		if err != nil {
			ctx.String(http.StatusInternalServerError, err.Error())
			return
		}
	}

	//iカレンダー更新
	//iカレンダーのURLを受け取る
	icalURL, _ := ctx.GetPostForm("iurl")
	if icalURL != "" {
		var classes []database.Class
		classes, err = ical.GetCalData(icalURL)
		if err != nil {
			ctx.String(http.StatusInternalServerError, err.Error())
			return
		}
		//データベース: classesテーブルに追加
		datainit.BuildClassDatas(classes)
		for _, s := range classes {
			true_name := true_title(s.Class)
			data := map[string]interface{}{"class_name": true_name, "start": s.Start, "end": s.End, "x": s.X, "y": s.Y, "length": s.Length, "uid": uid}
			_, err := db.NamedExec("INSERT INTO classes (class, start, end, x, y, length, uid) VALUES (:class_name, :start, :end, :x, :y, :length, :uid)", data)
			if err != nil {
				ctx.String(http.StatusInternalServerError, err.Error())
				return
			}
		}
	}
	ctx.Redirect(http.StatusSeeOther, "/account")
}

func Logout(ctx *gin.Context) {
	ctx.SetCookie("id", "", -1, "/", "localhost", false, true)
	ctx.HTML(http.StatusOK, "index.html", gin.H{"Title": "HOME"})
}

//退会
func Cancel(ctx *gin.Context) {
	// Get DB connection
	db, err := database.GetConnection()
	if err != nil {
		ctx.String(http.StatusInternalServerError, err.Error())
		return
	}

	//クッキーからユーザIDを得る
	cookie, err := ctx.Request.Cookie("id")
	uid, _ := url.QueryUnescape(cookie.Value)
	if err != nil {
		ctx.String(http.StatusOK, "cookie is nil")
		return
	}

	data := map[string]interface{}{"uid": uid}
	_, new_err := db.NamedExec("DELETE FROM users WHERE id=(:uid)", data)
	if new_err != nil {
		ctx.String(http.StatusInternalServerError, err.Error())
		return
	}

	ctx.SetCookie("id", "", -1, "/", "localhost", false, true)
	ctx.HTML(http.StatusOK, "login.html", gin.H{"Title": "LOGIN"})
}

func NotContain(texts []string, s string) bool {
	for _, text := range texts {
		if text == s {
			return false
		}
	}
	return true
}

func true_title(title string) string {
	titles := strings.Split(title, "【")
	return titles[0]
}

func delete_user(uid string) {
	// Get DB connection
	db, _ := database.GetConnection()
	data := map[string]interface{}{"uid": uid}
	_, _ = db.NamedExec("DELETE FROM users WHERE id=(:uid)", data)
}