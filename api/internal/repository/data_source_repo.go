package repository

import (
	"context"

	"pmo-agent/api/internal/domain"

	"gorm.io/gorm"
)

// DataSourceTypeRepo は進捗の取得元種別マスタを読み出す。
// 種別の追加はマイグレーション（データ投入）で行うため、書き込み経路は持たない。
type DataSourceTypeRepo struct{ db *gorm.DB }

func NewDataSourceTypeRepo(db *gorm.DB) *DataSourceTypeRepo { return &DataSourceTypeRepo{db: db} }

// ListActive は選択肢として提示できる種別のみを表示順で返す。
// 無効化済み（is_active=false）を除くのは、過去に設定されたプロジェクトの値は
// 保持したまま新規選択だけを止められるようにするため。
func (r *DataSourceTypeRepo) ListActive(ctx context.Context) ([]domain.DataSourceType, error) {
	var types []domain.DataSourceType
	err := r.db.WithContext(ctx).
		Where("is_active = ?", true).
		Order("sort_order, id").
		Find(&types).Error
	return types, err
}
