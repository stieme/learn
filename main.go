package main

import (
	"database/sql"
	"errors"
	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
	"log"
	"net/http"
)

type User struct {
	Username string
	Password string
}

func main() {
	// 数据库连接,打开text数数据库
	db, err := sql.Open("mysql", "root:as556564996@tcp(127.0.0.1:3306)/text")
	if err != nil {
		log.Fatal("数据库链接出错: ", err) //没链接到就打印错误信息
	}

	err = db.Ping()
	if err != nil {
		log.Fatal("数据库未仍在链接: ", err)
	}
	//检测数据库链接是否任然有效，否则打印错误信息

	defer func(db *sql.DB) {
		err := db.Close()
		if err != nil {
			log.Println("Error closing database connection:", err)
		}
	}(db) //延迟函数。断开数据库链接

	//创建路由引擎
	r := gin.Default()
	r.LoadHTMLGlob("static/*") //模板引擎

	//设置静态文件服务器，客户端发起以/static开头的url请求时，Gin从./static目录查找相应文件
	r.Static("/static", "./static")

	//设置路由处理器
	r.POST("/login", func(c *gin.Context) {
		var user User
		username := c.PostForm("username")
		password := c.PostForm("password")

		err := db.QueryRow("SELECT username, password FROM users WHERE username = ?", username).Scan(&user.Username, &user.Password)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				c.HTML(http.StatusOK, "error.html", gin.H{"error": "用户不存在"})
			} else {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "数据库查询失败"})
			}
			return
		}
		if user.Password != password {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "密码错误"})
			return
		}

		c.Redirect(http.StatusFound, "/static/blog.html")
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
		err := db.QueryRow("SELECT username FROM users WHERE username = ?", username).Scan(&user.Username)
		if err == nil {
			c.JSON(http.StatusConflict, gin.H{"error": "用户已存在"})
			return
		}
		if !errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "数据库查询失败"})
			return
		}

		_, err = db.Exec("INSERT INTO users (username, password) VALUES (?, ?)", username, password)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "注册失败"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "注册成功"})
	})

	r.POST("/reset-password", func(c *gin.Context) {
		username := c.PostForm("username")
		password := c.PostForm("newPassword")
		confirmPassword := c.PostForm("confirmPassword")

		if password != confirmPassword {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "密码不一致"})
			return
		}

		var user User
		err := db.QueryRow("SELECT username FROM users WHERE username = ?", username).Scan(&user.Username)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				c.JSON(http.StatusNotFound, gin.H{"error": "用户不存在"})
			} else {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "数据库错误"})
			}
			return
		}

		_, err = db.Exec("UPDATE users SET password = ? WHERE username = ?", password, username)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "更新密码失败"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "密码重置成功"})

	})

	//启动服务器
	log.Println("Server starting on port 8080...")
	err = r.Run(":8080")
	if err != nil {
		return
	}
}
