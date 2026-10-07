package service

import (
	"fmt"
	"testing"
)

func TestParseRows(t *testing.T) {
	// 模拟Excel数据
	rows := [][]string{
		{"序号", "校区", "月份", "年级", "99元咨询课", "是否上门", "是否签单", "金额"},
		{"1", "北京校区", "2024-01", "一年级", "是", "是", "是", "10000"},
		{"2", "北京校区", "2024-01", "二年级", "是", "否", "否", "5000"},
		{"3", "上海校区", "2024-01", "一年级", "是", "是", "是", "15000"},
		{"4", "上海校区", "2024-02", "三年级", "否", "是", "是", "20000"},
		{"5", "北京校区", "2024-02", "二年级", "是", "是", "否", "8000"},
		{"6", "广州校区", "2024-03", "四年级", "是", "否", "否", "3000"},
	}

	result := CalculateStats(parseRows(rows))

	fmt.Printf("Summary: TotalRecords=%d, TotalVisit=%d, TotalSign=%d, TotalAmount=%.2f\n",
		result.Summary.TotalRecords, result.Summary.TotalVisit, result.Summary.TotalSign, result.Summary.TotalAmount)
	fmt.Printf("VisitRate=%.1f%%, SignRate=%.1f%%\n",
		result.Summary.OverallVisitRate, result.Summary.OverallSignRate)

	for _, cs := range result.CampusStats {
		fmt.Printf("Campus: %s, Count=%d, Visit=%d, Sign=%d, Amount=%.2f, VisitRate=%.1f%%, SignRate=%.1f%%\n",
			cs.Campus, cs.TotalCount, cs.VisitCount, cs.SignCount, cs.TotalAmount, cs.VisitRate, cs.SignRate)
	}

	// 验证
	if result.Summary.TotalRecords != 6 {
		t.Errorf("Expected 6 records, got %d", result.Summary.TotalRecords)
	}
	if result.Summary.TotalVisit != 4 {
		t.Errorf("Expected 4 visits, got %d", result.Summary.TotalVisit)
	}
	if result.Summary.TotalSign != 3 {
		t.Errorf("Expected 3 signs, got %d", result.Summary.TotalSign)
	}
}