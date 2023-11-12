package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func main() {
	router := gin.Default()
	// 处理multipart forms提交文件时默认的内存限制是32 MiB
	// 可以通过下面的方式修改

	// router.MaxMultipartMemory = 8 << 20  // 8 MiB

	router.LoadHTMLGlob("a.html")
	router.GET("/upload", func(c *gin.Context) {
		c.HTML(http.StatusOK, "a.html", nil)
	})

	router.POST("/upload", func(c *gin.Context) {

		transport_type := c.PostForm("transport_type")
		orgin := c.PostForm("orgin")
		destination := c.PostForm("destination")

		c.JSON(http.StatusOK, gin.H{
			"status":         "posted",
			"message":        "success",
			"transport_type": transport_type,
			"orgin":          orgin,
			"destination":    destination,
		})
	})
	router.Run()
}
