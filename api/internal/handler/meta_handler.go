package handler

import (
	"net/http"

	"pmotto/api/internal/usecase"

	"github.com/gin-gonic/gin"
)

// MetaHandler はロール一覧・データソース種別などの参照系マスタを提供する
// （メンバー管理UIのロール選択、プロジェクト編集の取得元選択などで使う）。
type MetaHandler struct {
	roles       usecase.RoleRepository
	dataSources usecase.DataSourceTypeRepository
}

func NewMetaHandler(roles usecase.RoleRepository, dataSources usecase.DataSourceTypeRepository) *MetaHandler {
	return &MetaHandler{roles: roles, dataSources: dataSources}
}

func (h *MetaHandler) Roles(c *gin.Context) {
	roles, err := h.roles.List(c.Request.Context())
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"roles": roles})
}

// DataSourceTypes は選択可能な進捗の取得元種別を表示順で返す。
func (h *MetaHandler) DataSourceTypes(c *gin.Context) {
	types, err := h.dataSources.ListActive(c.Request.Context())
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data_source_types": types})
}
