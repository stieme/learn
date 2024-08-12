package main

import (
	"errors"
	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"log"
	"net/http"
)

// User struct 定义模型
type User struct {
	gorm.Model
	Username string
	Password string
}

func main() {
	// 数据库连接
	dsn := "root:as556564996@tcp(127.0.0.1:3306)/text?charset=utf8mb4&parseTime=True&loc=Local"
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("database connection failed: ", err)
	}

	// 迁移 User 表
	db.AutoMigrate(&User{})

	r := gin.Default()
	r.LoadHTMLGlob("static/*")      // 模板引擎
	r.Static("/static", "./static") // 设置静态文件服务器

	// 登录路由处理器
	r.POST("/login", func(c *gin.Context) {
		var user User
		username := c.PostForm("username")
		password := c.PostForm("password")

		// 使用 GORM 查询用户
		if err := db.Where("username = ?", username).First(&user).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				c.HTML(http.StatusOK, "error.html", gin.H{"error": "用户不存在"})
			} else {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "数据库查询失败"})
			}
			return
		}

		// 检查密码是否匹配
		if user.Password != password {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "密码错误"})
		} else {
			c.Redirect(http.StatusFound, "/static/blog.html")
		}
	})

	r.POST("/register", func(c *gin.Context) {
		username := c.PostForm("username")
		password := c.PostForm("password")
		confirmPassword := c.PostForm("confirm_password")

		if password != confirmPassword {
			c.JSON(http.StatusBadRequest, gin.H{"error": "两次输入的密码不一致"})
			return
		}

		var user User
		// 检查用户名是否已存在
		if err := db.Where("username = ?", username).First(&user).Error; err == nil {
			c.JSON(http.StatusConflict, gin.H{"error": "用户已存在"})
			return
		}

		// 创建新用户
		if err := db.Create(&User{Username: username, Password: password}).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "注册失败"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "注册成功"})
	})

	// 重置密码路由处理器
	r.POST("/reset-password", func(c *gin.Context) {
		username := c.PostForm("username")
		newPassword := c.PostForm("newPassword")
		confirmPassword := c.PostForm("confirmPassword")

		if newPassword != confirmPassword {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "密码不一致"})
			return
		}

		var user User
		if err := db.Where("username = ?", username).First(&user).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				c.JSON(http.StatusNotFound, gin.H{"error": "用户不存在"})
			} else {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "数据库错误"})
			}
			return
		}

		// 更新密码
		if err := db.Model(&user).Updates(map[string]interface{}{"password": newPassword}).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "更新密码失败"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "密码重置成功"})
	})

	// 启动服务器
	log.Println("Server starting on port 8080...")
	if err := r.Run(":8080"); err != nil {
		log.Fatal("server run failed: ", err)
	}
}
