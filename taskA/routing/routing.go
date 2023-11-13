package routing

import (

	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/spf13/viper"
)


func secondsToHMS(seconds int) (int, int, int) {
	hours := seconds / 3600
	minutes := (seconds % 3600) / 60
	remainingSeconds := seconds % 60

	return hours, minutes, remainingSeconds
}
func geocoding(address string) map[string]string {
	// 此处填写您在控制台-应用管理-创建应用后获取的AK
	ak := string(os.Getenv("BAIDU_AK"))

	// 服务地址
	host := "https://api.map.baidu.com"

	// 接口地址
	uri := "/geocoding/v3"

	// 设置请求参数
	params := url.Values{
		"address": []string{address},
		"output":  []string{"json"},
		"ak":      []string{ak},
	}

	// 发起请求
	request, err := url.Parse(host + uri + "?" + params.Encode())
	if nil != err {
		fmt.Printf("host error: %v", err)
		return nil
	}

	resp, err1 := http.Get(request.String())

	defer resp.Body.Close()
	if err1 != nil {
		fmt.Printf("request error: %v", err1)
		return nil
	}

	body, err2 := io.ReadAll(resp.Body)
	if err2 != nil {
		fmt.Printf("response error: %v", err2)
	}

	viper.SetConfigType("json")
	if err := viper.ReadConfig(strings.NewReader(string(body))); err != nil {
		fmt.Println(err)
	}

	return map[string]string{"lng": fmt.Sprintf("%f", viper.Get("result.location.lng")), "lat": fmt.Sprintf("%f", viper.Get("result.location.lat"))}
}

func Directionlite_driving(origin_name string, destination_name string, tatics int) string {
	ak := string(os.Getenv("BAIDU_AK"))
	host := "https://api.map.baidu.com"
	uri := "/directionlite/v1/driving"
	orgin := geocoding(origin_name)
	destination := geocoding(destination_name)

	params := url.Values{
		"origin":      []string{fmt.Sprintf("%s,%s", orgin["lat"], orgin["lng"])},
		"destination": []string{fmt.Sprintf("%s,%s", destination["lat"], destination["lng"])},
		"ak":          []string{ak},
		"tatics":      []string{fmt.Sprintf("%d", tatics)},
	}
	tatics_string := []string{"不走高速",
		"躲避拥堵",
		"最短距离",
		"花费最少",
		"大路优先"}
	request, err := url.Parse(host + uri + "?" + params.Encode())

	if nil != err {
		return fmt.Sprintf("host error: %v", err)
	}

	resp, err := http.Get(request.String())
	defer resp.Body.Close()
	if err != nil {
		return fmt.Sprintf("request error: %v", err)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Sprintf("response error: %v", err)
	}

	viper.SetConfigType("json")
	if err := viper.ReadConfig(strings.NewReader(string(body))); err != nil {
		return fmt.Sprintf("error reading config: %v", err)
	}

	var resultString strings.Builder

	resultString.WriteString("驾车路线结果：<br>")

	if viper.Get("status").(float64) != 0 {
		resultString.WriteString("无法到达目的地<br>")
		return resultString.String()
	}

	if tatics == 0 {
		result_route := 0
		result_duration := 0.0
		for route := range viper.Get("result.routes").([]interface{}) {
			sum_duration := 0.0
			out_duration := 0.0
			for step := range viper.Get(fmt.Sprintf("result.routes.%d.steps", route)).([]interface{}) {
				traffic_status := viper.Get(fmt.Sprintf("result.routes.%d.steps.%d.traffic_condition", route, step)).([]interface{})[0].(map[string]interface{})["status"]
				duration := viper.Get(fmt.Sprintf("result.routes.%d.steps.%d.duration", route, step)).(float64)
				switch traffic_status {
				case 2:
					sum_duration += duration * 1.2
				case 3:
					sum_duration += duration * 1.4
				case 4:
					sum_duration += duration * 1.6
				default:
					sum_duration += duration
				}
				out_duration += duration
			}

			out_duration = viper.Get(fmt.Sprintf("result.routes.%d.duration", route)).(float64) - out_duration + sum_duration
			if result_duration == 0.0 || result_duration > out_duration {
				result_duration = out_duration
				result_route = int(route)
			}
		}

		resultString.WriteString("时间最短的路线：<br>")
		hours, minutes, seconds := secondsToHMS(int(result_duration))
		resultString.WriteString(fmt.Sprintf("时间: %d小时%d分钟%d秒<br>", hours, minutes, seconds))

		resultString.WriteString("路线信息：<br>")
		for step := range viper.Get(fmt.Sprintf("result.routes.%d.steps", result_route)).([]interface{}) {
			resultString.WriteString(fmt.Sprintln(viper.Get(fmt.Sprintf("result.routes.%d.steps.%d.instruction", result_route, step)))+" <br>")
		}
	} else {

		sum_duration := 0.0
		out_duration := 0.0
		route := 0
		for step := range viper.Get(fmt.Sprintf("result.routes.%d.steps", route)).([]interface{}) {
			traffic_status := viper.Get(fmt.Sprintf("result.routes.%d.steps.%d.traffic_condition", route, step)).([]interface{})[0].(map[string]interface{})["status"]
			duration := viper.Get(fmt.Sprintf("result.routes.%d.steps.%d.duration", route, step)).(float64)
			switch traffic_status {
			case 2:
				sum_duration += duration * 1.2
			case 3:
				sum_duration += duration * 1.4
			case 4:
				sum_duration += duration * 1.6
			default:
				sum_duration += duration
			}
			out_duration += duration
		}
		out_duration = viper.Get(fmt.Sprintf("result.routes.%d.duration", route)).(float64) - out_duration + sum_duration
		resultString.WriteString(fmt.Sprintf("%s 的路线：<br>", tatics_string[tatics-1]))
		hours, minutes, seconds := secondsToHMS(int(out_duration))
		resultString.WriteString(fmt.Sprintf("时间: %d小时%d分钟%d秒<br>", hours, minutes, seconds))

		resultString.WriteString("路线信息：<br>")
		for step := range viper.Get(fmt.Sprintf("result.routes.%d.steps", route)).([]interface{}) {
			resultString.WriteString(fmt.Sprint(viper.Get(fmt.Sprintf("result.routes.%d.steps.%d.instruction", route, step)))+" <br>")
		}
	}

	return resultString.String()
}

