package router

import (
	"embed"
	"getbind2api/common"
	logger "getbind2api/common/loggger"
	"getbind2api/middleware"
	"net/http"
	"strings"

	"github.com/gin-contrib/gzip"
	"github.com/gin-contrib/static"
	"github.com/gin-gonic/gin"
)

func SetWebRouter(router *gin.Engine, buildFS embed.FS) {
	// 尝试从嵌入的文件系统中读取前端首页文件
	indexPageData, err := buildFS.ReadFile("web/dist/index.html")
	if err != nil {
		logger.Errorf(nil, "Failed to read web index.html: %s", err.Error())
		logger.SysLog("Frontend will not be available!")
		return
	}

	router.Use(gzip.Gzip(gzip.DefaultCompression))
	//router.Use(middleware.GlobalWebRateLimit())
	router.Use(middleware.Cache())
	router.Use(static.Serve("/", common.EmbedFolder(buildFS, "web/dist")))

	// 处理所有非API路由，将它们重定向到前端应用
	router.NoRoute(func(c *gin.Context) {
		path := c.Request.URL.Path

		// 处理 API 请求，让它们返回404
		if strings.HasPrefix(path, "/v1") || strings.HasPrefix(path, "/api") {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "API endpoint not found",
				"path":  path,
				"code":  404,
			})
			return
		}

		// 处理静态资源请求
		if strings.Contains(path, ".") {
			// 可能是静态资源请求 (.js, .css, .png 等)
			c.Status(http.StatusNotFound)
			return
		}

		// 所有其他请求都返回前端入口页面，让前端路由处理
		c.Header("Cache-Control", "no-cache")
		c.Data(http.StatusOK, "text/html; charset=utf-8", indexPageData)
	})
}
