package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"log"
	"net/http"
	"time"
)

var DB *gorm.DB

func init() {
	dsn := "root:Cc530357154@@tcp(127.0.0.1:3306)/test?charset=utf8mb4&parseTime=True&loc=Local"
	DB, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		log.Fatalf("无法链接到数据库%v", err.Error())
	}
	err = DB.AutoMigrate(&UserInfo{}, &Post{}, &Comment{})
	if err != nil {
		log.Printf("自动迁移数据表失败%v", err.Error())
	}
}

// UserInfo 用户信息
type UserInfo struct {
	ID           uint64 `gorm:"primary_key:AUTO_INCREMENT"`
	Name         string
	Password     string
	Email        string
	Gender       string
	Mistake      uint `gorm:"default:0"`
	Ban          uint `gorm:"default:0"`
	BanStartTime *time.Time
	BanDuration  int
	Birthday     *time.Time
	age          uint
}

// Post 帖子信息
type Post struct {
	ID         uint64 `gorm:"primary_key:auto_increment"`
	Title      string
	Content    string
	AuthorID   uint64
	CreatTime  time.Time
	UpdateTime time.Time
}

// Comment Comment信息
type Comment struct {
	ID        uint64 `gorm:"primary_key:auto_increment"`
	PostID    string
	Content   string
	AuthorID  uint64
	CreatTime time.Time
}

// BanUser 封号
func BanUser(x *UserInfo) {
	x.Ban = 1
	now := time.Now()
	x.BanStartTime = &now
	duration := time.Hour * 24 * 5
	x.BanDuration = int(duration.Seconds())
	DB.Save(x)
}

// ViolationCheck 检测用户违规的中间件*****
func ViolationCheck() gin.HandlerFunc {
	return func(c *gin.Context) {
		user := UserInfo{}
	_:
		c.BindJSON(&user)
		if user.Ban == 1 {
			elapsed := time.Since(*user.BanStartTime)
			remaining := user.BanDuration - int(elapsed)
			if remaining > 0 {
				c.JSON(200, gin.H{
					"msg":     "账号正在封禁中",
					"seconds": remaining,
				})
				c.Abort()
			} else {
				user.Ban = 0
				user.BanStartTime = nil
				user.BanDuration = 0
				DB.Save(&user)
				c.Next()
			}
		} else {
			c.Next()
		}
	}
}

// CreatPost 发帖
func CreatPost(c *gin.Context) {
	var post Post
	if err := c.BindJSON(&post); err != nil {
		c.JSON(200, gin.H{
			"error": "无效请求",
		})
	} else {
		post.CreatTime = time.Now()
		post.UpdateTime = time.Now()
		DB.Create(&post)
		c.JSON(200, post)
	}
}

// ReplyPost 回帖
func ReplyPost(c *gin.Context) {
	postID := c.Param("postID")
	var comment Comment
	if err := c.BindJSON(&comment); err != nil {
		c.JSON(200, gin.H{
			"err": "无效的请求",
		})
	} else {
		comment.PostID = postID
		comment.CreatTime = time.Now()
		DB.Create(&comment)
		c.JSON(200, comment)
	}
}

// DeletePost 删帖
func DeletePost(c *gin.Context) {
	postID := c.Param("postID")
	result := DB.Delete(&Post{}, postID)
	if result.RowsAffected == 0 {
		c.JSON(200, gin.H{"error": "帖子不存在！"})
	} else if result.Error != nil {
		c.JSON(200, gin.H{"error": "删除帖子失败！"})
	} else {
		c.JSON(200, gin.H{"msg": "删除帖子成功！"})
	}
}

// UpdatePost 更新帖子
func UpdatePost(c *gin.Context) {
	PostID := c.Param("postID")
	var updatePost Post
	if err := c.BindJSON(&updatePost); err != nil {
		c.JSON(200, gin.H{"error": "无效的请求数据！"})
	}
	result := DB.Model(&Post{}).Where("id = ?", PostID).Updates(updatePost)
	if result.RowsAffected == 0 {
		c.JSON(200, gin.H{"error": "帖子不存在!"})
	} else if result.Error != nil {
		c.JSON(200, gin.H{"error": "更新失败！"})
	} else {
		c.JSON(200, gin.H{"msg": "更新成功！"})
	}
}

// SearchPost 查帖子*****
func SearchPost(c *gin.Context) {
	query := c.Query("query")
	var posts []Post
	if query != "" {
		DB.Where("Title LINK OR content LIKE ?", "%"+query+"%", "%"+query+"%").Find(&posts)
	} else {
		DB.Find(&posts)
	}
	c.JSON(200, posts)
}

// ShowPost 帖子展示
func ShowPost(c *gin.Context) {
	PostID := c.Param("postID")
	var post Post
	result := DB.Preload("Comments").First(&post, PostID)
	if result.RowsAffected == 0 {
		c.JSON(200, gin.H{"error": "帖子不存在！"})
	} else {
		c.JSON(200, post)
	}
}
func main() {
	dsn := "root:Cc530357154@@tcp(127.0.0.1:3306)/test?charset=utf8mb4&parseTime=True&loc=Local"
	DB, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		log.Fatalf("无法链接到数据库%v", err.Error())
	}
	err = DB.AutoMigrate(&UserInfo{}, &Post{}, &Comment{})
	if err != nil {
		log.Printf("自动迁移数据表失败%v", err.Error())
	}
	r := gin.Default()
	//登录与注册
	v := r.Group("/get")
	{
		user := UserInfo{}
		v.POST("/register", func(c *gin.Context) {
		_:
			c.BindJSON(&user)
			res := DB.Where("name=?", user.Name).First(&user)
			if res.RowsAffected != 0 {
				c.JSON(http.StatusOK, gin.H{
					"msg": "注册失败，用户名字已经存在!",
				})
			} else {
				Hashedpassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
				if err != nil {
					c.JSON(200, gin.H{
						"msg": "密码加密错误",
					})
				}
				user.Password = string(Hashedpassword)
				DB.Create(&user)
				c.JSON(http.StatusOK, gin.H{
					"msg": "注册成功！",
				})
			}
		})
		v.POST("/login", ViolationCheck(), func(c *gin.Context) {
		_:
			c.BindJSON(&user)
			var ExistingUser UserInfo
			res := DB.Where("name=?", user.Name).First(&ExistingUser)
			if res.RowsAffected == 0 {
				c.JSON(http.StatusOK, gin.H{
					"msg": "登录失败，用户名不存在！",
				})
				return
			} else {
				err := bcrypt.CompareHashAndPassword([]byte(ExistingUser.Password), []byte(user.Password))
				if err != nil {
					c.JSON(200, gin.H{
						"msg": "登录失败，用户名或者密码错误!",
					})
					return
				}
				c.JSON(http.StatusOK, gin.H{
					"msg": "登录成功！",
				})
			}
		})
	}
	u := r.Group("/posts")
	{
		u.POST("/upload", func(c *gin.Context) {
			var post Post
			if err := c.BindJSON(&post); err != nil {
				c.JSON(200, gin.H{
					"error": "无效请求",
				})
			} else {
				post.CreatTime = time.Now()
				post.UpdateTime = time.Now()
				DB.Create(&post)
				c.JSON(200, post)
			}
		})
		u.GET("/search", SearchPost)
		post := r.Group("/postID")
		{
			post.POST("/comments", ReplyPost)
			post.PUT("", UpdatePost)
			post.DELETE("", DeletePost)
			post.GET("", ShowPost)
		}
	}
	err = r.Run("localhost:8080")
	if err != nil {
		fmt.Println("服务器启动失败！")
	}
}
