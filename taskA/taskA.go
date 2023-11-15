package main

import (
	"fmt"
	"html/template"
	"net/http"
	"taskA/routing"
	"taskA/simpleorm"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

func main() {
	router := gin.Default()
	e, er := simpleorm.NewMysql("root", "root", "localhost:3306", "record")
	if er != nil {
		fmt.Println(er)
	}
	// 处理multipart forms提交文件时默认的内存限制是32 MiB
	// 可以通过下面的方式修改

	// router.MaxMultipartMemory = 8 << 20  // 8 MiB

	router.LoadHTMLGlob("templates/**/*")
	router.GET("/", func(c *gin.Context) {
		c.Redirect(http.StatusMovedPermanently, "/index")
	})
	router.GET("/index", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index/index.html", nil)
	})
	router.GET("/routing", func(c *gin.Context) {
		c.HTML(http.StatusOK, "routing/routing.html", nil)
	})

	router.POST("/routing", func(c *gin.Context) {
		transport_type := c.PostForm("transport_type")
		origin := c.PostForm("orgin")
		destination := c.PostForm("destination")
		{
			orgin_alias, err := e.Table("alias").Where("alias", origin).Select()
			if len(orgin_alias) == 0 || err != nil {
			} else {
				origin = orgin_alias[0]["name"]
			}
		}
		{
			destination_alias, err := e.Table("alias").Where("alias", destination).Select()
			if len(destination_alias) == 0 || err != nil {
			} else {
				destination = destination_alias[0]["name"]
			}
		}
		println(origin, destination)
		route_preference := c.PostForm("route_preference")
		riding_type := c.PostForm("riding_type")
		transit_preference := c.PostForm("transit_preference")
		transit_output_type := c.PostForm("transit_output_type")

		if transport_type == "driving" {
			tatics_map := map[string]int{
				"fastest":           0,
				"no_highways":       1,
				"avoid_traffic":     2,
				"shortest_distance": 3,
				"minimize_cost":     4,
				"prefer_main_roads": 5,
			}
			dynamicHTML := routing.Directionlite_driving(origin, destination, tatics_map[route_preference])
			c.HTML(http.StatusOK, "routing/routing.html", gin.H{
				"DynamicHTML": template.HTML(dynamicHTML),
			})
		} else if transport_type == "riding" {
			dynamicHTML := routing.Directionlite_riding(origin, destination, riding_type)
			c.HTML(http.StatusOK, "routing/routing.html", gin.H{
				"DynamicHTML": template.HTML(dynamicHTML),
			})
		} else if transport_type == "transit" {
			dynamicHTML := routing.Directionlite_transit(origin, destination, transit_preference, transit_output_type)
			c.HTML(http.StatusOK, "routing/routing.html", gin.H{
				"DynamicHTML": template.HTML(dynamicHTML),
			})
		} else {
			dynamicHTML := routing.Directionlite_walking(origin, destination)
			c.HTML(http.StatusOK, "routing/routing.html", gin.H{
				"DynamicHTML": template.HTML(dynamicHTML),
			})
		}
	})
	router.GET("/route", func(c *gin.Context) {
		out, err := e.Table("route").Select()
		if err != nil {
			fmt.Println(err)
		}
		c.HTML(http.StatusOK, "route/route.html", gin.H{
			"StringDictArray": out,
		})
	})
	router.POST("/route", func(c *gin.Context) {
		route_name := c.PostForm("route_name")
		route_orgin := c.PostForm("route_orgin")
		route_destination := c.PostForm("route_destination")

		e.Table("route").Insert(simpleorm.Route{
			Route:       route_name,
			Origin:      route_orgin,
			Destination: route_destination,
		})
		c.Redirect(http.StatusMovedPermanently, "/route")
	})

	router.DELETE("/route", func(c *gin.Context) {
		id := c.Query("id")
		e.Table("route").Where("id", id).Delete()
		c.Redirect(http.StatusMovedPermanently, "/route")

	})
	router.PUT("/route", func(c *gin.Context) {
		id := c.Query("id")
		viper.SetConfigType("json")
		var data map[string]string
		if err := c.BindJSON(&data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON"})
	}
	route_to_update := simpleorm.Route{
		Route:  data["route"],
		Origin: data["origin"],
		Destination: data["destination"],
	}

	e.Table("route").Where("id", id).Update(route_to_update)
	
	})
		router.GET("/alias", func(c *gin.Context) {
		out, err := e.Table("alias").Select()
		if err != nil {
			fmt.Println(err)
		}
		c.HTML(http.StatusOK, "alias/alias.html", gin.H{
			"StringDictArray": out,
		})
	})
	router.POST("/alias", func(c *gin.Context) {
		alias_name := c.PostForm("alias_name")
		name := c.PostForm("name")

		alias := simpleorm.Alias{
			Name:  name,
			Alias: alias_name,
		}

		e.Table("alias").Insert(alias)
		c.Redirect(http.StatusMovedPermanently, "/alias")

	})
	router.DELETE("/alias", func(c *gin.Context) {

		id := c.Query("id")
		e.Table("alias").Where("id", id).Delete()
		c.Redirect(http.StatusMovedPermanently, "/alias")

	})
	router.PUT("/alias", func(c *gin.Context) {
		id := c.Query("id")
		viper.SetConfigType("json")
		var data map[string]string
		if err := c.BindJSON(&data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON"})
	}
	alias_to_update := simpleorm.Alias{
		Name:  data["name"],
		Alias: data["alias"],
	}
	e.Table("alias").Where("id", id).Update(alias_to_update)
	
	})
	router.Run()
}
