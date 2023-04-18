package service

import (
	"net/http"

	"net/url"

	"github.com/gin-gonic/gin"
	database "titable.go/db"
	"fmt"
)

//タスク詳細画面

func Task(ctx *gin.Context) {
	// Get DB connection
	db, err := database.GetConnection()
	if err != nil {
		ctx.String(http.StatusInternalServerError, err.Error())
		return
	}

	// parse class name given as a parameter
	task_id := ctx.Param("task")

	// Get tasks in DB
	var task database.Task
	query := "SELECT * FROM tasks WHERE id = " + task_id
	err = db.Get(&task, query)
	if err != nil {
		ctx.String(http.StatusInternalServerError, err.Error())
		return
	}

	ctx.SetCookie("task", task_id, 1000, "/", "localhost", false, true)

	// Render tasks
	ctx.HTML(http.StatusOK, "task.html", gin.H{"Title": "TASK", "Task": task})

}

func TaskEdit(ctx *gin.Context) {
	// Get DB connection
	db, err := database.GetConnection()
	if err != nil {
		ctx.String(http.StatusInternalServerError, err.Error())
		return
	}

	//クッキーからタスクidを得る
	cookie, err := ctx.Request.Cookie("task")
	task_id, _ := url.QueryUnescape(cookie.Value)
	if err != nil {
		ctx.String(http.StatusOK, "cookie is nil")
		return
	}

	//task title change
	title, _ := ctx.GetPostForm("title")
	if title != "" {
		data_title := map[string]interface{}{"task_id": task_id, "new_title": title}
		_, err := db.NamedExec("UPDATE tasks SET title=(:new_title) WHERE id=(:task_id)", data_title)
		if err != nil {
			ctx.String(http.StatusInternalServerError, err.Error())
			return
		}
	}

	// task level change
	level, _ := ctx.GetPostForm("level")
	if level != "" {
		data_title := map[string]interface{}{"task_id": task_id, "level": level}
		_, err := db.NamedExec("UPDATE tasks SET task_level=(:level) WHERE id=(:task_id)", data_title)
		if err != nil {
			ctx.String(http.StatusInternalServerError, err.Error())
			return
		}
	}

	// deadline_at level change
	deadline_o, _ := ctx.GetPostForm("deadline")
	if deadline_o != "" {
		deadline := DatetimeCast(deadline_o)
		data_title := map[string]interface{}{"task_id": task_id, "deadline": deadline}
		_, err := db.NamedExec("UPDATE tasks SET deadline_at=(:deadline) WHERE id=(:task_id)", data_title)
		if err != nil {
			ctx.String(http.StatusInternalServerError, err.Error())
			return
		}
	}
	//タスク詳細画面に戻る
	ctx.Redirect(http.StatusSeeOther, "/task/"+task_id)
}

func TaskDone(ctx *gin.Context) {
	// Get DB connection
	db, err := database.GetConnection()
	if err != nil {
		ctx.String(http.StatusInternalServerError, err.Error())
		return
	}

	//クッキーからタスクidを得る
	cookie, err := ctx.Request.Cookie("task")
	task_id, _ := url.QueryUnescape(cookie.Value)
	if err != nil {
		ctx.String(http.StatusOK, "cookie is nil")
		return
	}

	//done change
	data_title := map[string]interface{}{"task_id": task_id}
	_, err = db.NamedExec("UPDATE tasks SET is_done=true WHERE id=(:task_id)", data_title)
	if err != nil {
		ctx.String(http.StatusInternalServerError, err.Error())
		return
	}
	fmt.Println("たくすあいで",task_id)

	// Render tasks
	ctx.Redirect(http.StatusSeeOther, "/task/"+task_id)
}

func TaskUnDone(ctx *gin.Context) {
	// Get DB connection
	db, err := database.GetConnection()
	if err != nil {
		ctx.String(http.StatusInternalServerError, err.Error())
		return
	}

	//クッキーからタスクidを得る
	cookie, err := ctx.Request.Cookie("task")
	task_id, _ := url.QueryUnescape(cookie.Value)
	if err != nil {
		ctx.String(http.StatusOK, "cookie is nil")
		return
	}
	//undone change
	data_title := map[string]interface{}{"task_id": task_id}
	_, err = db.NamedExec("UPDATE tasks SET is_done=false WHERE id=(:task_id)", data_title)
	if err != nil {
		ctx.String(http.StatusInternalServerError, err.Error())
		return
	}

	// Render tasks
	ctx.Redirect(http.StatusSeeOther, "/task/"+task_id)
}

func TaskDelete(ctx *gin.Context) {
	// Get DB connection
	db, err := database.GetConnection()
	if err != nil {
		ctx.String(http.StatusInternalServerError, err.Error())
		return
	}

	//クッキーからタスクidを得る
	cookie, err := ctx.Request.Cookie("task")
	task_id, _ := url.QueryUnescape(cookie.Value)
	if err != nil {
		ctx.String(http.StatusOK, "cookie is nil")
		return
	}
	//taskdelete
	data_title := map[string]interface{}{"task_id": task_id}
	_, err = db.NamedExec("DELETE FROM tasks WHERE id=(:task_id)", data_title)
	if err != nil {
		ctx.String(http.StatusInternalServerError, err.Error())
		return
	}
	// Render tasks
	ctx.Redirect(http.StatusSeeOther, "/task/"+task_id)
}

//datetime-local(html) -> datetime(db)
func DatetimeCast(s string) string {
	s1 := s[0:10]
	s2 := s[11:16]
	s3 := s1 + " " + s2 + ":00.000"
	return s3
}
