package dmstest

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"testing"
)

type struCacheTestReq struct {
	Cache int `json:"cache"`
	Total int `json:"total"`
	Rate  int `json:"rate"`
}

func Test_Cache(t *testing.T) {
	resp, err := http.Get("http://127.0.0.1/dev/cacheTest")
	if err != nil {
		t.Error("http Get error")
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		// handle error
	}

	cache := new(struCacheTestReq)

	err = json.Unmarshal(body, cache)
	if err != nil {
		// handle error
	}

	if cache.Rate < 90 {
		t.Error(fmt.Sprintf("cache Rate : %d", cache.Rate))
	} else {
		t.Log(fmt.Sprintf("cache Rate : %d", cache.Rate))
	}
}
