package main

import (
	"sync"
)

func RandVerifyHttps(proxy Proxy, testTargets []string, regTexts map[string]string, localIP string, wg *sync.WaitGroup, result chan<- Proxy) {
	defer wg.Done()

	var res []bool
	for _, testTarget := range testTargets {
		res1 := CheckHttpsProxy(proxy, testTarget, regTexts[testTarget], localIP)
		res = append(res, res1)
	}
	if res[0] || res[1] || res[2] {
		proxy.Alive = true
		proxy.Type = "HTTPS"
		result <- proxy
	}

}

func RandVerifyHttp(proxy Proxy, testTargets []string, regTexts map[string]string, localIP string, wg *sync.WaitGroup, result chan<- Proxy) {
	defer wg.Done()

	var res []bool
	for _, testTarget := range testTargets {
		res1 := CheckHttpProxy(proxy, testTarget, regTexts[testTarget], localIP)
		res = append(res, res1)
	}
	if res[0] || res[1] || res[2] {
		proxy.Alive = true
		proxy.Type = "HTTP"
		result <- proxy
	}
}
