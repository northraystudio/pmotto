package main

import (
	"context"
	"log"

	"pmotto/api/internal/config"
	"pmotto/api/internal/di"
	"pmotto/api/internal/usecase"

	"github.com/gin-gonic/gin"
)

func main() {
	c, err := di.BuildContainer()
	if err != nil {
		log.Fatalf("DIコンテナの構築に失敗: %v", err)
	}
	// 解決後の設定値を先に出す。DB 接続より前に出すことで、接続に失敗した場合でも
	// 「どの接続先を見にいったか」がログに残る。秘匿値は Summary 側でマスクされる。
	if err := c.Invoke(func(cfg config.Config) {
		log.Printf("設定値（環境変数の解決結果）:\n%s", cfg.Summary())
	}); err != nil {
		log.Fatalf("設定の読み込みに失敗: %v", err)
	}
	// 起動時に期限切れトークンを一度掃除する（テーブルの無限増加を防ぐ低頻度メンテナンス）。
	// 失敗しても起動は継続する（クリーンアップはベストエフォート）。
	if err := c.Invoke(func(auth *usecase.AuthUsecase) {
		refN, setN, err := auth.CleanupExpiredTokens(context.Background())
		if err != nil {
			log.Printf("期限切れトークンのクリーンアップに失敗（起動は継続）: %v", err)
			return
		}
		log.Printf("期限切れトークンを削除しました: refresh=%d, set=%d", refN, setN)
	}); err != nil {
		log.Fatalf("トークンクリーンアップの実行に失敗: %v", err)
	}
	if err := c.Invoke(func(e *gin.Engine, cfg config.Config) error {
		log.Printf("pmotto API を :%s で起動します", cfg.Port)
		return e.Run(":" + cfg.Port)
	}); err != nil {
		log.Fatalf("サーバー起動に失敗: %v", err)
	}
}
