package dmstest

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
		codeID := 12370 + i
		uri := "http://112.74.51.51:4637/getDevTemp?codeID=" + strconv.Itoa(codeID)
		resp, err := http.Get(uri)
		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			// handle error
		}

		fmt.Println(string(body))
	}

}
