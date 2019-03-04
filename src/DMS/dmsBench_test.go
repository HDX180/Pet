package dmstest

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"testing"
)

func BenchmarkHealth(b *testing.B) {
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		resp, err := http.Get("http://112.74.51.51:4637/getDevTemp?codeID=17349")
		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			// handle error
		}

		fmt.Println(string(body))
	}

}
