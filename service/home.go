package service

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Home renders index.html
func Home(ctx *gin.Context) {
	ctx.HTML(http.StatusOK, "index.html", gin.H{"Title": "HOME"})
}

// Home renders index.html
func LoginPage(ctx *gin.Context) {
	ctx.HTML(http.StatusOK, "login.html", gin.H{"Title": "LOGIN"})
}

func SignupPage(ctx *gin.Context) {
	ctx.HTML(http.StatusOK, "signup.html", gin.H{"Title": "SIGNUP"})
}