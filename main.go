package main

import (
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"log"
	"net/http"
)

func main() {
	fs := http.FileServer(http.Dir("static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	http.HandleFunc("/blog", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./static/blog.html")
	})

	// 替换为您的数据库信息
	db, err := sql.Open("mysql", "root:password@tcp(127.0.0.1:3306)/dbname")
	checkErr(err)
	err = db.Ping()
	checkErr(err)
	defer db.Close()

	http.ListenAndServe(":8080", nil)
}

func checkErr(err error) {
	if err != nil {
		log.Fatal("Error: ", err)
	}
}
