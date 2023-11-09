package main

import (
	"fmt"
	"github.com/spf13/viper"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
)

func main() {
	// 此处填写您在控制台-应用管理-创建应用后获取的AK
	ak := string(os.Getenv("BAIDU_AK"))

	// 服务地址
	host := "https://api.map.baidu.com"

	// 接口地址
	uri := "/directionlite/v1/driving"

	// 设置请求参数
	params := url.Values{
		"origin":      []string{"40.01116,116.339303"},
		"destination": []string{"39.936404,116.452562"},
		"ak":          []string{ak},
	}

	// 发起请求
	request, err := url.Parse(host + uri + "?" + params.Encode())
	if nil != err {
		fmt.Printf("host error: %v", err)
		return
	}

	resp, err1 := http.Get(request.String())
	defer resp.Body.Close()
	if err1 != nil {
		fmt.Printf("request error: %v", err1)
		return
	}
	body, err2 := io.ReadAll(resp.Body)
	if err2 != nil {
		fmt.Printf("response error: %v", err2)
	}
	viper.SetConfigType("json")
	if err := viper.ReadConfig(strings.NewReader(string(body))); err != nil {
		fmt.Println(err)
	}

	fmt.Println(viper.Get("result.routes.0.duration"))
    

    for key :=range viper.Get("result.routes.0.steps").([]interface{}) {
        start_location:=viper.Get(fmt.Sprintf("result.routes.0.steps.%d.start_location", key)).(map[string]interface{})
        fmt.Println(fmt.Sprintf("第%d步起点经纬度",key),start_location["lng"],start_location["lat"])
        end_location:=viper.Get(fmt.Sprintf("result.routes.0.steps.%d.end_location", key)).(map[string]interface{})
        fmt.Println(fmt.Sprintf("第%d步终点经纬度",key),end_location["lng"],end_location["lat"])
    }


}
