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

func TestParseRowsColumnOrder(t *testing.T) {
	// 模拟打乱列顺序+额外无关列的Excel数据，验证按表头匹配
	rows := [][]string{
		{"备注", "是否签单", "金额", "序号", "是否上门", "月份", "99元咨询课", "校区", "年级"},
		{"", "是", "10000", "1", "是", "2024-01", "是", "北京校区", "一年级"},
		{"", "否", "5000", "2", "否", "2024-01", "是", "北京校区", "二年级"},
		{"", "是", "15000", "3", "是", "2024-01", "是", "上海校区", "一年级"},
		{"", "是", "20000", "4", "是", "2024-02", "否", "上海校区", "三年级"},
	}

	result := CalculateStats(parseRows(rows))

	fmt.Printf("ColumnOrder Test - TotalRecords=%d, TotalVisit=%d, TotalSign=%d, TotalAmount=%.2f\n",
		result.Summary.TotalRecords, result.Summary.TotalVisit, result.Summary.TotalSign, result.Summary.TotalAmount)

	if result.Summary.TotalRecords != 4 {
		t.Errorf("Expected 4 records, got %d", result.Summary.TotalRecords)
	}
	if result.Summary.TotalVisit != 3 {
		t.Errorf("Expected 3 visits, got %d", result.Summary.TotalVisit)
	}
	if result.Summary.TotalSign != 3 {
		t.Errorf("Expected 3 signs, got %d", result.Summary.TotalSign)
	}
	if result.Summary.TotalAmount != 50000.0 {
		t.Errorf("Expected 50000 total amount, got %.2f", result.Summary.TotalAmount)
	}
}
