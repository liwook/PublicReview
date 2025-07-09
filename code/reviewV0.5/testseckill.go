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

var success int32 = 0 //抢购的成功数
var send int32 = 0    //成功发起抢购的次数(即是成功发送http请求的次数)

const seckillUrl = "http://localhost:8080/api/v1/seckill/vouchers"

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
		//Go 1.21 版本通过语言规范调整，明确了循环变量在迭代体中的绑定行为，彻底解决了“闭包/Goroutine 捕获循环变量最终值”的历史问题。
		//Go 1.21 版本需要手动启动的，set GOEXPERIMENT=loopvar  # 临时生效（当前命令行窗口）。1.22版本才是默认生效

		//自己根据自己的需求修改VoucherId
		go sendSeckillRequest(10, i, &wg)

	}
	wg.Wait()

	fmt.Println("success:", success, " send:", send)
}

func sendSeckillRequest(voucherId, userId int, wg *sync.WaitGroup) {
	defer wg.Done()

	data := seckillBody{VoucherId: voucherId, UserId: userId}
	jsonData, err := json.Marshal(data)
	if err != nil {
		fmt.Printf("Error marshaling JSON: %v\n", err)
		return
	}

	client := &http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequest("POST", seckillUrl, bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Printf("Error creating request: %v\n", err)
		return
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("Error sending request: %v\n", err)
		return
	}
	defer resp.Body.Close()

	atomic.AddInt32(&send, 1)

	_, err = io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Error reading response body: %v\n", err)
		return
	}

	if resp.StatusCode == 200 {
		atomic.AddInt32(&success, 1)
	}
}
