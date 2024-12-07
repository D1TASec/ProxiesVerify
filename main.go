package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"
	"time"
)

// 使用多线程验证http以及https和socks5的代理是否可用
type Proxy struct {
	Ip    string
	Port  string
	Type  string
	Alive bool
}

// 将日志写入文件
func writeLog(filename, log string) {
	file, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Println("Open Log File Error:", err)
		return
	}
	defer file.Close()

	if _, err := file.WriteString(log + "\n"); err != nil {
		fmt.Println("Write Log Error:", err)
	}

	err = file.Sync() // 强制将缓冲区中的数据写入文件
	if err != nil {
		fmt.Println("Sync Log Error:", err)
	}
}

func main() {
	httpTestUrlList := []string{}
	httpTestMap := make(map[string]string)
	httpsTestUrlList := []string{}
	httpsTestMap := make(map[string]string)
	httpTestUrlList, httpTestMap = GetFileText("http")
	httpsTestUrlList, httpsTestMap = GetFileText("https")
	start := time.Now()
	// 从txt文件中读取
	file, err := os.Open("proxy.txt")
	if err != nil {
		log.Println("Open File Error:", err)
		return
	}
	defer file.Close()

	var proxies []Proxy
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		splitArray := strings.Split(line, ",")
		if len(splitArray) == 2 {
			proxies = append(proxies, Proxy{Ip: splitArray[0], Port: splitArray[1]})
		}
	}
	// 读取文件错误
	if err := scanner.Err(); err != nil {
		log.Println("Error reading file:", err)
		return
	}

	const maxGoroutines = 6000                // 最大并发数量
	var wg sync.WaitGroup                     // 用于等待所有goroutine完成
	results := make(chan Proxy, 4000)         // 用于接收验证结果
	sem := make(chan struct{}, maxGoroutines) // 信号量
	localIP := GetLocalIP()                   // 获取本地IP
	// 启动一个单独的协程来读取并处理results通道中的结果,该协程会一直运行，直到通道关闭
	//wg.Add(1)
	go func() {
		//defer wg.Done()
		for result := range results {
			fmt.Println(result.Ip, result.Port, result.Type, result.Alive)
			writeLog("successful_proxies.txt", fmt.Sprintf("%s://%s:%s", result.Type, result.Ip, result.Port))
		}
	}()

	for _, proxy := range proxies {
		wg.Add(3)         // 每个代理需要验证HTTP和HTTPS
		sem <- struct{}{} // 获取信号量
		go func(proxy Proxy) {
			defer func() { <-sem }() // 释放信号量
			httpTestUrl, regText1 := GetReg(httpTestUrlList, httpTestMap)
			//CheckHttpProxy(proxy, httpTestUrl, regText1, &wg, results)
			RandVerifyHttp(proxy, httpTestUrl, regText1, localIP, &wg, results)
		}(proxy)

		sem <- struct{}{}
		go func(proxy Proxy) {
			defer func() { <-sem }()
			httpsTestUrl, regText2 := GetReg(httpsTestUrlList, httpsTestMap)
			//CheckHttpsProxy(proxy, httpsTestUrl, regText2, &wg, results)
			RandVerifyHttps(proxy, httpsTestUrl, regText2, localIP, &wg, results)
		}(proxy)

		sem <- struct{}{}
		go func(proxy Proxy) {
			defer func() { <-sem }()
			Socks5Verify(proxy, &wg, results)
		}(proxy)
	}

	wg.Wait() // 等待所有goroutine完成
	time.Sleep(3 * time.Second)
	close(results)

	end := time.Now()
	// 计算并打印程序运行时间
	fmt.Printf("程序运行时间：%v\n", end.Sub(start))
}
