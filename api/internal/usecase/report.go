package usecase

import (
	"context"
	"fmt"

	"pmo-agent/api/internal/domain"
)

// executiveReportLimit は一覧で返す件数の上限。週次生成なので数ヶ月分に相当する。
// 全件返すと content の JSON が積み上がってレスポンスが肥大するため上限を設ける。
const executiveReportLimit = 26

// ReportUsecase は n8n が生成したレポートの参照を担う。
// エグゼクティブレポートは全プロジェクト横断の集計であり、個別プロジェクトの
// スコープ制御にはなじまないため、閲覧可否は view_executive_report の有無
// （ACLミドルウェア）だけで判定する。
type ReportUsecase struct{ reports ReportRepository }

func NewReportUsecase(reports ReportRepository) *ReportUsecase {
	return &ReportUsecase{reports: reports}
}

// ListExecutive は新しい順にエグゼクティブレポートを返す。
func (uc *ReportUsecase) ListExecutive(ctx context.Context) ([]domain.ExecutiveReport, error) {
	reports, err := uc.reports.ListExecutive(ctx, executiveReportLimit)
	if err != nil {
		return nil, fmt.Errorf("usecase.Report.ListExecutive: %w", err)
	}
	return reports, nil
}
