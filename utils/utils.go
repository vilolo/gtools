package utils

import (
	"io/ioutil"
	"net/http"
)

func GET(url string) (string, error) {
	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}
	req.Header.Add("charset", "utf-8")
	resp, err := client.Do(req)
	body, err := ioutil.ReadAll(resp.Body)
	return string(body), nil
}
