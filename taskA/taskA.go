package main

import (
	"fmt"
	"html/template"
	"net/http"
	"taskA/routing"
	"taskA/simpleorm"

	"github.com/gin-gonic/gin"
)

type route struct {
	id          int
	route       string
	origin      string
	destination string
}
type alias struct {
	id    int
	name  string
	alias string
}

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

	router.POST("/index", func(c *gin.Context) {
		transport_type := c.PostForm("transport_type")
		orgin := c.PostForm("orgin")
		destination := c.PostForm("destination")
		route_preference := c.PostForm("route_preference")

		riding_type := c.PostForm("riding_type")
		transit_preference := c.PostForm("transit_preference")

		if transport_type == "driving" {
			tatics_map := map[string]int{
				"fastest":           0,
				"no_highways":       1,
				"avoid_traffic":     2,
				"shortest_distance": 3,
				"minimize_cost":     4,
				"prefer_main_roads": 5,
			}
			dynamicHTML := routing.Directionlite_driving(orgin, destination, tatics_map[route_preference])
			c.HTML(http.StatusOK, "index/index.html", gin.H{
				"DynamicHTML": template.HTML(dynamicHTML),
			})
		} else if transport_type == "riding" {
			c.JSON(http.StatusOK, gin.H{
				"status":         "posted",
				"message":        "success",
				"transport_type": transport_type,
				"ridingOptions":  riding_type,
				"orgin":          orgin,
				"destination":    destination,
			})
		} else if transport_type == "transit" {
			c.JSON(http.StatusOK, gin.H{
				"status":         "posted",
				"message":        "success",
				"transport_type": transport_type,
				"transitOptions": transit_preference,
				"orgin":          orgin,
				"destination":    destination,
			})
		} else {
			c.JSON(http.StatusOK, gin.H{
				"status":         "posted",
				"message":        "success",
				"transport_type": transport_type,
				"orgin":          orgin,
				"destination":    destination,
			})
		}
	})
	router.POST("/route", func(c *gin.Context) {
		route_name := c.PostForm("route_name")
		route_orgin := c.PostForm("route_orgin")
		route_destination := c.PostForm("route_destination")

		c.JSON(http.StatusOK, gin.H{
			"status":            "posted",
			"message":           "success",
			"route_name":        route_name,
			"route_orgin":       route_orgin,
			"route_destination": route_destination,
		})
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
		out, err := e.Table("alias").Select()
		if err != nil {
			fmt.Println(err)
		}
		c.HTML(http.StatusOK, "alias/alias.html", gin.H{
			"StringDictArray": out,
		})
	})
	router.DELETE("/alias", func(c *gin.Context) {
		println("delete")
		id := c.Query("id")
		e.Table("alias").Where("id", id).Delete()

	})
	router.Run()
}
