package handler

import (
	"net/http"
	"os"
	"path/filepath"

	"xing-huo/internal/service"

	"github.com/gin-gonic/gin"
)

// DataApiGroup 注册数据统计相关接口
func DataApiGroup(engine *gin.Engine, exeDir string) {
	// API接口：返回统计数据JSON
	engine.GET("/data/index/api", func(c *gin.Context) {
		dataHandler(c, exeDir)
	})

	// 页面接口：返回前端页面
	engine.GET("/data/index", func(c *gin.Context) {
		pageHandler(c, exeDir)
	})
}

// dataHandler 返回统计数据JSON
func dataHandler(c *gin.Context, exeDir string) {
	rows, fileCount, err := service.ReadExcelDir(exeDir)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    -1,
			"message": err.Error(),
			"data":    nil,
		})
		return
	}

	stats := service.CalculateStats(rows)
	stats.Summary.FileCount = fileCount

	// 尝试读取excel日期（取所有文件中最早的年份）
	stats.RawRows = nil // 默认不返回明细

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    stats,
	})
}

// pageHandler 返回数据统计页面
func pageHandler(c *gin.Context, exeDir string) {
	// 先尝试统计，如果有错误也展示页面(带错误信息)
	rows, fileCount, err := service.ReadExcelDir(exeDir)
	errorMsg := ""
	stats := service.CalculateStats(rows)
	stats.Summary.FileCount = fileCount

	if err != nil {
		// 目录不存在或没有文件时，展示空数据
		if !os.IsNotExist(err) {
			errorMsg = err.Error()
		}
		// 仍然展示页面，只是数据为空
	}

	// 将exeDir转为绝对路径
	absDir, _ := filepath.Abs(exeDir)

	c.HTML(http.StatusOK, "dashboard.html", gin.H{
		"title":    "星火数据统计大屏",
		"error":    errorMsg,
		"stats":    stats,
		"excelDir": filepath.Join(absDir, "excel"),
	})
}
