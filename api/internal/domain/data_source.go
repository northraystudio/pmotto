package domain

import "time"

// DataSourceType は進捗の取得元種別マスタ。projects.source_type から code で参照される。
// 種別の追加をコード変更なしのデータ追加で行えるようにするためテーブルに切り出している。
type DataSourceType struct {
	ID   int    `json:"id" gorm:"primaryKey"`
	Code string `json:"code"`
	// Label は画面のセレクトに出す表示名。ValueLabel は source_value 入力欄のラベルで、
	// 種別によって求める値の意味が変わる（スプレッドシートID / Backlog プロジェクトキー）。
	Label      string    `json:"label"`
	ValueLabel string    `json:"value_label"`
	SortOrder  int       `json:"sort_order"`
	IsActive   bool      `json:"is_active"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

func (DataSourceType) TableName() string { return "data_source_types" }
