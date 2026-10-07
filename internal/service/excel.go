package service

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"xing-huo/internal/model"

	"github.com/xuri/excelize/v2"
)

// ReadExcelDir 读取exe所在目录下excel/文件夹中所有xlsx和xls文件的第一个sheet页
func ReadExcelDir(baseDir string) ([]model.ExcelRow, int, error) {
	excelDir := filepath.Join(baseDir, "excel")
	var allRows []model.ExcelRow
	fileCount := 0

	entries, err := os.ReadDir(excelDir)
	if err != nil {
		if os.IsNotExist(err) {
			// 目录不存在时创建
			_ = os.MkdirAll(excelDir, 0755)
			return nil, 0, fmt.Errorf("excel目录不存在，已在exe同级目录创建 excel/ 文件夹，请放入xlsx或xls文件后刷新")
		}
		return nil, 0, fmt.Errorf("读取excel目录失败: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		ext := strings.ToLower(filepath.Ext(entry.Name()))
		filePath := filepath.Join(excelDir, entry.Name())

		var rows []model.ExcelRow
		var readErr error

		switch ext {
		case ".xlsx":
			rows, readErr = readXlsx(filePath)
		case ".xls":
			rows, readErr = readXls(filePath)
		default:
			continue
		}

		if readErr != nil {
			fmt.Printf("警告：读取文件 %s 失败: %v\n", entry.Name(), readErr)
			continue
		}

		allRows = append(allRows, rows...)
		fileCount++
	}

	return allRows, fileCount, nil
}

// readXlsx 读取xlsx文件
func readXlsx(filePath string) ([]model.ExcelRow, error) {
	f, err := excelize.OpenFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("打开xlsx失败: %w", err)
	}
	defer f.Close()

	// 获取第一个sheet名
	sheetName := f.GetSheetName(0)
	if sheetName == "" {
		return nil, fmt.Errorf("未找到有效的sheet")
	}

	rows, err := f.GetRows(sheetName)
	if err != nil {
		return nil, fmt.Errorf("读取sheet失败: %w", err)
	}

	return parseRows(rows), nil
}

// readXls 读取xls文件（旧格式）
func readXls(filePath string) ([]model.ExcelRow, error) {
	f, err := excelize.OpenFile(filePath)
	if err != nil {
		// excelize 对 .xls 兼容性有限，尝试重命名为xlsx读取
		// 如果仍然失败，返回错误
		return nil, fmt.Errorf("xls文件请转换为xlsx格式后读取: %w", err)
	}
	defer f.Close()

	sheetName := f.GetSheetName(0)
	if sheetName == "" {
		return nil, fmt.Errorf("未找到有效的sheet")
	}

	rows, err := f.GetRows(sheetName)
	if err != nil {
		return nil, fmt.Errorf("读取sheet失败: %w", err)
	}

	return parseRows(rows), nil
}

// parseRows 解析行数据，跳过表头，提取B列(校区)、C列(月份)及后续统计列
// Excel列对应:
//
//	A(1)  B(2)校区  C(3)月份  D(4)年级  E(5)99元咨询课  F(6)是否上门  G(7)是否签单  H(8)金额
func parseRows(rows [][]string) []model.ExcelRow {
	if len(rows) < 2 {
		return nil
	}

	var result []model.ExcelRow

	// 跳过第一行（表头）
	for i := 1; i < len(rows); i++ {
		row := rows[i]
		excelRow := model.ExcelRow{}

		// B列 → index 1
		if len(row) > 1 {
			excelRow.Campus = strings.TrimSpace(row[1])
		}
		// C列 → index 2
		if len(row) > 2 {
			excelRow.Month = strings.TrimSpace(row[2])
		}
		// D列 → index 3 (年级)
		if len(row) > 3 {
			excelRow.Grade = strings.TrimSpace(row[3])
		}
		// E列 → index 4 (99元咨询课)
		if len(row) > 4 {
			excelRow.Consultation99 = strings.TrimSpace(row[4])
		}
		// F列 → index 5 (是否上门)
		if len(row) > 5 {
			excelRow.VisitStatus = strings.TrimSpace(row[5])
		}
		// G列 → index 6 (是否签单)
		if len(row) > 6 {
			excelRow.SignStatus = strings.TrimSpace(row[6])
		}
		// H列 → index 7 (金额)
		if len(row) > 7 {
			amount, _ := strconv.ParseFloat(strings.TrimSpace(strings.ReplaceAll(row[7], ",", "")), 64)
			excelRow.Amount = amount
		}

		// 跳过完全空行
		if excelRow.Campus == "" && excelRow.Month == "" && excelRow.Grade == "" {
			continue
		}

		result = append(result, excelRow)
	}

	return result
}

