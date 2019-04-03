package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"testing"
)

func BenchmarkHealth(b *testing.B) {
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		codeID := 13370 + i
		uri := "http://127.0.0.1:4637/getDevTemp?codeID=" + strconv.Itoa(codeID)
		resp, err := http.Get(uri)

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			// handle error
			continue
			resp.Body.Close()
		}

		fmt.Println(string(body))
		resp.Body.Close()
	}

}
