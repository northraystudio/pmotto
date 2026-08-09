package repository

import (
	"errors"
	"fmt"
	"testing"

	"pmo-agent/api/internal/domain"

	"github.com/go-sql-driver/mysql"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

func TestWrapConflict(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want error
	}{
		{
			name: "MySQLの重複キーエラー(1062)はErrConflictに写像する",
			err:  &mysql.MySQLError{Number: 1062, Message: "Duplicate entry"},
			want: domain.ErrConflict,
		},
		{
			name: "GORMのErrDuplicatedKeyもErrConflictに写像する",
			err:  gorm.ErrDuplicatedKey,
			want: domain.ErrConflict,
		},
		{
			name: "ラップされた重複キーエラーも検出する",
			err:  fmt.Errorf("insert: %w", &mysql.MySQLError{Number: 1062}),
			want: domain.ErrConflict,
		},
		{
			name: "FK違反(1451)は重複キーではないのでそのまま返す",
			err:  &mysql.MySQLError{Number: 1451, Message: "Cannot delete or update a parent row"},
			want: &mysql.MySQLError{Number: 1451, Message: "Cannot delete or update a parent row"},
		},
		{
			name: "無関係なエラーはそのまま返す",
			err:  gorm.ErrInvalidData,
			want: gorm.ErrInvalidData,
		},
		{
			name: "nilはnilのまま",
			err:  nil,
			want: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := wrapConflict(tt.err)
			if tt.want == nil {
				assert.NoError(t, got)
				return
			}
			assert.True(t, errors.Is(got, tt.want), "got=%v want=%v", got, tt.want)
		})
	}
}

func TestIsForeignKeyViolation(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "FK RESTRICT(1451)はtrue",
			err:  &mysql.MySQLError{Number: 1451, Message: "Cannot delete or update a parent row"},
			want: true,
		},
		{
			name: "ラップされたFK違反も検出する",
			err:  fmt.Errorf("delete: %w", &mysql.MySQLError{Number: 1451}),
			want: true,
		},
		{
			name: "重複キーエラー(1062)はfalse",
			err:  &mysql.MySQLError{Number: 1062},
			want: false,
		},
		{
			name: "無関係なエラーはfalse",
			err:  gorm.ErrInvalidData,
			want: false,
		},
		{
			name: "nilはfalse",
			err:  nil,
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, isForeignKeyViolation(tt.err))
		})
	}
}

// TestIsMissingReference は「参照先が存在しない値を書き込もうとした」ケースを
// 1451（親行が参照されている）と取り違えないことを固定する。両者は向きが逆で、
// 前者はクライアントの入力誤り（400）、後者は削除の競合（409）に写像される。
func TestIsMissingReference(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "参照先欠如(1452)を検出する",
			err:  &mysql.MySQLError{Number: 1452},
			want: true,
		},
		{
			name: "ラップされていても検出する",
			err:  fmt.Errorf("update: %w", &mysql.MySQLError{Number: 1452}),
			want: true,
		},
		{
			name: "向きが逆のFK違反(1451)はfalse",
			err:  &mysql.MySQLError{Number: 1451},
			want: false,
		},
		{
			name: "重複キーエラー(1062)はfalse",
			err:  &mysql.MySQLError{Number: 1062},
			want: false,
		},
		{
			name: "nilはfalse",
			err:  nil,
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, isMissingReference(tt.err))
		})
	}
}
