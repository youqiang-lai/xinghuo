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

// colMap 表头中文名 → 数据列索引的映射
type colMap struct {
	campusIdx         int // 校区
	monthIdx          int // 月份
	gradeIdx          int // 年级
	consultation99Idx int // 99元咨询课
	visitIdx          int // 是否上门
	signIdx           int // 是否签单
	amountIdx         int // 金额
}

// fieldKeywords 每个字段对应的表头匹配关键词（包含任一关键词即匹配）
var fieldKeywords = map[string][]string{
	"campus":         {"校区"},
	"month":          {"月份", "日期", "时间"},
	"grade":          {"年级"},
	"consultation99": {"99", "咨询课"},
	"visit":          {"上门"},
	"sign":           {"签单"},
	"amount":         {"金额", "收入"},
}

// buildColMap 根据表头行构建列索引映射
func buildColMap(header []string) colMap {
	cm := colMap{
		campusIdx:         -1,
		monthIdx:          -1,
		gradeIdx:          -1,
		consultation99Idx: -1,
		visitIdx:          -1,
		signIdx:           -1,
		amountIdx:         -1,
	}
	for i, h := range header {
		h = strings.TrimSpace(h)
		if h == "" {
			continue
		}
		for field, keywords := range fieldKeywords {
			for _, kw := range keywords {
				if strings.Contains(h, kw) {
					switch field {
					case "campus":
						// 校区：如果已匹配到更精准的（仅"校区"不含"学校"歧义），不覆盖；
						// 但如果标题本身包含"学校"也行，优先取包含"校区"的
						if cm.campusIdx == -1 || strings.Contains(h, "校区") {
							cm.campusIdx = i
						}
					case "month":
						if cm.monthIdx == -1 {
							cm.monthIdx = i
						}
					case "grade":
						if cm.gradeIdx == -1 {
							cm.gradeIdx = i
						}
					case "consultation99":
						if cm.consultation99Idx == -1 {
							cm.consultation99Idx = i
						}
					case "visit":
						if cm.visitIdx == -1 {
							cm.visitIdx = i
						}
					case "sign":
						// 签单：避免把"是否上门"列也匹配为签单（前序逻辑上门先匹配），
						// 但如果表头包含"签单"更精准，优先使用包含"签单"的列
						if cm.signIdx == -1 || strings.Contains(h, "签单") {
							cm.signIdx = i
						}
					case "amount":
						if cm.amountIdx == -1 {
							cm.amountIdx = i
						}
					}
					break // 该字段匹配到了就跳出 keywords 循环
				}
			}
		}
	}
	return cm
}

// getStr 安全获取行中第i列的值
func getStr(row []string, i int) string {
	if i < 0 || i >= len(row) {
		return ""
	}
	return strings.TrimSpace(row[i])
}

// parseRows 根据表头行动态匹配列，解析后续数据行
func parseRows(rows [][]string) []model.ExcelRow {
	if len(rows) < 2 {
		return nil
	}

	// 第一行是表头
	cm := buildColMap(rows[0])

	var result []model.ExcelRow

	for i := 1; i < len(rows); i++ {
		row := rows[i]
		excelRow := model.ExcelRow{
			Campus:         getStr(row, cm.campusIdx),
			Month:          getStr(row, cm.monthIdx),
			Grade:          getStr(row, cm.gradeIdx),
			Consultation99: getStr(row, cm.consultation99Idx),
			VisitStatus:    getStr(row, cm.visitIdx),
			SignStatus:     getStr(row, cm.signIdx),
		}

		// 金额
		if cm.amountIdx >= 0 && cm.amountIdx < len(row) {
			amount, _ := strconv.ParseFloat(strings.TrimSpace(strings.ReplaceAll(row[cm.amountIdx], ",", "")), 64)
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

// ==================== 筛选与分页 ====================

// FilterRows 按校区、月份、年级、关键词筛选
func FilterRows(rows []model.ExcelRow, campus, month, grade, keyword string) []model.ExcelRow {
	if campus == "" && month == "" && grade == "" && keyword == "" {
		return rows
	}

	var result []model.ExcelRow
	kw := strings.ToLower(strings.TrimSpace(keyword))

	for _, row := range rows {
		if campus != "" && row.Campus != campus {
			continue
		}
		if month != "" && row.Month != month {
			continue
		}
		if grade != "" && row.Grade != grade {
			continue
		}
		if kw != "" {
			if !strings.Contains(strings.ToLower(row.Campus), kw) &&
				!strings.Contains(strings.ToLower(row.Month), kw) &&
				!strings.Contains(strings.ToLower(row.Grade), kw) {
				continue
			}
		}
		result = append(result, row)
	}
	return result
}

// ParsePagination 解析分页参数
func ParsePagination(pageStr, pageSizeStr string) (page, pageSize, offset int) {
	page = 1
	pageSize = 20
	if v, err := strconv.Atoi(pageStr); err == nil && v > 0 {
		page = v
	}
	if v, err := strconv.Atoi(pageSizeStr); err == nil && v > 0 && v <= 200 {
		pageSize = v
	}
	offset = (page - 1) * pageSize
	return
}

// PaginateRows 分页截取
func PaginateRows(rows []model.ExcelRow, offset, pageSize int) []model.ExcelRow {
	if offset >= len(rows) {
		return nil
	}
	end := offset + pageSize
	if end > len(rows) {
		end = len(rows)
	}
	return rows[offset:end]
}

// ExtractFilterOptions 提取筛选下拉选项（去重排序）
func ExtractFilterOptions(rows []model.ExcelRow) (campuses, months, grades []string) {
	campusSet := make(map[string]struct{})
	monthSet := make(map[string]struct{})
	gradeSet := make(map[string]struct{})

	for _, row := range rows {
		if row.Campus != "" {
			campusSet[row.Campus] = struct{}{}
		}
		if row.Month != "" {
			monthSet[row.Month] = struct{}{}
		}
		if row.Grade != "" {
			gradeSet[row.Grade] = struct{}{}
		}
	}

	for k := range campusSet {
		campuses = append(campuses, k)
	}
	for k := range monthSet {
		months = append(months, k)
	}
	for k := range gradeSet {
		grades = append(grades, k)
	}

	sortStrings(campuses)
	sortStrings(months)
	sortStrings(grades)
	return
}

func sortStrings(arr []string) {
	for i := 0; i < len(arr); i++ {
		for j := i + 1; j < len(arr); j++ {
			if arr[i] > arr[j] {
				arr[i], arr[j] = arr[j], arr[i]
			}
		}
	}
}
