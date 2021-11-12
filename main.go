package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/henrylee2cn/mahonia"

	_ "github.com/go-sql-driver/mysql"

	"./structs"
	"./utils"
)

var dbClient *sql.DB

//mysql
const (
	USERNAME = "root"
	PASSWORD = "root"
	NETWORK  = "tcp"
	SERVER   = "localhost"
	PORT     = 3306
	DATABASE = "test"
)

//hq
const (
	START_DATE = "20210101"
)

func init() {
	dsn := fmt.Sprintf("%s:%s@%s(%s:%d)/%s", USERNAME, PASSWORD, NETWORK, SERVER, PORT, DATABASE)
	DB, err := sql.Open("mysql", dsn)
	if err != nil {
		fmt.Printf("Open mysql failed,err:%v\n", err)
		return
	}
	DB.SetConnMaxLifetime(100 * time.Second) //最大连接周期，超过时间的连接就close
	DB.SetMaxOpenConns(100)                  //设置最大连接数
	DB.SetMaxIdleConns(16)                   //设置闲置连接数
	dbClient = DB
}

func main() {
	fmt.Println("start !!")

	// 获取列表
	// saveList()

	// 获取历史
	updateHistory()

	// 策略
	// 历史验证策略的胜率
	// 展示图片
}

func saveList() {
	// str := "{\"data\":\"sadf\"}"
	// str := `{"data":"sadf"}`
	// var bbb structs.Test
	// json.Unmarshal([]byte(str), &bbb)
	// fmt.Println(bbb)
	// return

	url := "https://98.push2.eastmoney.com/api/qt/clist/get?cb=jQuery112405369178137598498_1636469855448&pn=1&pz=5000&po=1&np=1&ut=bd1d9ddb04089700cf9c27f6f7426281&fltt=2&invt=2&fid=f3&fs=m:0+t:6,m:0+t:80,m:1+t:2,m:1+t:23&fields=f12,f14,f100&_=1636469855466"
	resp, err := utils.GET(url)
	if err != nil {
		fmt.Println("1:" + err.Error())
		return
	}
	jsonStr := resp[42 : len(resp)-2]
	var qtData structs.QtData
	json.Unmarshal([]byte(jsonStr), &qtData)

	for _, v := range qtData.Data.Diff {
		// fmt.Println(k, v)
		_, err := dbClient.Exec("insert into st(name,code,sector)values(?,?,?)", v.Name, v.Code, v.Sector)
		if err != nil {
			fmt.Println("===", err)
		}
	}
}

func testDB() {
	tt := new(structs.DbTest)
	row := dbClient.QueryRow("select id,data from st")
	//row.scan中的字段必须是按照数据库存入字段的顺序，否则报错
	if err := row.Scan(&tt.ID, &tt.Data); err != nil {
		fmt.Printf("scan failed, err:%v", err)
		return
	}
	fmt.Println(*tt)
}

func updateHistory() {
	// rows, err := dbClient.Query("select code from st where updated_at is null or updated_at < ? limit 2", time.Now().Format("2006-01-02"))
	rows, err := dbClient.Query("select code from st where data is null")
	// rows, err := dbClient.Query("select code from st limit 2")
	defer func() {
		if rows != nil {
			rows.Close() //可以关闭掉未scan连接一直占用
		}
	}()
	endDate := time.Now().Format("20060102")
	if err != nil {
		fmt.Printf("Query failed,err:%v", err)
		return
	}
	st := new(structs.DbSt)
	for rows.Next() {
		err = rows.Scan(&st.Code) //不scan会导致连接不释放
		if err != nil {
			fmt.Printf("Scan failed,err:%v", err)
			return
		}

		url := fmt.Sprintf("https://q.stock.sohu.com/hisHq?code=cn_%s&start=%s&end=%s&stat=1&order=D&period=d&callback=historySearchHandler&rt=jsonp", *&st.Code, START_DATE, endDate)
		// fmt.Println(url)

		resp, err := utils.GET(url)
		if err != nil {
			fmt.Println("1:" + err.Error())
			return
		}
		// fmt.Println(resp)

		enc := mahonia.NewDecoder("gbk")
		// fmt.Println("!!!!", enc.ConvertString(resp))

		_, err = dbClient.Exec("update st set data = ? where code = ?", enc.ConvertString(resp), *&st.Code)
		if err != nil {
			fmt.Println("===", err)
		}
	}

}