func Directionlite_riding() {

	// 此处填写您在控制台-应用管理-创建应用后获取的AK
	ak := string(os.Getenv("BAIDU_AK"))

	// 服务地址
	host := "https://api.map.baidu.com"

	// 接口地址
	uri := "/directionlite/v1/riding"
	riding_type := 0
	fmt.Scanf("%d", &riding_type)
	// 设置请求参数
	params := url.Values{
		"origin":      []string{"40.01116,116.339303"},
		"destination": []string{"39.936404,116.452562"},
		"ak":          []string{ak},
		"riding_type": []string{fmt.Sprintf("%d", riding_type)},
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

	fmt.Println("骑行路线结果：")
	if viper.Get("status").(float64) != 0 {
		fmt.Println("无法到达目的地")
		return
	}
	if riding_type == 0 {
		fmt.Println("骑行类型：普通自行车")
	} else {
		fmt.Println("骑行类型：电动自行车")
	}
	total_seconds := viper.Get(fmt.Sprintf("result.routes.%d.duration", 0)).(float64)
	hours, minutes, seconds := secondsToHMS(int(total_seconds))
	fmt.Println("时间:", fmt.Sprintf("%d小时%d分钟%d秒<br>", hours, minutes, seconds))
	fmt.Println("路线信息：")
	for step := range viper.Get(fmt.Sprintf("result.routes.%d.steps", 0)).([]interface{}) {
		fmt.Println(viper.Get(fmt.Sprintf("result.routes.%d.steps.%d.turn_type", 0, step)), viper.Get(fmt.Sprintf("result.routes.%d.steps.%d.instruction", 0, step)))
	}
}
func Directionlite_walking() {

	// 此处填写您在控制台-应用管理-创建应用后获取的AK
	ak := string(os.Getenv("BAIDU_AK"))

	// 服务地址
	host := "https://api.map.baidu.com"

	// 接口地址
	uri := "/directionlite/v1/walking"

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

	fmt.Println("步行路线结果：")
	if viper.Get("status").(float64) != 0 {
		fmt.Println("无法到达目的地")
		return
	}
	total_seconds := viper.Get(fmt.Sprintf("result.routes.%d.duration", 0)).(float64)
	hours, minutes, seconds := secondsToHMS(int(total_seconds))
	fmt.Println("时间:", fmt.Sprintf("%d小时%d分钟%d秒<br>", hours, minutes, seconds))
	fmt.Println("路线信息：")
	for step := range viper.Get(fmt.Sprintf("result.routes.%d.steps", 0)).([]interface{}) {
		fmt.Println(viper.Get(fmt.Sprintf("result.routes.%d.steps.%d.instruction", 0, step)))
	}
}
func Directionlite_transit() {

	// 此处填写您在控制台-应用管理-创建应用后获取的AK
	ak := string(os.Getenv("BAIDU_AK"))

	// 服务地址
	host := "https://api.map.baidu.com"

	// 接口地址
	uri := "/directionlite/v1/transit"

	orgin := geocoding("北京市海淀区上地十街10号")
	destination := geocoding("北京市西城区阜外大街5号")
	params := url.Values{
		"origin":      []string{fmt.Sprintf("%s,%s", orgin["lat"], orgin["lng"])},
		"destination": []string{fmt.Sprintf("%s,%s", destination["lat"], destination["lng"])},
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
	if viper.Get("status").(float64) != 0 {
		fmt.Println("无法到达目的地")
		return
	}
	fmt.Println("公交路线结果：")

	result_route := 0
	{
		fmt.Println("时间最短的路线：")
		result_duration := 0.0
		for route := range viper.Get("result.routes").([]interface{}) {
			if viper.Get(fmt.Sprintf("result.routes.%d.duration", route)).(float64) < result_duration || result_duration == 0.0 {
				result_duration = viper.Get(fmt.Sprintf("result.routes.%d.duration", route)).(float64)
				result_route = int(route)
			}
		}
		hours, minutes, seconds := secondsToHMS(int(result_duration))
		fmt.Println("时间:", fmt.Sprintf("%d小时%d分钟%d秒<br>", hours, minutes, seconds))
	}

	{
		fmt.Println("花费最少的路线：")
		result_price := 0.0
		for route := range viper.Get("result.routes").([]interface{}) {
			if viper.Get(fmt.Sprintf("result.routes.%d.price", route)).(float64) < result_price || result_price == 0.0 {
				result_price = viper.Get(fmt.Sprintf("result.routes.%d.price", route)).(float64)
				result_route = int(route)
			}
		}
		fmt.Println("价格:", fmt.Sprintf("%.1f元<br>", result_price))
	}

	fmt.Println("路线信息：")
	for step := range viper.Get(fmt.Sprintf("result.routes.%d.steps", result_route)).([]interface{}) {
		fmt.Println(viper.Get(fmt.Sprintf("result.routes.%d.steps.%d.0.instruction", result_route, step)))
	}
	fmt.Println("出租车路线结果：")
	fmt.Println("白天价格：")
	fmt.Println(viper.Get("result.taxi.detail.0.total_price"),"元")

	fmt.Println("夜间价格：")
	fmt.Println(viper.Get("result.taxi.detail.1.total_price"),"元")

	fmt.Println("费用信息：")
	fmt.Println(viper.Get("result.taxi.remark"))
	
	result_duration := viper.Get("result.taxi.duration").(float64)
	hours, minutes, seconds := secondsToHMS(int(result_duration))
	fmt.Println("时间:", fmt.Sprintf("%d小时%d分钟%d秒<br>", hours, minutes, seconds))
}
