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
	// 大屏统计页
	engine.GET("/data/index/api", func(c *gin.Context) {
		dataHandler(c, exeDir)
	})
	engine.GET("/data/index", func(c *gin.Context) {
		pageHandler(c, exeDir)
	})

	// 数据明细列表页（深色大屏风）
	engine.GET("/data/list/api", func(c *gin.Context) {
		listApiHandler(c, exeDir)
	})
	engine.GET("/data/list", func(c *gin.Context) {
		listPageHandler(c, exeDir)
	})

	// 数据明细列表页（白色卡片风）
	engine.GET("/data/list2/api", func(c *gin.Context) {
		list2ApiHandler(c, exeDir)
	})
	engine.GET("/data/list2", func(c *gin.Context) {
		list2PageHandler(c, exeDir)
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
	stats.RawRows = nil

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    stats,
	})
}

// pageHandler 返回数据统计大屏页面
func pageHandler(c *gin.Context, exeDir string) {
	rows, fileCount, err := service.ReadExcelDir(exeDir)
	errorMsg := ""
	stats := service.CalculateStats(rows)
	stats.Summary.FileCount = fileCount
	if err != nil && !os.IsNotExist(err) {
		errorMsg = err.Error()
	}
	absDir, _ := filepath.Abs(exeDir)
	c.HTML(http.StatusOK, "dashboard.html", gin.H{
		"title":    "星火数据统计大屏",
		"error":    errorMsg,
		"stats":    stats,
		"excelDir": filepath.Join(absDir, "excel"),
	})
}

// listApiHandler 返回列表数据JSON（支持筛选+分页）
func listApiHandler(c *gin.Context, exeDir string) {
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
	stats.RawRows = nil // API不返回明细

	// 筛选原始行数据
	filteredRows := service.FilterRows(rows,
		c.Query("campus"),
		c.Query("month"),
		c.Query("grade"),
		c.Query("keyword"),
	)

	total := len(filteredRows)
	page, pageSize, offset := service.ParsePagination(c.Query("page"), c.Query("pageSize"))
	pagedRows := service.PaginateRows(filteredRows, offset, pageSize)

	// 提取筛选用的下拉选项
	campuses, months, grades := service.ExtractFilterOptions(rows)

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data": gin.H{
			"summary":  stats.Summary,
			"rows":     pagedRows,
			"total":    total,
			"page":     page,
			"pageSize": pageSize,
			"filters": gin.H{
				"campuses": campuses,
				"months":   months,
				"grades":   grades,
			},
		},
	})
}

// listPageHandler 返回数据明细列表页面
func listPageHandler(c *gin.Context, exeDir string) {
	absDir, _ := filepath.Abs(exeDir)
	c.HTML(http.StatusOK, "list.html", gin.H{
		"title":    "星火咨询数据明细",
		"excelDir": filepath.Join(absDir, "excel"),
	})
}

// list2ApiHandler 返回按校区聚合的卡片数据
func list2ApiHandler(c *gin.Context, exeDir string) {
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
	stats.RawRows = nil

	// 按筛选条件过滤后，按校区分组
	filteredRows := service.FilterRows(rows,
		c.Query("campus"),
		c.Query("month"),
		c.Query("grade"),
		c.Query("keyword"),
	)

	campusSummaries := service.GroupByCampus(filteredRows)
	campuses, months, grades := service.ExtractFilterOptions(rows)

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data": gin.H{
			"summary":  stats.Summary,
			"campuses": campusSummaries,
			"filters": gin.H{
				"campuses": campuses,
				"months":   months,
				"grades":   grades,
			},
		},
	})
}

// list2PageHandler 返回白色卡片风格的数据明细列表页面
func list2PageHandler(c *gin.Context, exeDir string) {
	absDir, _ := filepath.Abs(exeDir)
	c.HTML(http.StatusOK, "list2.html", gin.H{
		"title":    "星火咨询数据明细",
		"excelDir": filepath.Join(absDir, "excel"),
	})
}
