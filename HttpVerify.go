package main

import (
	"crypto/tls"
	"fmt"
	"log"
	"regexp"
	"time"

	"github.com/imroc/req/v3"
)

// 获取当前IP地址
func GetLocalIP() string {
	client := req.C().
		SetTimeout(20 * time.Second). // 设置超时时间
		SetCommonRetryCount(3)        // 设置重试次数
	// 禁用 HTTP/2
	client.DisableH2C().
		SetTLSClientConfig(&tls.Config{
			InsecureSkipVerify: true,
		})
	resp, err := client.R().Get("http://httpbin.org/ip")
	if err != nil {
		log.Fatalf("获取当前IP错误: %v", err)
	}
	body := resp.String()
	re := regexp.MustCompile(`"origin": "([0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3})"`)
	matches := re.FindStringSubmatch(body)
	return matches[1]
}

// 验证代理是否可用
func CheckHttpsProxy(proxy Proxy, testTarget string, regText string, localIP string) bool {

	client := req.C().
		SetTimeout(20 * time.Second). // 设置超时时间
		SetCommonRetryCount(3)        // 设置重试次数
	// 禁用 HTTP/2
	client.DisableH2C().
		SetTLSClientConfig(&tls.Config{
			InsecureSkipVerify: true,
		})

	proxyHttpsURL := fmt.Sprintf("https://%s:%s", proxy.Ip, proxy.Port)
	client.SetProxyURL(proxyHttpsURL)
	// 验证https代理
	req := client.R()
	resp1, err := req.Get(testTarget)
	//resp1, err := req.Get("https://www.baidu.com")
	if err != nil {
		log.Printf("Error checking HTTPS proxy %s:%s - %v", proxy.Ip, proxy.Port, err)
		return false
	}
	if resp1.IsSuccessState() {
		//body := resp1.String()
		//re := regexp.MustCompile(`您的IP地址是：(\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3})`)
		//matches := re.FindStringSubmatch(body)
		//if len(matches) > 1 && matches[1] == proxy.Ip {
		//	proxy.Type = "https"
		//	proxy.Alive = true
		//	result <- proxy
		body := resp1.String()
		re := regexp.MustCompile(regText)
		matches := re.FindStringSubmatch(body)
		//match, _ := regexp.MatchString(`百度一下，你就知道`, body)

		if len(matches) >= 1 {
			if regText != "百度一下，你就知道" && regText != "必应" {
				if matches[1] != localIP {
					// 写入日志文件
					logContent := fmt.Sprintf("HTTPS Proxy Success: %s:%s\nRequest: %s\nResponse: %s\n\n",
						proxy.Ip, proxy.Port, req, resp1)
					writeLog("success_logs.txt", logContent)
					return true

				} else {
					return false
				}
			} else {
				// 写入日志文件
				logContent := fmt.Sprintf("HTTPS Proxy Success: %s:%s\nRequest: %s\nResponse: %s\n\n",
					proxy.Ip, proxy.Port, req, resp1)
				writeLog("success_logs.txt", logContent)
				return true
			}

		} else {
			return false
		}
	}
	return false
}

func CheckHttpProxy(proxy Proxy, testTarget string, regText string, localIP string) bool {

	client := req.C().
		SetTimeout(20 * time.Second). // 设置超时时间
		SetCommonRetryCount(3)        // 设置重试次数
	// 禁用 HTTP/2
	client.DisableH2C().
		SetTLSClientConfig(&tls.Config{
			InsecureSkipVerify: true,
		})

	proxyHttpURL := fmt.Sprintf("http://%s:%s", proxy.Ip, proxy.Port)
	client.SetProxyURL(proxyHttpURL)
	// 验证http代理
	req := client.R()
	resp1, err := req.Get(testTarget)
	//resp1, err := req.Get("http://www.baidu.com")
	if err != nil {
		log.Printf("Error checking HTTP proxy %s:%s - %v", proxy.Ip, proxy.Port, err)
		return false
	}
	if resp1.IsSuccessState() {
		body := resp1.String()
		re := regexp.MustCompile(regText)
		//re := regexp.MustCompile(`您的IP地址是：(\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3})`)
		matches := re.FindStringSubmatch(body)
		if len(matches) >= 1 && matches[1] != localIP {
			// 写入日志文件
			logContent := fmt.Sprintf("HTTP Proxy Success: %s:%s\nRequest: %s\nResponse: %s\n\n",
				proxy.Ip, proxy.Port, req, resp1)
			writeLog("success_logs.txt", logContent)
			return true
		}
	}
	return false
}
