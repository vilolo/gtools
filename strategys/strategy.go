package strategys

import (
	"database/sql"
	"fmt"
	"strconv"
	"strings"

	"../structs"
	"../utils"

	// "reflect"	//reflect.TypeOf(v).String()
	"encoding/json"
	"os"
)

var db *sql.DB
var rows *sql.Rows
var pool []structs.QtInfo

func init() {
	db = utils.GetDB()
	getData()
}

func M() {
	fmt.Println("strategy >>>")
	analysis()
}

//分析
func analysis() {
	st := new(structs.DbSt)
	var stData structs.StData
	var err error
	var kArr [20]structs.K

	var pp0 = new(p0)
	var pp1 = new(p1)
	pArr := [ pp0, pp1 ]xxxxx

	for rows.Next() {
		err = rows.Scan(&st.Code, &st.Name, &st.Data, &st.Sector, &st.IncRate)
		if err != nil {
			fmt.Println("err2:", err)
			return
		}
		// fmt.Println((*&st.Data))
		jsonStr := (*&st.Data)[22 : len(*&st.Data)-3]
		// fmt.Println(jsonStr)
		json.Unmarshal([]byte(jsonStr), &stData)

		if len(stData.Hq) > 20 {
			for i := 0; i < 20; i++ {
				k := new(structs.K)
				createK(stData.Hq[i], k)
				kArr[i] = *k
			}

			//== 验证策略 ==
			check(kArr[:])

		}
	}
}

func check(kArr []structs.K) {
	i := 1
	//列多个
	checkP(kArr, new(p0), i)
}

func checkP(kArr []structs.K, p plan, i int) {
	if p.p(kArr, i) {

	}
}

func createK(hq []string, k *structs.K) {
	var err error
	k.Open, err = strconv.ParseFloat(hq[1], 2)
	if err != nil {
		fmt.Println("createK err:", err)
		return
	}
	k.Date = hq[0]
	k.Close, err = strconv.ParseFloat(hq[2], 2)
	k.High, err = strconv.ParseFloat(hq[6], 2)
	k.Low, err = strconv.ParseFloat(hq[5], 2)
	k.Vol, err = strconv.ParseFloat(hq[7], 2)
	k.TRate, err = strconv.ParseFloat(strings.Replace(hq[9], "%", "", 1), 2)
	k.IncRate, err = strconv.ParseFloat(strings.Replace(hq[4], "%", "", 1), 2)
}

type plan interface {
	p(kArr []structs.K, i int) bool
}

type p0 struct {
	up   int
	down int
}

//最简单的对比
func (pp p0) p(kArr []structs.K, i int) bool {
	if kArr[i].IncRate > 0 && kArr[i].Close > kArr[i].Open {
		return true
	}
	return false
}

type p1 struct {
	up   int
	down int
}

//3up	yy
func (pp p1) p(kArr []structs.K, i int) bool {
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

func getData() {
	var err error
	rows, err = db.Query("select code,name,data,sector,inc_rate from st")
	if err != nil {
		fmt.Println("err getData：", err)
		os.Exit(0)
	}
	defer func() {
		if rows != nil {
			rows.Close()
		}
	}()
}
