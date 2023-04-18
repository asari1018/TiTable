package main

import (
	"fmt"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"

	"titable.go/db"
	"titable.go/service"
)

const port = 8000

func main() {

	// initialize DB connection
	dsn := db.DefaultDSN(
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"))
	if err := db.Connect(dsn); err != nil {
		log.Fatal(err)
	}

	// initialize Gin engine
	engine := gin.Default()
	engine.LoadHTMLGlob("views/*.html")


	// routing
	engine.Static("/assets", "./assets")
	engine.GET("/", service.Home)
	engine.GET("/login", service.LoginPage)
	engine.POST("/login", service.Login)
	engine.GET("/signup", service.SignupPage)
	engine.POST("/signup", service.Signup)
	engine.GET("/main", service.Main)
	engine.GET("/main/:date", service.Main)
	engine.GET("/class/:class", service.Class)
	engine.GET("/NewTask", service.TaskInsert)
	engine.POST("/taskinsertedit", service.TaskInsertEdit)
	engine.POST("/NewComment", service.ClassEdit)
	engine.GET("/task/:task", service.Task)
	engine.POST("/taskedit", service.TaskEdit)
	engine.GET("/account", service.AccountEditPage)
	engine.POST("/accountedit", service.AccountEdit)
	engine.GET("/cancel", service.Cancel)
	engine.GET("/logout", service.Logout)
	engine.POST("/done", service.TaskDone)
	engine.POST("/undone", service.TaskUnDone)
	engine.POST("/taskdelete", service.TaskDelete)

	// start server
	engine.Run(fmt.Sprintf(":%d", port))
}
