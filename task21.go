package main

import (
    "fmt"
    "io"
    "net/http"
    "net/url"
	"os"
	"strings"
  	"github.com/spf13/viper"
)

func traffic(road_name,city string) string {
    // 此处填写您在控制台-应用管理-创建应用后获取的AK
  	ak := string(os.Getenv("BAIDU_AK"))
    
    // 服务地址
    host := "https://api.map.baidu.com"

    // 接口地址
    uri := "/traffic/v1/road"

    // 设置请求参数
    params := url.Values {
          "road_name": []string{road_name},
          "city": []string{city},
          "ak": []string{ak},
    }

    // 发起请求
    request, err := url.Parse(host + uri + "?" + params.Encode())
    if nil != err {
        fmt.Printf("host error: %v", err)
        return ""
    }

    resp, err1 := http.Get(request.String())
    fmt.Printf("url: %s\n", request.String())
    defer resp.Body.Close()
    if err1 != nil {
        fmt.Printf("request error: %v", err1)
        return ""
    }
    body, err2 := io.ReadAll(resp.Body)
    if err2 != nil {
        fmt.Printf("response error: %v", err2)
    }
  	viper.SetConfigType("json")
  	if err := viper.ReadConfig(strings.NewReader(string(body))); err != nil {
		fmt.Println(err)
	}
    return viper.Get("evaluation.status").(string)

}