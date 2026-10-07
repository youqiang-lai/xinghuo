package main

import (
	"embed"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"xing-huo/internal/handler"

	"github.com/gin-gonic/gin"
)

//go:embed internal/handler/templates/*
var templatesFS embed.FS

func main() {
	// 获取exe所在目录
	exeDir := getExeDir()

	// 设置 gin 模式
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(gin.LoggerWithWriter(gin.DefaultWriter, "/data/index/api", "/data/list/api"), gin.Recovery())

	// 加载HTML模板（从embed或外部的templates目录）
	loadTemplates(router)

	// 注册数据统计接口
	handler.DataApiGroup(router, exeDir)

	// 首页重定向
	router.GET("/", func(c *gin.Context) {
		c.Redirect(http.StatusMovedPermanently, "/data/index")
	})

	// 静态资源（如果存在static目录）
	staticDir := filepath.Join(exeDir, "static")
	if _, err := os.Stat(staticDir); err == nil {
		router.Static("/static", staticDir)
	}

	// 确保excel目录存在
	excelDir := filepath.Join(exeDir, "excel")
	_ = os.MkdirAll(excelDir, 0755)

	port := "80"
	if p := os.Getenv("XINGHUO_PORT"); p != "" {
		port = p
	}

	fmt.Println(strings.Repeat("=", 60))
	fmt.Println("  数据统计系统  XingHuo Data Analytics")
	fmt.Println(strings.Repeat("=", 60))
	fmt.Printf("  服务地址: http://localhost:%s/data/index\n", port)
	fmt.Printf("  API接口:  http://localhost:%s/data/index/api\n", port)
	fmt.Printf("  Excel目录: %s\n", excelDir)
	fmt.Println(strings.Repeat("=", 60))
	fmt.Println("  按 Ctrl+C 退出服务")
	fmt.Println(strings.Repeat("=", 60))

	if err := router.Run(":" + port); err != nil {
		fmt.Fprintf(os.Stderr, "启动服务失败: %v\n", err)
		os.Exit(1)
	}
}

// getExeDir 获取可执行文件所在目录
func getExeDir() string {
	exePath, err := os.Executable()
	if err != nil {
		return "."
	}
	exePath, err = filepath.EvalSymlinks(exePath)
	if err != nil {
		return "."
	}
	return filepath.Dir(exePath)
}

// loadTemplates 加载HTML模板
func loadTemplates(router *gin.Engine) {
	// 优先尝试从embed加载
	tmpl, err := template.ParseFS(templatesFS, "internal/handler/templates/*.html")
	if err != nil {
		// embed 加载失败，尝试从外部目录加载（开发模式）
		exeDir := getExeDir()
		externalPath := filepath.Join(exeDir, "internal", "handler", "templates", "*.html")
		tmpl, err = template.ParseGlob(externalPath)
		if err != nil {
			// 尝试从当前目录加载
			tmpl, err = template.ParseGlob("internal/handler/templates/*.html")
			if err != nil {
				fmt.Fprintf(os.Stderr, "警告: 无法加载HTML模板: %v\n", err)
				return
			}
		}
	}
	router.SetHTMLTemplate(tmpl)

	// 验证模板已加载
	if tmpl != nil {
		tmplNames := []string{}
		for _, t := range tmpl.Templates() {
			tmplNames = append(tmplNames, t.Name())
		}
		fmt.Printf("  已加载模板: %v\n", tmplNames)
	}
}
