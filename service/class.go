package service

import (
	"net/http"

	"net/url"

	"github.com/gin-gonic/gin"
	database "titable.go/db"
	"fmt"
)

func commentIsNull(class database.Class) bool {
	return (class.Comment != "hoge")
}

func urlIsNull(class database.Class) bool {
	return (class.URL != "hoge")
} 

//授業編集画面

func Class(ctx *gin.Context) {
	// Get DB connection
	db, err := database.GetConnection()
	if err != nil {
		ctx.String(http.StatusInternalServerError, err.Error())
		return
	}

	// parse class id given as a parameter
	class_id := ctx.Param("class")

	//クッキーからユーザIDを得る
	cookie, err := ctx.Request.Cookie("id")
	uid, _ := url.QueryUnescape(cookie.Value)
	if err != nil {
		ctx.String(http.StatusOK, "cookie is nil")
		return
	}

	// Get class in DB
	var class database.Class
	query := "SELECT * FROM classes WHERE id =" + class_id
	err = db.Get(&class, query)
	if err != nil {
		ctx.String(http.StatusInternalServerError, err.Error())
		return
	}


	// Get tasks in DB
	var tasks []database.Task
	query = "SELECT * FROM tasks WHERE class = '" + class.Class +"' AND id IN (SELECT task_id FROM user_info WHERE user_id=" + uid +")"
	err = db.Select(&tasks, query)
	if err != nil {
		ctx.String(http.StatusInternalServerError, err.Error())
		return
	} 

	fmt.Println(tasks)

	ctx.SetCookie("class", class_id, 1000, "/", "localhost", false, true)

	commentFlag := commentIsNull(class)
	urlFlag := urlIsNull(class)

	// Render tasks
	ctx.HTML(http.StatusOK, "class.html", gin.H{"Title": "Class", "Tasks": tasks, "Class": class, "CommentFlag": commentFlag,"URLFlag": urlFlag})

}

func ClassEdit(ctx *gin.Context) {
	// Get DB connection
	db, err := database.GetConnection()
	if err != nil {
		ctx.String(http.StatusInternalServerError, err.Error())
		return
	}

	//クッキーからクラス名を得る
	cookie, err := ctx.Request.Cookie("class")
	class_id, _ := url.QueryUnescape(cookie.Value)
	if err != nil {
		ctx.String(http.StatusOK, "cookie is nil")
		return
	}

	// Get class in DB
	var class database.Class
	query := "SELECT * FROM classes WHERE id =" + class_id
	err = db.Get(&class, query)
	if err != nil {
		ctx.String(http.StatusInternalServerError, err.Error())
		return
	}

	//comment change
	comment, _exists := ctx.GetPostForm("detail")
	if _exists {
		data_comment := map[string]interface{}{"class": class.Class, "comment": comment}
		_, err := db.NamedExec("UPDATE classes SET comment=(:comment) WHERE class = (:class)", data_comment)
		if err != nil {
			ctx.String(http.StatusInternalServerError, err.Error())
			return
		}
	}

	// Render tasks
	ctx.Redirect(http.StatusSeeOther, "/class/"+class_id)
}

func TaskInsert(ctx *gin.Context) {
	// Get DB connection
	db, err := database.GetConnection()
	if err != nil {
		ctx.String(http.StatusInternalServerError, err.Error())
		return
	}
	//クッキーからクラスIDを得る
	cookie, err := ctx.Request.Cookie("class")
	class_id, _ := url.QueryUnescape(cookie.Value)
	if err != nil {
		ctx.String(http.StatusOK, "cookie is nil")
		return
	}
	// Get class in DB
	var class database.Class
	query := "SELECT * FROM classes WHERE id =" + class_id
	err = db.Get(&class, query)
	if err != nil {
		ctx.String(http.StatusInternalServerError, err.Error())
		return
	}
	ctx.HTML(http.StatusOK, "taskinsert.html", gin.H{"Title": "TaskInsert", "ClassName": class.Class})
}

func TaskInsertEdit(ctx *gin.Context) {
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

	//クッキーから授業IDを得る
	cookie, err = ctx.Request.Cookie("class")
	class_id, _ := url.QueryUnescape(cookie.Value)
	if err != nil {
		ctx.String(http.StatusOK, "cookie is nil")
		return
	}

	// Get class in DB
	var class database.Class
	query := "SELECT * FROM classes WHERE id =" + class_id
	err = db.Get(&class, query)
	if err != nil {
		ctx.String(http.StatusInternalServerError, err.Error())
		return
	}

	//task Insert
	task_name, _ := ctx.GetPostForm("title")
	level, _ := ctx.GetPostForm("level")
	deadline_o, _ := ctx.GetPostForm("deadline")

	deadline := DatetimeCast(deadline_o)
	data := map[string]interface{}{"title": task_name, "class": class.Class, "level": level, "deadline": deadline}
	res, err := db.NamedExec("INSERT INTO tasks (title, class, task_level, deadline_at) VALUES (:title, :class, :level, :deadline)", data)
	if err != nil {
		ctx.String(http.StatusInternalServerError, err.Error())
		return
	}
	task_id, _ := res.LastInsertId()
	data = map[string]interface{}{"task_id": task_id, "uid": uid}
	_, err = db.NamedExec("INSERT INTO user_info (task_id, user_id) VALUES (:task_id, :uid)", data)
	if err != nil {
		ctx.String(http.StatusInternalServerError, err.Error())
		return
	}

	link := "/class/" + class_id
	//授業詳細画面に戻る
	ctx.Redirect(http.StatusSeeOther, link)
}
