package domain

import "time"

// ExecutiveReport は全プロジェクト横断のエグゼクティブレポート。
// 生成と AI コメント付与は n8n の週次ワークフローが行い、API は読み取りのみ担う。
// そのため書き込み用のメソッドやバリデーションはこの層に持たない。
type ExecutiveReport struct {
	ID         int       `json:"id" gorm:"primaryKey"`
	ReportDate time.Time `json:"report_date" gorm:"type:date"`
	ReportType string    `json:"report_type"`
	// Content は n8n が組み立てた JSON。AIComment は AI 生成ステップが後から
	// UPDATE で埋めるため、生成直後は NULL になりうる。
	Content   ExecutiveReportContent `json:"content" gorm:"serializer:json"`
	AIComment *string                `json:"ai_comment" gorm:"column:ai_comment"`
	CreatedAt time.Time              `json:"created_at"`
}

func (ExecutiveReport) TableName() string { return "executive_reports" }

// ExecutiveReportContent は executive_reports.content の JSON 構造。
// 生成元は n8n weekly_workflow.json の「エグゼクティブレポート生成」ノード。
// 未知のキーは encoding/json が黙って読み飛ばすため、ワークフロー側が項目を
// 増やしても API は壊れない。
type ExecutiveReportContent struct {
	ReportWeek       string                    `json:"report_week"`
	TotalProjects    int                       `json:"total_projects"`
	OnTrackCount     int                       `json:"on_track_count"`
	HighRiskCount    int                       `json:"high_risk_count"`
	AverageProgress  float64                   `json:"average_progress"`
	Projects         []ExecutiveProjectSummary `json:"projects"`
	HighRiskProjects []string                  `json:"high_risk_projects"`
	ExecutiveSummary string                    `json:"executive_summary"`
}

// ExecutiveProjectSummary は content.projects の要素（プロジェクト1件の週次サマリ）。
type ExecutiveProjectSummary struct {
	ProjectID     int     `json:"project_id"`
	ProjectName   string  `json:"project_name"`
	AvgProgress   float64 `json:"avg_progress"`
	ProgressTrend float64 `json:"progress_trend"`
	HighRiskCount int     `json:"high_risk_count"`
	// RiskCounts はリスク種別ごとの件数。キーはワークフロー側の定義に依存するため
	// 固定の構造体にせずマップで受ける。
	RiskCounts map[string]int `json:"risk_counts"`
}
