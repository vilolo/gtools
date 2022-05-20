package utils

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"../structs"

	_ "github.com/go-sql-driver/mysql"
)

//mysql
// const (
// 	USERNAME = "root"
// 	PASSWORD = "root"
// 	NETWORK  = "tcp"
// 	SERVER   = "localhost"
// 	PORT     = 3306
// 	DATABASE = "test"
// )

func GetDB() *sql.DB {
	file, _ := os.Open("conf.json")
	defer file.Close()
	decoder := json.NewDecoder(file)
	conf := structs.Conf{}
	err := decoder.Decode(&conf)
	if err != nil {
		fmt.Println("Error:", err)
	}

	dsn := fmt.Sprintf("%s:%s@%s(%s:%d)/%s", conf.Db_username, conf.Db_pwd, conf.Db_network, conf.Db_server, conf.Db_port, conf.Db_database)
	DB, err := sql.Open("mysql", dsn)
	if err != nil {
		fmt.Printf("Open mysql failed,err:%v\n", err)
		os.Exit(0)
	}
	DB.SetConnMaxLifetime(100 * time.Second) //最大连接周期，超过时间的连接就close
	DB.SetMaxOpenConns(100)                  //设置最大连接数
	DB.SetMaxIdleConns(16)                   //设置闲置连接数
	return DB
}
