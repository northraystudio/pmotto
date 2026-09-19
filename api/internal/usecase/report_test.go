package usecase

import (
	"context"
	"errors"
	"testing"

	"pmotto/api/internal/domain"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type stubReportRepo struct {
	reports []domain.ExecutiveReport
	err     error
	// gotLimit は usecase が渡した件数上限を記録する。
	gotLimit int
}

func (s *stubReportRepo) ListExecutive(_ context.Context, limit int) ([]domain.ExecutiveReport, error) {
	s.gotLimit = limit
	return s.reports, s.err
}

func TestReportUsecase_ListExecutive(t *testing.T) {
	comment := "ポートフォリオは概ね順調です。"

	cases := []struct {
		name    string
		repo    *stubReportRepo
		wantErr bool
		wantLen int
	}{
		{
			name: "レポートをそのまま返す",
			repo: &stubReportRepo{reports: []domain.ExecutiveReport{
				{
					ID:         2,
					ReportType: "weekly",
					AIComment:  &comment,
					Content: domain.ExecutiveReportContent{
						TotalProjects: 3,
						OnTrackCount:  2,
						HighRiskCount: 1,
					},
				},
				{ID: 1, ReportType: "weekly"},
			}},
			wantLen: 2,
		},
		{
			name:    "レポートが1件も無くても空で返す（エラーにしない）",
			repo:    &stubReportRepo{reports: nil},
			wantLen: 0,
		},
		{
			name:    "リポジトリのエラーは伝播する",
			repo:    &stubReportRepo{err: errors.New("db down")},
			wantErr: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			uc := NewReportUsecase(tc.repo)

			got, err := uc.ListExecutive(context.Background())
			if tc.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Len(t, got, tc.wantLen)
		})
	}
}

// TestReportUsecase_ListExecutive_件数上限を渡す は、全件取得によるレスポンス肥大を
// 防ぐ上限がリポジトリまで伝わっていることを固定する。
func TestReportUsecase_ListExecutive_件数上限を渡す(t *testing.T) {
	repo := &stubReportRepo{}
	uc := NewReportUsecase(repo)

	_, err := uc.ListExecutive(context.Background())

	require.NoError(t, err)
	assert.Equal(t, executiveReportLimit, repo.gotLimit)
	assert.Positive(t, repo.gotLimit, "上限が 0 だと1件も返らなくなる")
}
