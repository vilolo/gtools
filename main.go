package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
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

var pool []structs.QtInfo

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
	// updateHistory()

	// 数据处理
	handleData()

	// 处理结果
	handlePool()

}

func handlePool() {
	if pool != nil {
		jsonStr, _ := json.Marshal(pool)
		// utils.WriteFile("./data/"+time.Now().Format("20060102")+".js","data="+string(jsonStr))

		html := fmt.Sprintf(`<html>
		<head>
			<meta http-equiv="Content-Type" content="text/html; charset=utf-8" />
		</head>
		<div class="info"></div>
		<div class="box">
		</div>
	<script src="https://code.jquery.com/jquery-3.1.1.min.js"></script>
	<script>
		var html = ""
		data = %s
		$(".info").html(data.length)
		data.forEach((element,i) => {
			html += "<span><span>"+(i+1)+"//"+element.f12+"//"+element.f14+"//"+element.f100+"//"+element.f3+"%%</span><br>"+"<img src=\"https://image.sinajs.cn/newchart/daily/n/"+(element.f12[0]==6 ? ("sh"+element.f12) : ("sz"+element.f12))+".gif\"><br>"
		});
		$(".box").html(html)
	</script>
	</html>`, jsonStr)
		utils.WriteFile("./data/"+time.Now().Format("20060102")+".html", html)
	} else {
		fmt.Println("结果为空")
	}
}

var total = 0
var totalUp = 0
var totalDown = 0
var findTotal = 0
var up = 0
var down = 0

func handleData() {
	rows, err := dbClient.Query("select code,name,data,sector,inc_rate from st")
	if err != nil {
		fmt.Println("err1:", err)
		return
	}
	defer func() {
		if rows != nil {
			rows.Close()
		}
	}()
	st := new(structs.DbSt)

	//策略验证开始index
	curI := checkI
	checkDay := checkDay
	var kArr [10]structs.K
	for rows.Next() {
		err = rows.Scan(&st.Code, &st.Name, &st.Data, &st.Sector, &st.IncRate)
		if err != nil {
			fmt.Println("err2:", err)
			return
		}
		// fmt.Println((*&st.Data))
		jsonStr := (*&st.Data)[22 : len(*&st.Data)-3]
		// fmt.Println(jsonStr)
		var stData structs.StData
		json.Unmarshal([]byte(jsonStr), &stData)

		if len(stData.Hq) > 10 {
			for i := 0; i < 10; i++ {
				k := new(structs.K)
				createK(stData.Hq[i], k)
				kArr[i] = *k
			}
			// fmt.Println(kArr)

			//== 验证策略 ==

			totalCheck(kArr[:], curI, curI-checkDay)

			// res := check1(kArr[:], 0) //546 == 331 == 1.65

			check2(kArr[:], curI, curI-checkDay)

			//== 保存结果 ==
			toPool(kArr[:], *&st.Code, *&st.Name, *&st.Sector, *&st.IncRate)
		}
	}

	totalRatio, _ := strconv.ParseFloat(fmt.Sprintf("%.2f", float64(totalUp)/float64(total)*100), 64)
	ratio, _ := strconv.ParseFloat(fmt.Sprintf("%.2f", float64(up)/float64(findTotal)*100), 64)
	upRate, _ := strconv.ParseFloat(fmt.Sprintf("%.2f", ratio-totalRatio), 64)
	fmt.Println("Date:", kArr[curI].Date)
	fmt.Println("Total:", total, "==", totalUp, "==", totalDown, "==", totalRatio, "%")
	fmt.Println("Check:", findTotal, "==", up, "==", down, "==", ratio, "%")
	fmt.Println("UpRate:", upRate, "%")
	// fmt.Println(*st)
}

func toPool(kArr []structs.K, code string, name string, sector string, inc_rate string) {
	if strategy(kArr, 0) {
		pool = append(pool, structs.QtInfo{code, name, sector, inc_rate})
	}
}

//全部达标的
func totalCheck(kArr []structs.K, curI int, endI int) {
	res := 0
	if curI <= 0 || curI <= endI {
		fmt.Println("i 设置有误")
		return
	}

	check := false

	//红的就算
	if kArr[curI].Close > kArr[curI].Open {
		check = true
	}

	if check {
		total++
		for j := curI - 1; j >= endI; j-- {
			//kArr[j].High > kArr[j+1].High && kArr[j].Low > kArr[j+1].Low
			//kArr[j].IncRate > 0 && kArr[j].Close > kArr[j].Open
			if kArr[j].High > kArr[j+1].High && kArr[j].Low > kArr[j+1].Low {
				res = 1
			} else {
				res = 2
				break
			}
		}
	}
	if res == 1 {
		totalUp++
	} else if res == 2 {
		totalDown++
	}
}

func check2(kArr []structs.K, curI int, endI int) int {
	res := 0
	if curI <= 0 || curI <= endI {
		fmt.Println("i 设置有误")
		return res
	}

	if strategy(kArr, curI) {
		findTotal++
		for j := curI - 1; j >= endI; j-- {
			if kArr[j].High > kArr[j+1].High && kArr[j].Low > kArr[j+1].Low {
				// if kArr[j].High > kArr[curI].High && kArr[j].Low > kArr[curI].Low {
				res = 1
			} else {
				res = 2
				break
			}
		}
	}
	if res == 1 {
		up++
	} else if res == 2 {
		down++
	}
	return res
}

