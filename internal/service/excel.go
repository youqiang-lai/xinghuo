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
	followUpIdx       int // 校区跟进结果
}

// fieldKeywords 每个字段对应的表头匹配关键词
// campus、followUp 使用全字匹配（h == kw），其余使用包含匹配（Contains）
var fieldKeywords = map[string][]string{
	"campus":         {"校区"},
	"month":          {"月份"},
	"grade":          {"年级"},
	"consultation99": {"99", "咨询课"},
	"visit":          {"是否已上门"},
	"sign":           {"是否签单"},
	"amount":         {"金额", "收入"},
	"followUp":       {"校区跟进结果"},
}

// matchField 判断表头单元格 h 是否匹配字段 field
func matchField(field string, h string, kw string) bool {
	switch field {
	case "campus", "followUp":
		// 全字匹配：避免"校区跟进结果"误匹配到"校区"
		return h == kw
	default:
		return strings.Contains(h, kw)
	}
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
		followUpIdx:       -1,
	}
	for i, h := range header {
		h = strings.TrimSpace(h)
		if h == "" {
			continue
		}
		for field, keywords := range fieldKeywords {
			for _, kw := range keywords {
				if matchField(field, h, kw) {
					switch field {
					case "campus":
						if cm.campusIdx == -1 {
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
						if cm.signIdx == -1 {
							cm.signIdx = i
						}
					case "amount":
						if cm.amountIdx == -1 {
							cm.amountIdx = i
						}
					case "followUp":
						if cm.followUpIdx == -1 {
							cm.followUpIdx = i
						}
					}
					break
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
			FollowUp:       getStr(row, cm.followUpIdx),
		}

		if cm.amountIdx >= 0 && cm.amountIdx < len(row) {
			amount, _ := strconv.ParseFloat(strings.TrimSpace(strings.ReplaceAll(row[cm.amountIdx], ",", "")), 64)
			excelRow.Amount = amount
		}

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

		isVisit := isYes(row.VisitStatus)
		if isVisit {
			totalVisit++
		}
		isSign := isYes(row.SignStatus)
		if isSign {
			totalSign++
		}
		totalAmount += row.Amount

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
			FileCount:        0,
		},
		CampusStats: campusStats,
		MonthStats:  monthStats,
		GradeStats:  gradeStats,
		RawRows:     rows,
	}
}

// isYes 判断字符串是否为"是"
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
				!strings.Contains(strings.ToLower(row.Grade), kw) &&
				!strings.Contains(strings.ToLower(row.FollowUp), kw) {
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
