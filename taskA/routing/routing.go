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
			resultString.WriteString(fmt.Sprintln(viper.Get(fmt.Sprintf("result.routes.%d.steps.%d.instruction", result_route, step))) + " <br>")
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
			resultString.WriteString(fmt.Sprint(viper.Get(fmt.Sprintf("result.routes.%d.steps.%d.instruction", route, step))) + " <br>")
		}
	}

	return resultString.String()
}

func Directionlite_riding(origin_name string, destination_name string, riding_type string) string {
	 
	ak := string(os.Getenv("BAIDU_AK"))

	// 服务地址
	host := "https://api.map.baidu.com"

	// 接口地址
	uri := "/directionlite/v1/riding"

	orgin := geocoding(origin_name)
	destination := geocoding(destination_name)
	if riding_type == "regular_bike" {
		riding_type = "0"
	} else {
		riding_type = "1"
	}
	// 设置请求参数
	params := url.Values{
		"origin":      []string{fmt.Sprintf("%s,%s", orgin["lat"], orgin["lng"])},
		"destination": []string{fmt.Sprintf("%s,%s", destination["lat"], destination["lng"])},
		"ak":          []string{ak},
		"riding_type": []string{riding_type},
	}

	// 发起请求
	request, err := url.Parse(host + uri + "?" + params.Encode())
	if nil != err {
		fmt.Printf("host error: %v", err)
		return "host error"
	}

	resp, err1 := http.Get(request.String())
	defer resp.Body.Close()
	if err1 != nil {
		fmt.Printf("request error: %v", err1)
		return "request error"
	}
	body, err2 := io.ReadAll(resp.Body)
	if err2 != nil {
		fmt.Printf("response error: %v", err2)
	}

	viper.SetConfigType("json")
	if err := viper.ReadConfig(strings.NewReader(string(body))); err != nil {
		fmt.Println(err)
	}

	var result string

	result += "骑行路线结果：<br>"

	if viper.Get("status").(float64) != 0 {
		result += "无法到达目的地"
		return result
	}

	if riding_type == "regular_bike" {
		result += "骑行类型：普通自行车<br>"
	} else {
		result += "骑行类型：电动自行车<br>"
	}

	total_seconds := viper.Get(fmt.Sprintf("result.routes.%d.duration", 0)).(float64)
	hours, minutes, seconds := secondsToHMS(int(total_seconds))
	result += fmt.Sprintf("时间: %d小时%d分钟%d秒<br>", hours, minutes, seconds)

	result += "路线信息：<br>"
	steps := viper.Get(fmt.Sprintf("result.routes.%d.steps", 0)).([]interface{})
	for step := range steps {
		turnType := viper.Get(fmt.Sprintf("result.routes.%d.steps.%d.turn_type", 0, step))
		instruction := viper.Get(fmt.Sprintf("result.routes.%d.steps.%d.instruction", 0, step))
		result += fmt.Sprintf("%s %s<br>", turnType, instruction)
	}

	return result
}
func Directionlite_walking(origin_name string, destination_name string) string {

	 
	ak := string(os.Getenv("BAIDU_AK"))

	// 服务地址
	host := "https://api.map.baidu.com"

	// 接口地址
	uri := "/directionlite/v1/walking"

	origin := geocoding(origin_name)
	destination := geocoding(destination_name)
	// 设置请求参数
	params := url.Values{
		"origin":      []string{fmt.Sprintf("%s,%s", origin["lat"], origin["lng"])},
		"destination": []string{fmt.Sprintf("%s,%s", destination["lat"], destination["lng"])},
		"ak":          []string{ak},
	}

	// 发起请求
	request, err := url.Parse(host + uri + "?" + params.Encode())
	if nil != err {
		fmt.Printf("host error: %v", err)
		return "host error"
	}

	resp, err1 := http.Get(request.String())

	defer resp.Body.Close()
	if err1 != nil {
		fmt.Printf("request error: %v", err1)
		return "request error"
	}
	body, err2 := io.ReadAll(resp.Body)
	if err2 != nil {
		fmt.Printf("response error: %v", err2)
	}

	viper.SetConfigType("json")
	if err := viper.ReadConfig(strings.NewReader(string(body))); err != nil {
		fmt.Println(err)
	}

	var result string

	result += "步行路线结果：<br>"

	if viper.Get("status").(float64) != 0 {
		result += "无法到达目的地"
		return result
	}

	total_seconds := viper.Get(fmt.Sprintf("result.routes.%d.duration", 0)).(float64)
	hours, minutes, seconds := secondsToHMS(int(total_seconds))
	result += fmt.Sprintf("时间: %d小时%d分钟%d秒<br>", hours, minutes, seconds)

	result += "路线信息：<br>"
	steps := viper.Get(fmt.Sprintf("result.routes.%d.steps", 0)).([]interface{})
	for step := range steps {
		result += fmt.Sprintf("%s<br>", viper.Get(fmt.Sprintf("result.routes.%d.steps.%d.instruction", 0, step)))
	}

	return result
}

