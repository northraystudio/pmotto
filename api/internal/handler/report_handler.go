package handler

import (
	"net/http"

	"pmo-agent/api/internal/usecase"

	"github.com/gin-gonic/gin"
)

// ReportHandler は n8n が生成したレポートの参照系を提供する。
type ReportHandler struct {
	uc *usecase.ReportUsecase
}

func NewReportHandler(uc *usecase.ReportUsecase) *ReportHandler {
	return &ReportHandler{uc: uc}
}

// Executive はエグゼクティブレポートを新しい順に返す。
func (h *ReportHandler) Executive(c *gin.Context) {
	reports, err := h.uc.ListExecutive(c.Request.Context())
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"reports": reports})
}
