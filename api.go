package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
)

func main() {
	fmt.Println("api >>>")

	http.HandleFunc("/test", test)
	http.HandleFunc("/tGet", testGet)
	http.ListenAndServe("0.0.0.0:8089", nil)
}

func test(writer http.ResponseWriter, request *http.Request) {
	fmt.Println("test ok")
}

func testGet(writer http.ResponseWriter, request *http.Request) {
	get("https://qianggwang.zhonglingguoji.com/api/subBanner")
}

func get(url string) {
	client := &http.Client{}

	//提交请求
	reqest, err := http.NewRequest("GET", url, nil)

	if err != nil {
		fmt.Println(err)
	}

	//增加header选项
	reqest.Header.Add("OK-ACCESS-KEY", "xxxxxx")
	reqest.Header.Add("OK-ACCESS-SIGN", "xxx")
	reqest.Header.Add("OK-ACCESS-TIMESTAMP", "xxxx")
	reqest.Header.Add("OK-ACCESS-PASSPHRASE", "xxxx")

	//处理返回结果
	response, _ := client.Do(reqest)
	defer response.Body.Close()

	body, _ := ioutil.ReadAll(response.Body)
	//fmt.Println(string(body))
	fmt.Printf("Get request result: %s\n", string(body))
}
