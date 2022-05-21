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
)

var db *sql.DB
var pool []structs.QtInfo

//todo
var kNum = 40
var checkDays = 2 //控制后面验证的天数

func init() {
	db = utils.GetDB()
}

func M() {
	fmt.Println("strategy >>>")
	analysis()
}

//分析
func analysis() {
	st := new(structs.DbSt)
	var stData structs.StData

	rows, err := db.Query("select code,name,data,sector,inc_rate from st")
	if err != nil {
		fmt.Println("err1:", err)
		return
	}
	defer func() {
		if rows != nil {
			rows.Close()
		}
	}()

	var r0 []structs.Res
	var r1 []structs.Res
	// var r2 []structs.Res

	for rows.Next() {
		var kArr []structs.K
		err = rows.Scan(&st.Code, &st.Name, &st.Data, &st.Sector, &st.IncRate)
		if err != nil {
			fmt.Println("err2:", err)
			return
		}
		// fmt.Println((*&st.Data))
		jsonStr := (st.Data)[22 : len(st.Data)-3]
		// fmt.Println(jsonStr)
		json.Unmarshal([]byte(jsonStr), &stData)

		if len(stData.Hq) > kNum {
			for i := 0; i < kNum; i++ {
				k := new(structs.K)
				createK(stData.Hq[i], k)
				kArr = append(kArr, *k)
			}

			//todo
			checkP(kArr, new(p1), &r0)
			checkP(kArr, new(p6), &r1)

		}
	}

	r0Avg := float64(0)
	r1Avg := float64(0)
	upRateAvg := float64(0)
	for i := 0; i < len(r0); i++ {
		fmt.Println(r0[i].Date)
		ratio0, _ := strconv.ParseFloat(fmt.Sprintf("%.2f", float64(r0[i].UpNum)/float64(r0[i].FindNum)*100), 64)
		fmt.Println("r0：", r0[i].FindNum, r0[i].UpNum, r0[i].DownNum, ratio0, "%")

		ratio1, _ := strconv.ParseFloat(fmt.Sprintf("%.2f", float64(r1[i].UpNum)/float64(r1[i].FindNum)*100), 64)
		fmt.Println("r1：", r1[i].FindNum, r1[i].UpNum, r1[i].DownNum, ratio1, "%")

		upRate, _ := strconv.ParseFloat(fmt.Sprintf("%.2f", ratio1-ratio0), 64)
		fmt.Println(upRate, "%")

		r0Avg += ratio0
		r1Avg += ratio1
	}

	r0Avg, _ = strconv.ParseFloat(fmt.Sprintf("%.2f", (r0Avg/float64(len(r0)))), 64)
	r1Avg, _ = strconv.ParseFloat(fmt.Sprintf("%.2f", (r1Avg/float64(len(r1)))), 64)
	upRateAvg, _ = strconv.ParseFloat(fmt.Sprintf("%.2f", r1Avg-r0Avg), 64)

	fmt.Println("r0 avg：", r0Avg, "%, r1 avg：", r1Avg, "，", upRateAvg, "%")
}

func checkP(kArr []structs.K, p plan, r *[]structs.Res) {
	for i := checkDays; i < len(kArr)-2; i++ {
		ri := i - checkDays
		if len(*r) <= ri {
			*r = append(*r, structs.Res{kArr[i].Date, 0, 0, 0})
		}

		if p.p(kArr, i) {
			(*r)[ri].FindNum += 1

			//后面checkDays验证
			if checkWin2(kArr, i, checkDays, false) {
				(*r)[ri].UpNum += 1
			} else {
				(*r)[ri].DownNum += 1
			}
		}
	}
}

func checkWin(kArr []structs.K, i int, checkDays int, isSuccessive bool) bool {
	var res int
	for j := i - 1; j >= i-checkDays; j-- {
		if kArr[j].High > kArr[j+1].High && kArr[j].Low > kArr[j+1].Low && kArr[j].Close > kArr[j+1].Close {
			res = 1
			if !isSuccessive {
				break
			}
		} else {
			res = 2
			if isSuccessive {
				break
			}
		}
	}
	if res == 1 {
		return true
	} else {
		return false
	}
}

