package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"github.com/spf13/pflag"
)

var seckillSuccess int32 = 0 //抢购的成功数
var seckillSend int32 = 0    //成功发起抢购的次数(即是成功发送http请求的次数)

type seckillBody struct {
	VoucherId int `json:"voucherId"`
	UserId    int `json:"userId"`
}

func main() {
	num := pflag.IntP("num", "n", 400, "number of requests")
	pflag.Parse()
	fmt.Println("num:", *num)
	wg := sync.WaitGroup{}
	wg.Add(*num)
	for i := 0; i < *num; i++ {
		//使用闭包。不然可能很多userId都是最后的num值
		//在Go语言中，循环变量i在循环结束后仍然存在，并且它的值是循环结束时的值。
		//在代码中，i的值在循环结束后是最后的num值
		go func(i int) {
			//自己根据自己的需求修改VoucherId
			data := seckillBody{VoucherId: 11, UserId: i}
			jsonData, _ := json.Marshal(data)
			client := &http.Client{Timeout: 10 * time.Second}
			req, err := http.NewRequest("POST", "http://localhost:30000/seckill", bytes.NewBuffer(jsonData))
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
			atomic.AddInt32(&seckillSend, 1)
			defer resp.Body.Close()

			_, err = io.ReadAll(resp.Body)
			if err != nil {
				fmt.Println("Error reading response body: %v", err)
				wg.Done()
				return
			}
			if resp.StatusCode == 200 {
				atomic.AddInt32(&seckillSuccess, 1)
			}
			wg.Done()
		}(i)
	}
	wg.Wait()

	fmt.Println("success:", seckillSuccess, " send:", seckillSend)
}