func check1(kArr []structs.K, i int) int {
	res := 0
	if kArr[i+1].IncRate > 0 &&
		kArr[i+1].High > kArr[i+2].High &&
		kArr[i+2].High > kArr[i+3].High &&
		kArr[i+1].Low > kArr[i+2].Low &&
		kArr[i+2].Low > kArr[i+3].Low {
		if kArr[i].High > kArr[i+1].High {
			res = 1
		} else {
			res = 2
		}
	}
	return res
}

func createK(hq []string, k *structs.K) {
	// k := new(structs.K)
	var err error
	k.Open, err = strconv.ParseFloat(hq[1], 2)
	if err != nil {
		fmt.Println("createK err:", err)
		return
	}
	//0=日期，1=开盘，2=收盘，3=涨跌额，4=涨跌幅，5=最低，6=最高，7=成交量，8=成交额，9=换手率
	k.Date = hq[0]
	k.Close, err = strconv.ParseFloat(hq[2], 2)
	k.High, err = strconv.ParseFloat(hq[6], 2)
	k.Low, err = strconv.ParseFloat(hq[5], 2)
	k.Vol, err = strconv.ParseFloat(hq[7], 2)
	k.TRate, err = strconv.ParseFloat(strings.Replace(hq[9], "%", "", 1), 2)
	k.IncRate, err = strconv.ParseFloat(strings.Replace(hq[4], "%", "", 1), 2)
}

func saveList() {
	// str := "{\"data\":\"sadf\"}"
	// str := `{"data":"sadf"}`
	// var bbb structs.Test
	// json.Unmarshal([]byte(str), &bbb)
	// fmt.Println(bbb)
	// return

	url := "https://98.push2.eastmoney.com/api/qt/clist/get?cb=jQuery112405369178137598498_1636469855448&pn=1&pz=5000&po=1&np=1&ut=bd1d9ddb04089700cf9c27f6f7426281&fltt=2&invt=2&fid=f3&fs=m:0+t:6,m:0+t:80,m:1+t:2,m:1+t:23&fields=f12,f14,f100,f3&_=1636469855466"
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
		_, err := dbClient.Exec("insert into st(name,code,sector,inc_rate)values(?,?,?,?)", v.Name, v.Code, v.Sector, v.IncRate)
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

		data := enc.ConvertString(resp)

		jsonStr := data[22 : len(data)-3]
		var stData structs.StData
		json.Unmarshal([]byte(jsonStr), &stData)
		inc_rate := 0.0
		if len(stData.Hq) > 0 {
			k := new(structs.K)
			createK(stData.Hq[0], k)
			inc_rate = k.IncRate
		}

		_, err = dbClient.Exec("update st set data = ?, inc_rate = ? where code = ?", enc.ConvertString(resp), inc_rate, *&st.Code)
		if err != nil {
			fmt.Println("===", err)
		}
	}

}

var checkI = 1
var checkDay = 1

func strategy(kArr []structs.K, i int) bool {
	return p5(kArr, i)
}

//3up	yy
func p1(kArr []structs.K, i int) bool {
	if kArr[i].IncRate > 0 && kArr[i].Close > kArr[i].Open &&
		kArr[i].Close > (kArr[i].High+kArr[i].Low)/2 &&
		kArr[i].High > kArr[i+1].High &&
		kArr[i+1].High > kArr[i+2].High &&
		kArr[i].Low > kArr[i+1].Low &&
		kArr[i+1].Low > kArr[i+2].Low {
		return true
	}
	return false
}

//in+up		y
func p2(kArr []structs.K, i int) bool {
	if kArr[i].IncRate > 0 && kArr[i].Close > kArr[i].Open && //突
		kArr[i].Close > (kArr[i].High+kArr[i].Low)/2 && //力
		kArr[i].High > kArr[i+1].High &&
		kArr[i+1].High <= kArr[i+2].High &&
		kArr[i].Low > kArr[i+1].Low &&
		kArr[i+1].Low > kArr[i+2].Low {
		return true
	}
	return false
}

//ii+up		x
func p3(kArr []structs.K, i int) bool {
	if kArr[i].IncRate > 0 && kArr[i].Close > kArr[i].Open && //突
		kArr[i].Close > (kArr[i].High+kArr[i].Low)/2 && //力
		kArr[i].High > kArr[i+1].High &&
		kArr[i].Low > kArr[i+1].Low &&
		kArr[i+1].High <= kArr[i+2].High &&
		kArr[i+1].Low > kArr[i+2].Low &&
		kArr[i+2].High <= kArr[i+3].High &&
		kArr[i+2].Low > kArr[i+3].Low {
		return true
	}
	return false
}

//2up	y
func p4(kArr []structs.K, i int) bool {
	if kArr[i].IncRate > 0 && kArr[i].Close > kArr[i].Open &&
		kArr[i].Close > (kArr[i].High+kArr[i].Low)/2 &&
		kArr[i].High > kArr[i+1].High &&
		kArr[i].Low > kArr[i+1].Low {
		return true
	}
	return false
}

//out
func p5(kArr []structs.K, i int) bool {
	if kArr[i].IncRate > 0 && kArr[i].Close > kArr[i].Open &&
		kArr[i].Close > (kArr[i].High+kArr[i].Low)/2 &&
		kArr[i].High > kArr[i+1].High &&
		kArr[i].Low < kArr[i+1].Low {
		return true
	}
	return false
}
