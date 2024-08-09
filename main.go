package main

import (
	"database/sql"
	"errors"
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
		log.Fatal("Database connection failed: ", err) //没链接到就打印错误信息
	}
	defer db.Close() //延迟函数。断开数据库链接

	err = db.Ping()
	if err != nil {
		log.Fatal("Database ping failed: ", err)
	}
	//检测数据库链接是否任然有效，否则打印错误信息

	// 设置静态文件服务器
	fs := http.FileServer(http.Dir("static")) //http.Dir("static")指定了文件系统的根目录
	http.Handle("/static/", http.StripPrefix("/static/", fs))
	//创建了一个新的HTTP处理器，fs可以提供每一个static里的文件

	// 登录路由
	http.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost { //检测请求方法，如果为post方法
			r.ParseForm()                       //解析表单
			username := r.FormValue("username") //赋值
			password := r.FormValue("password") //赋值

			// 从数据库验证用户
			var user User
			err := db.QueryRow("SELECT username, password FROM users WHERE username = ?", username).Scan(&user.Username, &user.Password)
			//QueryRow：这个方法执行一个查询，只返回结果集的第一行。如果查询没有返回任何行，它会返回 sql.ErrNoRows 错误，适用于预期只有一行结果的情况，如根据唯一标识符（如用户名）查询。
			//检测数据库中有没有用户名和username一致，如果有，将用户名和对应密码赋值给user

			//处理错误
			if err != nil {
				if errors.Is(err, sql.ErrNoRows) { //error.Is检测两错误是否相同
					http.Error(w, "User not found.", http.StatusNotFound) //404，没找到
				} else {
					http.Error(w, "Database error.", http.StatusInternalServerError) //500，服务器错误
				}
				return
			}

			// 直接比较明文密码，数据库中是明文，错误就显示无效密码
			if user.Password != password {
				http.Error(w, "Invalid password.", http.StatusUnauthorized)
				return
			}

			// 登录成功，重定向到博客页面
			http.Redirect(w, r, "/static/blog.html", http.StatusFound)
		} else {
			// 如果不是 POST 请求，显示登录表单
			http.ServeFile(w, r, "/static/login.html") // 修改为正确的路径
		}
	})

	// 启动服务器
	log.Println("Server starting on port 8080...")
	http.ListenAndServe(":8080", nil) //监听服务8080 端口
}
