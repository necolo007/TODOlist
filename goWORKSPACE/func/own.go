package _func

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"net/http"
)

type UserInfo struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Id       string `json:"id"`
} //打了tag的结构体
func getu(c *gin.Context) {
	var u UserInfo
	err := c.ShouldBindWith(&u, binding.JSON)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
	} else {
		c.JSON(http.StatusOK, gin.H{
			"status":   "ok",
			"username": u.Username,
			"id":       u.Id,
		})
	}
} //绑定数据结构
func Sayhello() {
	fmt.Println("hello world")
} //say hello
