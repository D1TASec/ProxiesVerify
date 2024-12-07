package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"sync"
	"time"

	"golang.org/x/net/proxy"
)

// 将请求对象转换为字符串
func requestToString(req *http.Request) string {
	var buf bytes.Buffer
	buf.WriteString(fmt.Sprintf("%s %s %s\r\n", req.Method, req.URL, req.Proto))
	for k, v := range req.Header {
		for _, vv := range v {
			buf.WriteString(fmt.Sprintf("%s: %s\r\n", k, vv))
		}
	}
	buf.WriteString("\r\n")
	return buf.String()
}

// 将响应对象转换为字符串
func responseToString(resp *http.Response, body []byte) string {
	var buf bytes.Buffer
	buf.WriteString(fmt.Sprintf("HTTP/%d.%d %d %s\r\n", resp.ProtoMajor, resp.ProtoMinor, resp.StatusCode, resp.Status))
	for k, v := range resp.Header {
		for _, vv := range v {
			buf.WriteString(fmt.Sprintf("%s: %s\r\n", k, vv))
		}
	}
	buf.WriteString("\r\n")
	buf.Write(body)
	return buf.String()
}

func Socks5Verify(proxyStruct Proxy, wg *sync.WaitGroup, result chan<- Proxy) {
	defer wg.Done()
	socks5Proxy := fmt.Sprintf("socks5://%s:%s", proxyStruct.Ip, proxyStruct.Port)
	// 解析代理url
	proxyURL, err := url.Parse(socks5Proxy)
	if err != nil {
		log.Printf("解析代理地址失败: %v", err)
		return
	}

	// 创建socks拨号器,使用"golang.org/x/net/proxy"
	dialer, err := proxy.FromURL(proxyURL, nil) // 不设置备用拨号器
	if err != nil {
		log.Printf("创建socks拨号器失败: %v", err)
		return
	}

	// 创建自定义Transport
	transport := &http.Transport{
		Dial: dialer.Dial, // 使用 SOCKS5 拨号器
	}

	// 创建自定义HTTP客户端
	client := &http.Client{
		Transport: transport,
		Timeout:   20 * time.Second, // 增加超时时间
	}

	// 构建请求
	req, err := http.NewRequest("GET", "http://2024.ip138.com/", nil)
	//req, err := http.NewRequest("GET", "http://httpbin.org/ip", nil)
	if err != nil {
		log.Printf("创建请求失败: %v", err)
		return
	}

	// 发送请求
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("发送HTTP请求失败: %v", err)
		return
	}
	//defer resp.Body.Close()
	defer func() {
		if resp != nil && resp.Body != nil {
			resp.Body.Close()
		}
	}()

	// 读取响应体
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("读取响应体失败: %v", err)
		return
	}

	// 匹配响应包，检测是否可用
	if resp.StatusCode >= http.StatusOK && resp.StatusCode < http.StatusMultipleChoices {
		re := regexp.MustCompile(`您的IP地址是：(\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3})`)
		//re := regexp.MustCompile(`"origin":"(\d+\.\d+\.\d+\.\d+)"`)
		matches := re.FindStringSubmatch(string(body))
		//if len(matches) > 1 && matches[1] == proxyStruct.Ip {
		//	proxyStruct.Type = "socks5"
		//	proxyStruct.Alive = true
		//	result <- proxyStruct
		//
		//	// 将请求和响应信息转换为字符串
		//	reqStr := requestToString(req)
		//	respStr := responseToString(resp, body)
		//
		//	// 写入日志文件
		//	logContent := fmt.Sprintf("SOCKS5 Proxy Success: %s:%s\nRequest: %s\nResponse: %s\n\n",
		//		proxyStruct.Ip, proxyStruct.Port, reqStr, respStr)
		//	writeLog("success_logs.txt", logContent)
		//
		//	// 写入成功代理文件
		//	successProxy := fmt.Sprintf("%s:%s:%s", proxyStruct.Ip, proxyStruct.Port, proxyStruct.Type)
		//	writeLog("successful_proxies.txt", successProxy)
		//} else {
		//	fmt.Printf("IP 地址不匹配\n")
		//}
		if len(matches) >= 1 {
			proxyStruct.Type = "socks5"
			proxyStruct.Alive = true
			result <- proxyStruct

			// 将请求和响应信息转换为字符串
			reqStr := requestToString(req)
			respStr := responseToString(resp, body)

			// 写入日志文件
			logContent := fmt.Sprintf("SOCKS5 Proxy Success: %s:%s\nRequest: %s\nResponse: %s\n\n",
				proxyStruct.Ip, proxyStruct.Port, reqStr, respStr)
			writeLog("success_logs.txt", logContent)

			// 写入成功代理文件
			//successProxy := fmt.Sprintf("%s://%s:%s", proxyStruct.Type, proxyStruct.Ip, proxyStruct.Port)
			//writeLog("successful_proxies.txt", successProxy)
		}
	} else {
		fmt.Printf("收到非200的状态码: %d\n", resp.StatusCode)
	}
}
