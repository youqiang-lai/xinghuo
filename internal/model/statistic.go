package model

// ExcelRow 从Excel读取的原始数据行
type ExcelRow struct {
	Campus         string  `json:"campus"`          // 校区
	Month          string  `json:"month"`           // 月份
	Grade          string  `json:"grade"`           // 年级
	Consultation99 string  `json:"consultation_99"` // 99元咨询课
	VisitStatus    string  `json:"visit_status"`    // 是否已上门
	SignStatus     string  `json:"sign_status"`     // 是否签单
	Amount         float64 `json:"amount"`          // 金额
	FollowUp       string  `json:"follow_up"`       // 校区跟进结果
}

// CampusStat 校区维度统计
type CampusStat struct {
	Campus           string  `json:"campus"`
	TotalCount       int     `json:"total_count"`        // 总咨询数
	VisitCount       int     `json:"visit_count"`        // 已上门数
	VisitAmountCount int     `json:"visit_amount_count"` // 已上门且金额>0
	AmountCount      int     `json:"amount_count"`       // 金额>0的数量
	SignCount        int     `json:"sign_count"`         // 签单数（原始值保留）
	TotalAmount      float64 `json:"total_amount"`       // 总金额
	VisitRate        float64 `json:"visit_rate"`         // 上门转化率
	SignRate         float64 `json:"sign_rate"`          // 签单转化率
}

// MonthStat 月度维度统计
type MonthStat struct {
	Month            string  `json:"month"`
	TotalCount       int     `json:"total_count"`
	VisitCount       int     `json:"visit_count"`
	VisitAmountCount int     `json:"visit_amount_count"`
	AmountCount      int     `json:"amount_count"`
	SignCount        int     `json:"sign_count"`
	TotalAmount      float64 `json:"total_amount"`
	VisitRate        float64 `json:"visit_rate"`
	SignRate         float64 `json:"sign_rate"`
}

// GradeStat 年级维度统计
type GradeStat struct {
	Grade            string  `json:"grade"`
	TotalCount       int     `json:"total_count"`
	VisitCount       int     `json:"visit_count"`
	VisitAmountCount int     `json:"visit_amount_count"`
	AmountCount      int     `json:"amount_count"`
	SignCount        int     `json:"sign_count"`
	TotalAmount      float64 `json:"total_amount"`
	VisitRate        float64 `json:"visit_rate"`
	SignRate         float64 `json:"sign_rate"`
}

// SummaryStat 汇总统计
type SummaryStat struct {
	TotalRecords     int     `json:"total_records"`
	TotalVisit       int     `json:"total_visit"`
	TotalSign        int     `json:"total_sign"`
	TotalAmount      float64 `json:"total_amount"`
	OverallVisitRate float64 `json:"overall_visit_rate"` // 已上门且金额>0 / 已上门
	OverallSignRate  float64 `json:"overall_sign_rate"`  // 金额>0 / 总数
	FileCount        int     `json:"file_count"`
}

// DashboardData 仪表盘完整数据
type DashboardData struct {
	Summary     SummaryStat  `json:"summary"`
	CampusStats []CampusStat `json:"campus_stats"`
	MonthStats  []MonthStat  `json:"month_stats"`
	GradeStats  []GradeStat  `json:"grade_stats"`
	RawRows     []ExcelRow   `json:"raw_rows,omitempty"`
}
