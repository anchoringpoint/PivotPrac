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

func main() {
	// 此处填写您在控制台-应用管理-创建应用后获取的AK
	ak := string(os.Getenv("BAIDU_AK"))

    var tatics int
    fmt.Scanf("%s",&tatics)
	// 服务地址
	host := "https://api.map.baidu.com"

	// 接口地址
	uri := "/directionlite/v1/driving"

	// 设置请求参数
	orgin:=geocoding("北京市海淀区上地十街10号")
	destination:=geocoding("北京市西城区阜外大街2号")
	params := url.Values{
		"origin":      []string{fmt.Sprintf("%s,%s",orgin["lat"],orgin["lng"])},
		"destination": []string{fmt.Sprintf("%s,%s",destination["lat"],destination["lng"])},
		"ak":          []string{ak},
        "tatics":      []string{string(tatics)},
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


    sum_duration:=0.0
    for key :=range viper.Get("result.routes.0.steps").([]interface{}) {
        traffic_status:=viper.Get(fmt.Sprintf("result.routes.0.steps.%d.traffic_condition", key)).( []  interface {})[0].(map[string]interface{})["status"]
        duration:=viper.Get(fmt.Sprintf("result.routes.0.steps.%d.duration", key)).(float64)
        switch traffic_status {
            case 2:
                sum_duration+=duration*1.2
            case 3:
                sum_duration+=duration*1.4
            case 4:
                sum_duration+=duration*1.6
            default:
                sum_duration+=duration
        }

    }
    fmt.Println("总耗时",sum_duration)
    fmt.Println(viper.Get("result.routes.0.duration"))
    //for key :=range viper.Get("result.routes.0.steps").([]interface{}) {
		//start_location:=viper.Get(fmt.Sprintf("result.routes.0.steps.%d.start_location", key)).(map[string]interface{})
        //fmt.Println(fmt.Sprintf("第%d步起点经纬度",key),start_location["lng"],start_location["lat"])

        //end_location:=viper.Get(fmt.Sprintf("result.routes.0.steps.%d.end_location", key)).(map[string]interface{})
        //fmt.Println(fmt.Sprintf("第%d步终点经纬度",key),end_location["lng"],end_location["lat"])
    //}

}