func Directionlite_transit(origin_name string, destination_name string, transitOptions string, transit_output_type string) string {

	 
	ak := string(os.Getenv("BAIDU_AK"))

	// 服务地址
	host := "https://api.map.baidu.com"

	// 接口地址
	uri := "/directionlite/v1/transit"

	orgin := geocoding(origin_name)
	destination := geocoding(destination_name)
	params := url.Values{
		"origin":      []string{fmt.Sprintf("%s,%s", orgin["lat"], orgin["lng"])},
		"destination": []string{fmt.Sprintf("%s,%s", destination["lat"], destination["lng"])},
		"ak":          []string{ak},
	}

	// 发起请求
	request, err := url.Parse(host + uri + "?" + params.Encode())
	if nil != err {
		fmt.Printf("host error: %v", err)
		return "host error"
	}

	resp, err1 := http.Get(request.String())
	defer resp.Body.Close()
	if err1 != nil {
		fmt.Printf("request error: %v", err1)
		return "request error"
	}
	body, err2 := io.ReadAll(resp.Body)
	if err2 != nil {
		fmt.Printf("response error: %v", err2)
	}

	viper.SetConfigType("json")
	if err := viper.ReadConfig(strings.NewReader(string(body))); err != nil {
		fmt.Println(err)
	}

	var result string

	result += "公交路线结果：<br>"

	if viper.Get("status").(float64) != 0 {
		result += "无法到达目的地"
		return result
	}

	result_route := 0
	if transitOptions == "shortest_time" {
		result += "时间最短的路线：<br>"
		result_duration := 0.0
		for route := range viper.Get("result.routes").([]interface{}) {
			if viper.Get(fmt.Sprintf("result.routes.%d.duration", route)).(float64) < result_duration || result_duration == 0.0 {
				result_duration = viper.Get(fmt.Sprintf("result.routes.%d.duration", route)).(float64)
				result_route = int(route)
			}
		}
		hours, minutes, seconds := secondsToHMS(int(result_duration))
		result += fmt.Sprintf("时间: %d小时%d分钟%d秒<br>", hours, minutes, seconds)
	} else if transitOptions == "minimize_cost" {
		result += "花费最少的路线：<br>"
		result_price := 0.0
		for route := range viper.Get("result.routes").([]interface{}) {
			if viper.Get(fmt.Sprintf("result.routes.%d.price", route)).(float64) < result_price || result_price == 0.0 {
				result_price = viper.Get(fmt.Sprintf("result.routes.%d.price", route)).(float64)
				result_route = int(route)
			}
		}
		result += fmt.Sprintf("价格: %.1f元<br>", result_price)
	}

	if transit_output_type == "stops_information" {
		result += "站点信息：<br>"

		result += "<table border='1'>"
		result += "<tr>"
		result += "<th>名称</th><th>路线方向</th><th>起点</th><th>终点</th><th>首班车时间</th><th>末班车时间</th><th>路段经过的站点数量</th>"
		result += "</tr>"

		steps := viper.Get(fmt.Sprintf("result.routes.%d.steps", result_route)).([]interface{})
		for step := range steps {
			vehicle := viper.Get(fmt.Sprintf("result.routes.%d.steps.%d.0.vehicle", result_route, step)).(map[string]interface{})
			if vehicle["name"] != "" {
				vehicle_string := fmt.Sprintf(
					"<tr><td>%s</td><td>%s</td><td>%s</td><td>%s</td><td>%s</td><td>%s</td><td>%d</td></tr>",
					vehicle["name"],
					vehicle["direct_text"],
					vehicle["start_name"],
					vehicle["end_name"],
					vehicle["start_time"],
					vehicle["end_time"],
					int(vehicle["stop_num"].(float64)),
				)
				result += vehicle_string
			}
		}

		result += "</table>"
	} else {
		result += "路线内容：<br>"
		steps := viper.Get(fmt.Sprintf("result.routes.%d.steps", result_route)).([]interface{})
		for step := range steps {
			result += fmt.Sprintf("%s<br>", viper.Get(fmt.Sprintf("result.routes.%d.steps.%d.0.instruction", result_route, step)))
		}
	}

	if transitOptions == "taxi" {
		result = ""
		result += "出租车路线结果：<br>"
		result += "白天价格：<br>"
		result += fmt.Sprintf("%v元<br>", viper.Get("result.taxi.detail.0.total_price"))

		result += "夜间价格：<br>"
		result += fmt.Sprintf("%v元<br>", viper.Get("result.taxi.detail.1.total_price"))

		result += "费用信息：<br>"
		result += fmt.Sprintf("%v<br>", viper.Get("result.taxi.remark"))

		result_duration := viper.Get("result.taxi.duration").(float64)
		hours, minutes, seconds := secondsToHMS(int(result_duration))
		result += fmt.Sprintf("时间: %d小时%d分钟%d秒<br>", hours, minutes, seconds)
	}

	return result
}
