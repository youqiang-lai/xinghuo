package service

import (
	"fmt"
	"testing"
)

func TestParseRows(t *testing.T) {
	rows := [][]string{
		{"序号", "校区", "月份", "年级", "99元咨询课", "是否已上门", "是否签单", "金额"},
		{"1", "北京校区", "2024-01", "一年级", "是", "是", "是", "10000"},
		{"2", "北京校区", "2024-01", "二年级", "是", "否", "否", "5000"},
		{"3", "上海校区", "2024-01", "一年级", "是", "是", "是", "15000"},
		{"4", "上海校区", "2024-02", "三年级", "否", "是", "是", "20000"},
		{"5", "北京校区", "2024-02", "二年级", "是", "是", "否", "8000"},
		{"6", "广州校区", "2024-03", "四年级", "是", "否", "否", "3000"},
	}

	result := CalculateStats(parseRows(rows))

	fmt.Printf("Summary: Records=%d Visit=%d VisitAmount=%d AmountCount=%d Sign=%d TotalAmount=%.0f\n",
		result.Summary.TotalRecords, result.Summary.TotalVisit, 0, 0,
		result.Summary.TotalSign, result.Summary.TotalAmount)
	fmt.Printf("VisitRate(已上门且金额>0/已上门)=%.1f%% SignRate(金额>0/总数)=%.1f%%\n",
		result.Summary.OverallVisitRate, result.Summary.OverallSignRate)

	for _, cs := range result.CampusStats {
		fmt.Printf("Campus:%s total=%d visit=%d visitAmt=%d amtCnt=%d sign=%d amt=%.0f vr=%.1f%% sr=%.1f%%\n",
			cs.Campus, cs.TotalCount, cs.VisitCount, cs.VisitAmountCount, cs.AmountCount,
			cs.SignCount, cs.TotalAmount, cs.VisitRate, cs.SignRate)
	}

	if result.Summary.TotalRecords != 6 {
		t.Errorf("TotalRecords: want 6, got %d", result.Summary.TotalRecords)
	}
	if result.Summary.TotalVisit != 4 {
		t.Errorf("TotalVisit: want 4, got %d", result.Summary.TotalVisit)
	}
	// 4条已上门且全部金额>0 → VisitRate=100%
	if result.Summary.OverallVisitRate != 100.0 {
		t.Errorf("OverallVisitRate: want 100%%, got %.1f%%", result.Summary.OverallVisitRate)
	}
	// 6条全部金额>0 → SignRate=100%
	if result.Summary.OverallSignRate != 100.0 {
		t.Errorf("OverallSignRate: want 100%%, got %.1f%%", result.Summary.OverallSignRate)
	}
	if result.Summary.TotalSign != 3 {
		t.Errorf("TotalSign: want 3, got %d", result.Summary.TotalSign)
	}
}

func TestParseRowsColumnOrder(t *testing.T) {
	rows := [][]string{
		{"备注", "是否签单", "金额", "序号", "是否已上门", "月份", "99元咨询课", "校区", "年级"},
		{"", "是", "10000", "1", "是", "2024-01", "是", "北京校区", "一年级"},
		{"", "否", "5000", "2", "否", "2024-01", "是", "北京校区", "二年级"},
		{"", "是", "15000", "3", "是", "2024-01", "是", "上海校区", "一年级"},
		{"", "是", "20000", "4", "是", "2024-02", "否", "上海校区", "三年级"},
	}

	result := CalculateStats(parseRows(rows))

	fmt.Printf("ColOrder: Records=%d Visit=%d Sign=%d Amount=%.0f VisitRate=%.1f%% SignRate=%.1f%%\n",
		result.Summary.TotalRecords, result.Summary.TotalVisit, result.Summary.TotalSign,
		result.Summary.TotalAmount, result.Summary.OverallVisitRate, result.Summary.OverallSignRate)

	if result.Summary.TotalRecords != 4 {
		t.Errorf("TotalRecords: want 4, got %d", result.Summary.TotalRecords)
	}
	if result.Summary.TotalVisit != 3 {
		t.Errorf("TotalVisit: want 3, got %d", result.Summary.TotalVisit)
	}
	// 3条已上门且全部金额>0 → 100%
	if result.Summary.OverallVisitRate != 100.0 {
		t.Errorf("OverallVisitRate: want 100%%, got %.1f%%", result.Summary.OverallVisitRate)
	}
	// 4条全部金额>0 → 100%
	if result.Summary.OverallSignRate != 100.0 {
		t.Errorf("OverallSignRate: want 100%%, got %.1f%%", result.Summary.OverallSignRate)
	}
	if result.Summary.TotalAmount != 50000.0 {
		t.Errorf("TotalAmount: want 50000, got %.2f", result.Summary.TotalAmount)
	}
}
