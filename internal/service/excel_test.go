package service

import (
	"fmt"
	"testing"

	"xing-huo/internal/model"
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

	fmt.Printf("Records=%d Visit=%d Sign=%d TotalAmount=%.0f\n",
		result.Summary.TotalRecords, result.Summary.TotalVisit,
		result.Summary.TotalSign, result.Summary.TotalAmount)
	fmt.Printf("VisitRate(已上门/总数)=%.1f%% SignRate(金额>0/总数)=%.1f%%\n",
		result.Summary.OverallVisitRate, result.Summary.OverallSignRate)

	for _, cs := range result.CampusStats {
		fmt.Printf("Campus:%s total=%d visit=%d amtCnt=%d sign=%d amt=%.0f vr=%.1f%% sr=%.1f%%\n",
			cs.Campus, cs.TotalCount, cs.VisitCount,
			cs.AmountCount, cs.SignCount, cs.TotalAmount, cs.VisitRate, cs.SignRate)
	}

	if result.Summary.TotalRecords != 6 {
		t.Errorf("TotalRecords: want 6, got %d", result.Summary.TotalRecords)
	}
	if result.Summary.TotalVisit != 4 {
		t.Errorf("TotalVisit: want 4, got %d", result.Summary.TotalVisit)
	}
	// 上门率 = 已上门 4 / 总数 6
	if want, got := 4.0/6.0*100, result.Summary.OverallVisitRate; got-want > 1e-9 || want-got > 1e-9 {
		t.Errorf("VisitRate: want %.4f%%, got %.4f%%", want, got)
	}
	// 6条全部金额>0 → SignRate=100%
	if result.Summary.OverallSignRate != 100.0 {
		t.Errorf("SignRate: want 100%%, got %.1f%%", result.Summary.OverallSignRate)
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
	// 上门率 = 已上门 3 / 总数 4
	if want, got := 75.0, result.Summary.OverallVisitRate; got != want {
		t.Errorf("VisitRate: want %.1f%%, got %.1f%%", want, got)
	}
	// 4条全部金额>0 → 100%
	if result.Summary.OverallSignRate != 100.0 {
		t.Errorf("SignRate: want 100%%, got %.1f%%", result.Summary.OverallSignRate)
	}
	if result.Summary.TotalAmount != 50000.0 {
		t.Errorf("TotalAmount: want 50000, got %.2f", result.Summary.TotalAmount)
	}
}

func TestMonthKey(t *testing.T) {
	cases := []struct {
		in   string
		want int
	}{
		{"2024-01", 202401},
		{"2024-1", 202401},
		{"2024/1", 202401},
		{"2024年1月", 202401},
		{"202401", 202401},
		{"2024-12", 202412},
		{"1月", 1},
		{"12", 12},
		{"", 0},
		{"未填", 0},
	}
	for _, c := range cases {
		if got := MonthKey(c.in); got != c.want {
			t.Errorf("MonthKey(%q): want %d, got %d", c.in, c.want, got)
		}
	}
}

func TestMonthRangeSort(t *testing.T) {
	rs := []model.ExcelRow{
		{Campus: "C", Month: "2024-2"},
		{Campus: "C", Month: "2024-10"},
		{Campus: "C", Month: "2024-1"},
		{Campus: "C", Month: "2023-12"},
	}
	_, months, _ := ExtractFilterOptions(rs)
	want := []string{"2023-12", "2024-1", "2024-2", "2024-10"}
	if len(months) != len(want) {
		t.Fatalf("months: want %v, got %v", want, months)
	}
	for i := range want {
		if months[i] != want[i] {
			t.Fatalf("months order: want %v, got %v", want, months)
		}
	}
}

func TestFilterRowsByMonthRange(t *testing.T) {
	rows := []model.ExcelRow{
		{Campus: "A", Month: "2024-01", VisitStatus: "是"},
		{Campus: "A", Month: "2024-02", VisitStatus: "是"},
		{Campus: "A", Month: "2024-03", VisitStatus: "否"},
		{Campus: "B", Month: "2024-04", VisitStatus: "是"},
		{Campus: "B", Month: "2024-05", VisitStatus: "否"},
	}

	// 1~3 月
	got := FilterRows(rows, RowFilter{MonthFrom: "2024-01", MonthTo: "2024-03"})
	if len(got) != 3 {
		t.Fatalf("range 2024-01~2024-03: want 3, got %d", len(got))
	}
	stats := CalculateStats(got)
	if stats.Summary.TotalRecords != 3 || stats.Summary.TotalVisit != 2 {
		t.Errorf("range stats: want total=3 visit=2, got total=%d visit=%d",
			stats.Summary.TotalRecords, stats.Summary.TotalVisit)
	}
	if want, got := 2.0/3.0*100, stats.Summary.OverallVisitRate; got-want > 1e-9 || want-got > 1e-9 {
		t.Errorf("range visit rate: want %.4f, got %.4f", want, got)
	}

	// 仅起始月份
	if got := FilterRows(rows, RowFilter{MonthFrom: "2024-04"}); len(got) != 2 {
		t.Errorf("from 2024-04: want 2, got %d", len(got))
	}
	// 仅结束月份
	if got := FilterRows(rows, RowFilter{MonthTo: "2024-02"}); len(got) != 2 {
		t.Errorf("to 2024-02: want 2, got %d", len(got))
	}
	// 区间 + 校区
	if got := FilterRows(rows, RowFilter{Campus: "A", MonthFrom: "2024-02", MonthTo: "2024-04"}); len(got) != 2 {
		t.Errorf("campus A + range: want 2, got %d", len(got))
	}
	// 无筛选
	if got := FilterRows(rows, RowFilter{}); len(got) != len(rows) {
		t.Errorf("empty filter: want %d, got %d", len(rows), len(got))
	}
}
