package main

import (
	"database/sql"
	"encoding/json"
	"flag"
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
	DB.SetConnMaxLifetime(100 * time.Minute) //最大连接周期，超过时间的连接就close
	DB.SetMaxOpenConns(1000)                 //设置最大连接数
	DB.SetMaxIdleConns(100)                  //设置闲置连接数
	dbClient = DB
}

var p = 0

func main() {
	fmt.Println("start !!")
	t := flag.String("t", "", "type")
	pp := flag.Int("p", 0, "p")
	flag.Parse()
	p = *pp

	// 获取列表
	// saveList()
	// return

	// 获取历史	-t=update
	if *t == "update" {
		updateHistory()
	} else if *t == "get" {
		// 数据处理
		handleData()

		// 处理结果
		handlePool()
	} else if *t == "list" {
		//重新获取列表
		saveList()
	} else {
		fmt.Println("type ERR !!!")
	}

	fmt.Println("end >>")
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
	<script src="http://libs.baidu.com/jquery/2.1.4/jquery.min.js"></script>
	<script>
		var html = ""
		data = %s
		$(".info").html(data.length)
		data.forEach((element,i) => {
			var code = element.f12[0]==6 ? ("sh"+element.f12) : ("sz"+element.f12);
			html += "<span><span>+++"+(i+1)+"// "+element.f12+" //"+element.f14+"//"+element.f100+"//"+element.f3+"%%</span><br><hr><br>"+"<img src=\"https://image.sinajs.cn/newchart/daily/n/"+code+".gif\"><img src=\"https://image.sinajs.cn/newchart/weekly/n/"+code+".gif\"><img src=\"https://image.sinajs.cn/newchart/monthly/n/"+code+".gif\"><br><br>"
		});
		$(".box").html(html)
	</script>
	</html>`, jsonStr)
		// utils.WriteFile("./data/"+time.Now().Format("20060102")+"-"+pName+".html", html)
		utils.WriteFile("./data/"+kDate+"-"+pName+".html", html)
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
var kDate = ""

const KNum int = 20 //使用到的k线数量

func handleData() {
	rows, err := dbClient.Query("select code,name,data,sector,inc_rate from st where locate('ST',name)=0 and locate('退',name)=0 order by sector")
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

	var kArr [KNum + 10]structs.K
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

		if len(stData.Hq) > KNum {
			for i := 0; i < KNum; i++ {
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

	kDate = kArr[0].Date

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
	// if strategy(kArr, 0) && !strings.Contains(name, "ST") {
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
		if strings.Contains(v.Name, "ST") {
			continue
		}
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
	// _, err := dbClient.Exec("update st set data = null")
	rows, err := dbClient.Query("select code, sector from st where updated_at is null or updated_at < ?", time.Now().Format("2006-01-02"))
	// rows, err := dbClient.Query("select code from st limit 2")

	//报错可能是超时：https://blog.csdn.net/xiangzaixiansheng/article/details/125558282
	// defer func() {
	// 	if rows != nil {
	// 		rows.Close() //可以关闭掉未scan连接一直占用
	// 	}
	// }()
	defer rows.Close()

	endDate := time.Now().Format("20060102")
	if err != nil {
		fmt.Printf("Query failed,err:%v", err)
		return
	}
	st := new(structs.DbSt)
	var code string
	for rows.Next() {
		err = rows.Scan(&st.Code, &st.Sector) //不scan会导致连接不释放
		if err != nil {
			fmt.Printf("Scan failed,err:%v", err)
			return
		}

		if *&st.Sector == "大盘指数" {
			code = "zs_" + *&st.Code
		} else {
			code = "cn_" + *&st.Code
		}

		url := fmt.Sprintf("https://q.stock.sohu.com/hisHq?code=%s&start=%s&end=%s&stat=1&order=D&period=d&callback=historySearchHandler&rt=jsonp", code, START_DATE, endDate)
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

		_, err = dbClient.Exec("update st set data = ?, inc_rate = ?, updated_at = ? where code = ?", enc.ConvertString(resp), inc_rate, time.Now().Format("2006-01-02 03:04:05"), *&st.Code)
		if err != nil {
			fmt.Println("===", err)
		}
	}

}

var checkI = 1
var checkDay = 1
var pName = "no"

//todo
func strategy(kArr []structs.K, i int) bool {
	if p == 1 {
		return p1(kArr, i)
	}

	if p == 5 {
		return p5(kArr, i)
	}

	if p == 6 {
		return p6(kArr, i)
	}

	if p == 7 {
		return p7(kArr, i)
	}

	if p == 22 {
		return p22(kArr, i)
	}

	if p == 23 {
		return p23(kArr, i)
	}

	if p == 100 {
		return p100(kArr, i)
	}

	if p == 101 {
		return p101(kArr, i)
	}

	return p1(kArr, i)
}

//3up	yy
func p1(kArr []structs.K, i int) bool {
	pName = "up"
	if kArr[i].IncRate > 0 && kArr[i].Close > kArr[i].Open &&
		// kArr[i].Close > (kArr[i].High+kArr[i].Low)/2 &&
		kArr[i].Close > ((kArr[i].High-kArr[i].Low)*0.7+kArr[i].Low) &&
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
	pName = "out"
	if kArr[i].IncRate > 0 && kArr[i].Close > kArr[i].Open &&
		kArr[i].Close > (kArr[i].High+kArr[i].Low)/2 &&
		kArr[i].High > kArr[i+1].High &&
		kArr[i].Low < kArr[i+1].Low {
		return true
	}
	return false
}

//red2
func p6(kArr []structs.K, i int) bool {
	pName = "2up"
	if kArr[i].Close > kArr[i].Open &&
		kArr[i].Close > ((kArr[i].High-kArr[i].Low)*0.6+kArr[i].Low) &&
		kArr[i+1].Close > kArr[i+1].Open &&
		kArr[i].High > kArr[i+1].High &&
		kArr[i].Low > kArr[i+1].Low {
		return true
	}
	return false
}

func p7(kArr []structs.K, i int) bool {
	pName = "2s"
	if kArr[i].Close > kArr[i].Open &&
		kArr[i].Close > ((kArr[i].High-kArr[i].Low)*0.6+kArr[i].Low) &&
		kArr[i+1].Close > ((kArr[i+1].High-kArr[i+1].Low)*0.6+kArr[i+1].Low) {
		return true
	}
	return false
}

func p22(kArr []structs.K, i int) bool {
	pName = "p22"
	if kArr[i].Close > kArr[i].Open && //52.45 %

		//加强
		kArr[i].Close > kArr[i+1].Close && //52.45 %
		kArr[i].High > kArr[i+1].High && //50.54 %
		kArr[i].Open <= ((kArr[i].High-kArr[i].Low)*0.2+kArr[i].Low) && //46.73 %
		// kArr[i].Low >= kArr[i+1].Low && //52.3 %	52.45 %

		// positionRate <= 0.4 && //51.74 %	52.45 %
		kArr[i].Close > ((kArr[i].High-kArr[i].Low)*0.9+kArr[i].Low) && //43.14 %
		(kArr[i].Low >= kArr[i+1].Low ||
			kArr[i].Close > ((kArr[i].High-kArr[i].Low)*0.95+kArr[i].Low)) && //52.29 %
		1 == 1 {
		return true
	}
	return false
}

func p23(kArr []structs.K, i int) bool {
	pName = "p23"
	if kArr[i].IncRate > 0 && //52.61
		kArr[i].Close > kArr[i].Open && //52.61
		kArr[i].Close > ((kArr[i].High-kArr[i].Low)*0.9+kArr[i].Low) && //43.56
		kArr[i].High > kArr[i+1].High && //52.13
		kArr[i+1].High > kArr[i+2].High && //52.32
		kArr[i].Low > kArr[i+1].Low && //52.23
		kArr[i+1].Low > kArr[i+2].Low && //52.79

		kArr[i].Open <= ((kArr[i].High-kArr[i].Low)*0.2+kArr[i].Low) && //49.13

		1 == 1 {
		return true
	}
	return false
}

func p100(kArr []structs.K, i int) bool {
	pName = "check100"
	if 1 == 1 &&
		kArr[i].IncRate > 1.5 &&
		kArr[i].Close > ((kArr[i].High-kArr[i].Low)*0.7+kArr[i].Low) &&
		1 == 1 {
		return true
	}
	return false
}

func p101(kArr []structs.K, i int) bool {
	pName = "101"
	max := float64(0)
	min := float64(0)
	for j := i; j < KNum+i; j++ {
		if max == 0 {
			max = kArr[j].High
			min = kArr[j].Low
		} else {
			if kArr[j].High > max {
				max = kArr[j].High
			}
			if kArr[j].Low < min {
				min = kArr[j].Low
			}
		}
	}
	positionRate := (kArr[i].Low - min) / (max - min)

	if 1 == 1 &&
		//1k:red,up,over-75,open-l
		kArr[i].IncRate > 0 &&
		kArr[i].Close > kArr[i].Open &&
		kArr[i].Close > ((kArr[i].High-kArr[i].Low)*0.9+kArr[i].Low) && kArr[i].Open <= ((kArr[i].High-kArr[i].Low)*0.2+kArr[i].Low) &&
		//3k:c-up,h-up,l-up,
		kArr[i].Close > kArr[i+1].Close &&
		kArr[i+1].Close > kArr[i+2].Close &&
		kArr[i].High > kArr[i+1].High &&
		kArr[i+1].High > kArr[i+2].High &&
		kArr[i].Low > kArr[i+1].Low &&
		kArr[i+1].Low > kArr[i+2].Low &&
		//10k:l-less-30
		positionRate < 0.45 &&
		1 == 1 {
		return true
	}
	return false
}
