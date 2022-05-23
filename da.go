package main

import (
	"database/sql"
	"encoding/json"
	"fmt"

	structs "./structs"
	utils "./utils"
)

//todo
var kNum = 10

var db *sql.DB

func init() {
	db = utils.GetDB()
}

func main() {
	fmt.Println("Data Analysis")
	da1()
}

//选出达标的
//列出要素
//判断要素是否达标
//计算要素达标比例，要素可能不同百分百是不同要素，如高1%和高2%
//找出所有比例高的，说明相关性高
func da1() {
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

	// data := make(map[string]interface{})
	// data := make(map[string]structs.DaItems)

	total := 0
	res := structs.DaItems{}
	for rows.Next() {
		var kArr []structs.K
		err = rows.Scan(&st.Code, &st.Name, &st.Data, &st.Sector, &st.IncRate)
		if err != nil {
			fmt.Println("err2:", err)
			return
		}
		jsonStr := (st.Data)[22 : len(st.Data)-3]
		json.Unmarshal([]byte(jsonStr), &stData)

		if len(stData.Hq) > kNum {
			for i := 0; i < kNum; i++ {
				k := new(structs.K)
				utils.CreateK(stData.Hq[i], k)
				kArr = append(kArr, *k)
			}

			cpi := 1
			if checkP(kArr, cpi) {
				total++

				if kArr[cpi+1].Low > kArr[cpi+2].Low {
					res.Low_up++
					// if _, ok := data[kArr[cpi].Date]; ok {
					// 	item := data[kArr[cpi].Date]
					// 	item.Low_up = item.Low_up + 1
					// 	data[kArr[cpi].Date] = item
					// } else {
					// 	item := structs.DaItems{}
					// 	item.Low_up = 0
					// 	data[kArr[cpi].Date] = item
					// }
				}

				if kArr[cpi+1].High > kArr[cpi+2].High {
					res.High_up++
				}

				if kArr[cpi+1].Close > kArr[cpi+2].Close {
					res.Close_up++
				}
			}
		}
	}
	fmt.Println(total)
	fmt.Println(res)
}

func checkP(kArr []structs.K, i int) bool {
	return p0(kArr, i)
}

func p0(kArr []structs.K, i int) bool {
	if kArr[i].Close > kArr[i].Open &&
		kArr[i].Close > kArr[i+1].Close &&
		kArr[i].High > kArr[i+1].High &&
		kArr[i].Open <= ((kArr[i].High-kArr[i].Low)*0.3+kArr[i].Low) &&
		kArr[i].Low >= kArr[i+1].Low &&
		kArr[i].Close > ((kArr[i].High-kArr[i].Low)*0.8+kArr[i].Low) {
		return true
	}
	return false
}
