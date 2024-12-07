package main

import (
	"fmt"
	"io"
	"log"
	"math/rand"
	"os"
	"strings"
	"time"

	"github.com/imroc/req/v3"
	"gopkg.in/yaml.v3"
)

// ProxyConfig 定义配置文件的结构
type ProxyConfig struct {
	HTTP  []ProxyItem `yaml:"http"`
	HTTPS []ProxyItem `yaml:"https"`
}

type ProxyItem struct {
	URL   string `yaml:"url"`
	Regex string `yaml:"regex"`
}

// 测试站点是是否可用
func CheckURL(URL string) bool {
	client := req.C()
	resp, err := client.R().Get(URL)
	if err != nil {
		log.Printf("测试站点请求失败: %v", err)
		return false
	}
	if resp.IsSuccessState() {
		return true
	}
	return false
}
func GetFileText(matchPath string) (proxyList []string, proxyMap map[string]string) {
	proxyMap = make(map[string]string)
	file, err := os.Open("VerifyURL.yaml") // 打开YAML文件
	if err != nil {
		log.Println("Open File Error:", err)
		return nil, nil
	}
	defer file.Close()

	yamlText, err := io.ReadAll(file)
	if err != nil {
		log.Println("Read File Error:", err)
		return nil, nil
	}

	var config ProxyConfig
	err = yaml.Unmarshal(yamlText, &config)
	if err != nil {
		log.Println("YAML Unmarshal Error:", err)
		return nil, nil
	}

	var items []ProxyItem // 定义一个切片，用于存储ProxyItem
	if matchPath == "http" {
		items = config.HTTP
	} else if matchPath == "https" {
		items = config.HTTPS
	} else {
		log.Println("Invalid matchPath:", matchPath)
		return nil, nil
	}

	for _, item := range items {
		url := item.URL
		regex := item.Regex
		mapKey := strings.TrimSpace(url)
		mapValue := regex
		//fmt.Println(mapValue)
		//test, _ := strconv.Unquote(mapValue)
		//fmt.Println(test)
		if CheckURL(mapKey) {
			proxyList = append(proxyList, mapKey)
			proxyMap[mapKey] = mapValue
		}
	}
	return
}

func GetReg(testURL []string, testMap map[string]string) (proxyURLs []string, regTexts map[string]string) {
	// 初始化
	regTexts = make(map[string]string)
	// 根据testURL 的长度生成随机整数
	// 设置随机数种子，利用当前时间的纳秒数，确保每次运行结果有随机性
	rand.Seed(time.Now().UnixNano())

	// 获取列表长度，以此来确定有效随机数的范围（索引范围）
	listLength := len(testURL)
	if listLength < 3 {
		fmt.Println("列表中少于三个可用测试站点。")
		return
	}

	// 创建一个切片来存储已经选择的索引
	selectedIndices := make(map[int]bool)

	// 循环三次以选择三个不同的随机索引
	for i := 0; i < 3; i++ {
		var randomIndex int
		// 确保生成的索引不重复
		for {
			randomIndex = rand.Intn(listLength)
			if !selectedIndices[randomIndex] {
				selectedIndices[randomIndex] = true
				break
			}
		}
		proxyURLs = append(proxyURLs, testURL[randomIndex])
		regTexts[testURL[randomIndex]] = testMap[testURL[randomIndex]]
	}
	return proxyURLs, regTexts
}

//func GetFileText(matchPath string) (proxyList []string, proxyMap map[string]string) {
//	proxyMap = make(map[string]string)
//	file, err := os.Open("VerifyURL.json")
//	if err != nil {
//		log.Println("Open File Error:", err)
//		return nil, nil
//	}
//	defer file.Close()
//
//	jsonText, err := io.ReadAll(file)
//	if err != nil {
//		log.Println("Read File Error:", err)
//		return nil, nil
//	}
//
//	results := gjson.Get(string(jsonText), matchPath) // 获取JSON中的"http"字段
//
//	for _, result := range results.Array() {
//		splitList := strings.Split(result.String(), ":")
//		mapKey := splitList[0] + ":" + splitList[1]
//		mapKey = strings.ReplaceAll(mapKey, "\r\n", "")
//		mapKey = strings.ReplaceAll(mapKey, "\"", "")
//		mapKey = strings.ReplaceAll(mapKey, "{", "")
//		mapKey = strings.TrimSpace(mapKey)
//		mapValue := splitList[2]
//		fmt.Println(mapValue)
//		test, _ := strconv.Unquote(mapValue)
//		fmt.Println(test)
//		if CheckURL(mapKey) {
//			proxyList = append(proxyList, mapKey)
//			proxyMap[mapKey] = mapValue
//		}
//
//	}
//	return
//
//}
