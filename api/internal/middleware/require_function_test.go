package middleware

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// stubFuncs は保持する function コードを差し替えられる FunctionResolver。
// 既存の fakeFuncs は常に nil を返すため、権限あり/なしを撃ち分けられない。
type stubFuncs struct {
	codes []string
	err   error
}

func (s stubFuncs) FunctionsByUserID(context.Context, int) ([]string, error) {
	return s.codes, s.err
}

// TestRequireFunction は ACL の境界を固定する。
// エグゼクティブレポートのように「権限を持つロールだけが読める」エンドポイントで、
// 権限のないユーザーに 403 を返すことがリグレッションで壊れないようにする。
func TestRequireFunction(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cases := []struct {
		name     string
		required string
		funcs    stubFuncs
		want     int
	}{
		{
			name:     "必要な権限を持つなら通過",
			required: "view_executive_report",
			funcs:    stubFuncs{codes: []string{"view_dashboard", "view_executive_report"}},
			want:     http.StatusOK,
		},
		{
			name:     "権限を1つも持たないなら403",
			required: "view_executive_report",
			funcs:    stubFuncs{codes: nil},
			want:     http.StatusForbidden,
		},
		{
			name:     "他の権限は持つが当該権限がないなら403",
			required: "view_executive_report",
			funcs:    stubFuncs{codes: []string{"view_dashboard", "view_project_report"}},
			want:     http.StatusForbidden,
		},
		{
			name:     "前方一致する別コードでは通さない",
			required: "view_executive_report",
			funcs:    stubFuncs{codes: []string{"view_executive_report_detail"}},
			want:     http.StatusForbidden,
		},
		{
			name:     "権限解決に失敗したら500（権限なしとして通さない）",
			required: "view_executive_report",
			funcs:    stubFuncs{err: errors.New("db down")},
			want:     http.StatusInternalServerError,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			mw := New(fakeVerifier{id: 1}, tc.funcs, fakeActive{active: true}, fakeScope{})
			r := gin.New()
			r.GET("/x", mw.Authenticate(), mw.RequireFunction(tc.required),
				func(c *gin.Context) { c.Status(http.StatusOK) })

			req := httptest.NewRequest(http.MethodGet, "/x", nil)
			req.Header.Set("Authorization", "Bearer tok")
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			assert.Equal(t, tc.want, w.Code)
		})
	}
}