// CalculateStats 根据原始数据计算各维度统计
func CalculateStats(rows []model.ExcelRow) model.DashboardData {
	campusMap := make(map[string]*model.CampusStat)
	monthMap := make(map[string]*model.MonthStat)
	gradeMap := make(map[string]*model.GradeStat)

	totalRecords := 0
	totalVisit := 0
	totalSign := 0
	totalAmount := 0.0

	for _, row := range rows {
		totalRecords++

		// 判断是否上门
		isVisit := isYes(row.VisitStatus)
		if isVisit {
			totalVisit++
		}
		// 判断是否签单
		isSign := isYes(row.SignStatus)
		if isSign {
			totalSign++
		}
		totalAmount += row.Amount

		// 校区维度
		if row.Campus != "" {
			if _, ok := campusMap[row.Campus]; !ok {
				campusMap[row.Campus] = &model.CampusStat{Campus: row.Campus}
			}
			campusMap[row.Campus].TotalCount++
			campusMap[row.Campus].TotalAmount += row.Amount
			if isVisit {
				campusMap[row.Campus].VisitCount++
			}
			if isSign {
				campusMap[row.Campus].SignCount++
			}
		}

		// 月份维度
		if row.Month != "" {
			if _, ok := monthMap[row.Month]; !ok {
				monthMap[row.Month] = &model.MonthStat{Month: row.Month}
			}
			monthMap[row.Month].TotalCount++
			monthMap[row.Month].TotalAmount += row.Amount
			if isVisit {
				monthMap[row.Month].VisitCount++
			}
			if isSign {
				monthMap[row.Month].SignCount++
			}
		}

		// 年级维度
		if row.Grade != "" {
			if _, ok := gradeMap[row.Grade]; !ok {
				gradeMap[row.Grade] = &model.GradeStat{Grade: row.Grade}
			}
			gradeMap[row.Grade].TotalCount++
			gradeMap[row.Grade].TotalAmount += row.Amount
			if isVisit {
				gradeMap[row.Grade].VisitCount++
			}
			if isSign {
				gradeMap[row.Grade].SignCount++
			}
		}
	}

	// 计算比率
	var campusStats []model.CampusStat
	for _, v := range campusMap {
		if v.TotalCount > 0 {
			v.VisitRate = float64(v.VisitCount) / float64(v.TotalCount) * 100
			v.SignRate = float64(v.SignCount) / float64(v.TotalCount) * 100
		}
		campusStats = append(campusStats, *v)
	}

	var monthStats []model.MonthStat
	for _, v := range monthMap {
		if v.TotalCount > 0 {
			v.VisitRate = float64(v.VisitCount) / float64(v.TotalCount) * 100
			v.SignRate = float64(v.SignCount) / float64(v.TotalCount) * 100
		}
		monthStats = append(monthStats, *v)
	}

	var gradeStats []model.GradeStat
	for _, v := range gradeMap {
		if v.TotalCount > 0 {
			v.VisitRate = float64(v.VisitCount) / float64(v.TotalCount) * 100
			v.SignRate = float64(v.SignCount) / float64(v.TotalCount) * 100
		}
		gradeStats = append(gradeStats, *v)
	}

	overallVisitRate := 0.0
	overallSignRate := 0.0
	if totalRecords > 0 {
		overallVisitRate = float64(totalVisit) / float64(totalRecords) * 100
		overallSignRate = float64(totalSign) / float64(totalRecords) * 100
	}

	return model.DashboardData{
		Summary: model.SummaryStat{
			TotalRecords:     totalRecords,
			TotalVisit:       totalVisit,
			TotalSign:        totalSign,
			TotalAmount:      totalAmount,
			OverallVisitRate: overallVisitRate,
			OverallSignRate:  overallSignRate,
			FileCount:        0, // 由调用者设置
		},
		CampusStats: campusStats,
		MonthStats:  monthStats,
		GradeStats:  gradeStats,
		RawRows:     rows,
	}
}

// isYes 判断字符串是否为"是"或"✓"等确认标记
func isYes(s string) bool {
	s = strings.TrimSpace(strings.ToLower(s))
	return s == "是" || s == "yes" || s == "y" || s == "1" || s == "✓" || s == "true"
}