//高点有抬升1%，低点不破
func checkWin2(kArr []structs.K, i int, checkDays int, isSuccessive bool) bool {
	//破低
	for j := i - 1; j >= i-checkDays; j-- {
		if kArr[j].Low < kArr[i].Low {
			return false
		}
	}

	for j := i - 1; j >= i-checkDays; j-- {
		if ((kArr[j].High - kArr[i].Close) / kArr[i].Close) > 0.01 {
			return true
		}
	}

	return false
}

func createK(hq []string, k *structs.K) {
	var err error
	k.Open, err = strconv.ParseFloat(hq[1], 2)
	if err != nil {
		fmt.Println("createK err:", err)
		return
	}
	k.Date = hq[0]
	k.Close, _ = strconv.ParseFloat(hq[2], 64)
	k.High, _ = strconv.ParseFloat(hq[6], 64)
	k.Low, _ = strconv.ParseFloat(hq[5], 64)
	k.Vol, _ = strconv.ParseFloat(hq[7], 64)
	k.TRate, _ = strconv.ParseFloat(strings.Replace(hq[9], "%", "", 1), 64)
	k.IncRate, _ = strconv.ParseFloat(strings.Replace(hq[4], "%", "", 1), 64)
}

type plan interface {
	p(kArr []structs.K, i int) bool
}

type p0 struct{}

//最简单的对比
func (pp p0) p(kArr []structs.K, i int) bool {
	if kArr[i].IncRate > 0 && kArr[i].Close > kArr[i].Open {
		return true
	}
	return false
}

type p1 struct{}

//3up	yy
func (pp p1) p(kArr []structs.K, i int) bool {
	if kArr[i].IncRate > 0 && kArr[i].Close > kArr[i].Open &&
		kArr[i].Close > ((kArr[i].High-kArr[i].Low)*0.8+kArr[i].Low) &&
		kArr[i].High > kArr[i+1].High &&
		kArr[i+1].High > kArr[i+2].High &&
		kArr[i].Low > kArr[i+1].Low &&
		kArr[i+1].Low > kArr[i+2].Low {
		return true
	}
	return false
}

type p2 struct{}

//2up
func (pp p2) p(kArr []structs.K, i int) bool {
	if kArr[i].IncRate > 0 && kArr[i].Close > kArr[i].Open &&
		kArr[i].Close > ((kArr[i].High-kArr[i].Low)*0.7+kArr[i].Low) &&
		kArr[i+1].Close > kArr[i+1].Open {
		return true
	}
	return false
}

type p3 struct{}

//out
func (pp p3) p(kArr []structs.K, i int) bool {
	if kArr[i].IncRate > 0 && kArr[i].Close > kArr[i].Open &&
		kArr[i].Close > ((kArr[i].High-kArr[i].Low)*0.7+kArr[i].Low) &&
		kArr[i].High > kArr[i+1].High &&
		kArr[i].Low < kArr[i+1].Low {
		return true
	}
	return false
}

type p4 struct{}

//red2
func (pp p4) p(kArr []structs.K, i int) bool {
	if kArr[i].Close > kArr[i].Open &&
		kArr[i].Close > ((kArr[i].High-kArr[i].Low)*0.6+kArr[i].Low) &&
		kArr[i+1].Close > kArr[i+1].Open &&
		kArr[i].High > kArr[i+1].High &&
		kArr[i].Low > kArr[i+1].Low {
		return true
	}
	return false
}

type p5 struct{}

//一般强
func (pp p5) p(kArr []structs.K, i int) bool {
	//收红，二分一上，低点抬升或高于前高
	if kArr[i].Close > kArr[i].Open &&
		kArr[i].Close > ((kArr[i].High-kArr[i].Low)*0.6+kArr[i].Low) &&
		(kArr[i].Low >= kArr[i+1].Low || kArr[i].High > kArr[i+1].High) {
		return true
	}
	return false
}

type p6 struct{}

//测试软件公式
func (pp p6) p(kArr []structs.K, i int) bool {
	//红，高于前收盘，二分一高，低点抬高，高点抬高
	if kArr[i].Close > kArr[i].Open &&
		kArr[i].IncRate > 0 &&
		kArr[i].Close > ((kArr[i].High-kArr[i].Low)*0.7+kArr[i].Low) &&
		kArr[i].Low > kArr[i+1].Low &&
		kArr[i].High > kArr[i+1].High &&
		kArr[i+1].High > kArr[i+2].High &&
		kArr[i+1].Low > kArr[i+2].Low {
		return true
	}
	return false
}
