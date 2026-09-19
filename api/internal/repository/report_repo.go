package repository

import (
	"context"

	"pmotto/api/internal/domain"

	"gorm.io/gorm"
)

// ReportRepo は n8n が書き込んだレポートを読み出す。書き込み経路は持たない。
type ReportRepo struct{ db *gorm.DB }

func NewReportRepo(db *gorm.DB) *ReportRepo { return &ReportRepo{db: db} }

// ListExecutive は新しい順にエグゼクティブレポートを返す。
// 同一日に複数の report_type がありうるため、日付が同じ場合は id の降順で安定させる。
func (r *ReportRepo) ListExecutive(ctx context.Context, limit int) ([]domain.ExecutiveReport, error) {
	var reports []domain.ExecutiveReport
	err := r.db.WithContext(ctx).
		Order("report_date DESC, id DESC").
		Limit(limit).
		Find(&reports).Error
	return reports, err
}
