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
	"sync"
	"time"
)

var DB *gorm.DB
var UserStore sync.Map

// 链接数据库
func init() {
	dsn := "root:Cc530357154@@tcp(127.0.0.1:3306)/test?charset=utf8mb4&parseTime=True&loc=Local"
	var err error
	DB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{
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
	age          uint `gorm:"column:age"`
}

// Post 帖子信息
type Post struct {
	ID         uint64 `gorm:"primary_key:auto_increment"`
	Title      string
	Content    string
	AuthorID   uint64
	CreatTime  time.Time
	UpdateTime time.Time
	Liking     int `gorm:"default:0"`
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
		username := c.Param("username")
		value, err := UserStore.Load(username)
		if !err {
			c.JSON(200, gin.H{"error": "读取用户信息失败！"})
		}
		user := value.(UserInfo)
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
	username := c.Param("username")
	if err := c.BindJSON(&post); err != nil {
		c.JSON(200, gin.H{
			"error": "无效请求",
		})
	} else {
		value, ok := UserStore.Load(username)
		if !ok {
			c.JSON(200, gin.H{"error": "未读取到用户信息！"})
		}
		post.AuthorID = value.(UserInfo).ID
		post.CreatTime = time.Now()
		post.UpdateTime = time.Now()
		DB.Create(&post)
		c.JSON(200, post)
	}
}

// ReplyPost 回帖
func ReplyPost(c *gin.Context) {
	postID := c.Param("postID")
	username := c.Param("username")
	var comment Comment
	if err := c.BindJSON(&comment); err != nil {
		c.JSON(200, gin.H{
			"err": "无效的请求",
		})
	} else {
		value, ok := UserStore.Load(username)
		if !ok {
			c.JSON(200, gin.H{"error": "未读取到用户信息！"})
		}
		if value == nil {
			c.JSON(200, gin.H{"error": "用户信息为空！"})
		}
		comment.AuthorID = value.(UserInfo).ID
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
	updatePost.UpdateTime = time.Now()
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
		SortByLike := c.Query("SortByLikes")
		if SortByLike == "true" {
			DB = DB.Order("likes desc")
		}
		DB.Find(&posts)
	}
	c.JSON(200, posts)
}

// Login 登录
func Login(c *gin.Context) {
	var user UserInfo
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
		//对比加密后的密码，识别进入
		err := bcrypt.CompareHashAndPassword([]byte(ExistingUser.Password), []byte(user.Password))
		if err != nil {
			c.JSON(200, gin.H{
				"msg": "登录失败，用户名或者密码错误!",
			})
			return
		} else {
			UserStore.Store(user.Name, ExistingUser)
			c.JSON(http.StatusOK, gin.H{
				"msg": "登录成功！",
			})
		}
	}
}

// Register 注册
func Register(c *gin.Context) {
	var user UserInfo
_:
	c.BindJSON(&user)
	res := DB.Where("name=?", user.Name).First(&user)
	if res.RowsAffected != 0 {
		c.JSON(http.StatusOK, gin.H{
			"msg": "注册失败，用户名字已经存在!",
		})
	} else {
		//对密码进行加密储存
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
}

// Like 点赞帖子
func Like(c *gin.Context) {
	postID := c.Param("postID")
	var post Post
	result := DB.First(&post, postID)
	if result.RowsAffected == 0 {
		c.JSON(http.StatusOK, gin.H{"error": "帖子不存在！"})
	} else {
		post.Liking += 1
		DB.Save(&post)
		c.JSON(http.StatusOK, gin.H{"msg": "点赞成功！"})
	}
}

func main() {
	r := gin.Default()
	//登录与注册
	v := r.Group("/get")
	{
		//注册功能
		v.POST("/register", Register)
		//登录功能
		v.POST("/login", Login)
	}
	//帖子功能实现
	u := r.Group("/posts")
	{
		//发帖功能
		u.POST("/upload/:username", ViolationCheck(), CreatPost)
		//搜帖功能
		u.GET("/search", SearchPost)
		post := u.Group("/:postID")
		{
			//回复贴子
			post.POST("/comments/:username", ReplyPost)
			//更新帖子
			post.PUT("/update", UpdatePost)
			//删除帖子
			post.DELETE("/delete", DeletePost)
			//点赞帖子
			post.POST("/like", Like)
		}
	}
	err := r.Run("localhost:8080")
	if err != nil {
		fmt.Println("服务器启动失败！")
	}
}
