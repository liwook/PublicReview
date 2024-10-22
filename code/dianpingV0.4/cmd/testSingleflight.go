package main

import (
	"fmt"
	"io"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"github.com/spf13/pflag"
)

var success int32 = 0 //抢购的成功数
var send int32 = 0    //成功发起抢购的次数(即是成功发送http请求的次数)

func main() {
	num := pflag.IntP("num", "n", 500, "number of requests")
	pflag.Parse()
	fmt.Println("num:", *num)
	var wg sync.WaitGroup
	wg.Add(*num)
	for i := 0; i < *num; i++ {
		go func() {
			client := &http.Client{Timeout: 10 * time.Second}
			req, err := http.NewRequest("GET", "http://localhost:30000/shop/2", nil)
			if err != nil {
				fmt.Println("Error creating request: %v", err)
				wg.Done()
				return
			}

			resp, err := client.Do(req)
			if err != nil {
				fmt.Println("Error sending request: %v", err)
				wg.Done()
				return
			}
			atomic.AddInt32(&send, 1)
			defer resp.Body.Close()

			_, err = io.ReadAll(resp.Body)
			if err != nil {
				fmt.Println("Error reading response body: %v", err)
				wg.Done()
				return
			}
			if resp.StatusCode == 200 {
				atomic.AddInt32(&success, 1)
			}
			wg.Done()
		}()
	}
	wg.Wait()

	fmt.Println("success:", success, " send:", send)
}
