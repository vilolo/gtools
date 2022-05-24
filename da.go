package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"sort"

	structs "./structs"
	utils "./utils"
)

//todo
var kNum = 20

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
//== total = 0当前强势 数据 = 1前k 对比 2前前k
func da1() {
	st := new(structs.DbSt)
	var stData structs.StData
	rows, err := db.Query("select code,name,data,sector,inc_rate from st where sector <> '大盘指数' and locate('ST',name)=0")
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
	data := make(map[string]structs.DaItems)

	marketKArr := marketKArr()

	// total := 0
	// res := structs.DaItems{}
	var all_well_count int
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

			for cpi := 0; cpi < kNum-5; cpi++ {
				var item structs.DaItems
				if _, ok := data[kArr[cpi].Date]; ok {
					item = data[kArr[cpi].Date]
					// item.Low_up = item.Low_up + 1
					// data[kArr[cpi].Date] = item
				} else {
					item = structs.DaItems{}
					// item.Low_up = 0
					// data[kArr[cpi].Date] = item

					//加入大盘数据
					for k, v := range marketKArr {
						if v.Date == kArr[cpi].Date {
							item.Sz1_close = marketKArr[k+1].Close
							item.Sz2_close = marketKArr[k+2].Close
							break
						}
					}
				}

				if checkP(kArr, cpi) {
					all_well_count = 0
					item.Total++
					if kArr[cpi+1].Low > kArr[cpi+2].Low {
						item.Low_up++
						all_well_count++
					}

					if kArr[cpi+1].High > kArr[cpi+2].High {
						item.High_up++
						all_well_count++
					}

					if kArr[cpi+1].Close > kArr[cpi+2].Close {
						item.Close_up++
						all_well_count++
					}
					if all_well_count == 3 {
						item.All_well++
					}
				}
				data[kArr[cpi].Date] = item
			}

		}
	}

	var keys []string

	for k, _ := range data {
		keys = append(keys, k)
		// if v.Total > 10 {
		// 	fmt.Printf("date %s; total %d; low-up %s; high-up %s; close-up %s; all_well %s \n", k, v.Total,
		// 		fmt.Sprintf("%.2f", float32(v.Low_up)/float32(v.Total)*100),
		// 		fmt.Sprintf("%.2f", float32(v.High_up)/float32(v.Total)*100),
		// 		fmt.Sprintf("%.2f", float32(v.Close_up)/float32(v.Total)*100),
		// 		fmt.Sprintf("%.2f", float32(v.All_well)/float32(v.Total)*100))
		// }
	}

	sort.Strings(keys)

	for _, k := range keys {
		v := data[k]
		if v.Total > 10 {
			fmt.Printf("date %s; total %d; low-up %s; high-up %s; close-up %s; all-well %s; sz-close-up %s; \n", k, v.Total,
				fmt.Sprintf("%.2f", float32(v.Low_up)/float32(v.Total)*100),
				fmt.Sprintf("%.2f", float32(v.High_up)/float32(v.Total)*100),
				fmt.Sprintf("%.2f", float32(v.Close_up)/float32(v.Total)*100),
				fmt.Sprintf("%.2f", float32(v.All_well)/float32(v.Total)*100),
				fmt.Sprintf("%.2f", v.Sz1_close-v.Sz2_close))
		}
	}
}

func marketKArr() []structs.K {
	st := new(structs.DbSt)
	szRow := db.QueryRow("select code,name,data,sector,inc_rate from st where sector = '大盘指数'")
	err := szRow.Scan(&st.Code, &st.Name, &st.Data, &st.Sector, &st.IncRate)
	if err != nil {
		fmt.Println("err2:", err)
		return nil
	}
	var kArr []structs.K
	var stData structs.StData
	jsonStr := (st.Data)[22 : len(st.Data)-3]
	json.Unmarshal([]byte(jsonStr), &stData)
	for i := 0; i < 80; i++ {
		k := new(structs.K)
		utils.CreateK(stData.Hq[i], k)
		kArr = append(kArr, *k)
	}
	// fmt.Println(kArr)
	return kArr
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
