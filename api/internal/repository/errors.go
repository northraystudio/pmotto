package repository

import (
	"errors"

	"pmo-agent/api/internal/domain"

	"github.com/go-sql-driver/mysql"
	"gorm.io/gorm"
)

// mysqlDuplicateEntry は MySQL の重複キーエラーコード（ER_DUP_ENTRY）。
const mysqlDuplicateEntry = 1062

// mysqlRowIsReferenced は親行を参照する子行が存在するため削除・更新を拒否した
// FK RESTRICT のエラーコード（ER_ROW_IS_REFERENCED_2）。
const mysqlRowIsReferenced = 1451

// wrapNotFound は GORM の未検出エラーをドメインエラーへ写像する。
func wrapNotFound(err error) error {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return domain.ErrNotFound
	}
	return err
}

// wrapConflict は一意制約違反（採番の同時実行やコード・メールの重複）を
// domain.ErrConflict へ写像する。handler はこれを 409 として返す。
func wrapConflict(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, gorm.ErrDuplicatedKey) {
		return domain.ErrConflict
	}
	var me *mysql.MySQLError
	if errors.As(err, &me) && me.Number == mysqlDuplicateEntry {
		return domain.ErrConflict
	}
	return err
}

// mysqlNoReferencedRow は参照先の親行が存在しないため INSERT/UPDATE を拒否した
// エラーコード（ER_NO_REFERENCED_ROW_2）。1451 とは向きが逆で、こちらは
// 「存在しない値を書こうとした」＝クライアントの入力誤りにあたる。
const mysqlNoReferencedRow = 1452

// isMissingReference は FK 先に存在しない値を書き込もうとしたことを判定する。
// 呼び出し側は domain.ErrValidation へ写像し、500 ではなく 400 で返す。
func isMissingReference(err error) bool {
	var me *mysql.MySQLError
	return errors.As(err, &me) && me.Number == mysqlNoReferencedRow
}

// isForeignKeyViolation は FK RESTRICT による削除・更新拒否（親行が参照されている）を判定する。
// 呼び出し側は文脈に応じたメッセージを付けて domain.ErrConflict へ写像する。
func isForeignKeyViolation(err error) bool {
	var me *mysql.MySQLError
	return errors.As(err, &me) && me.Number == mysqlRowIsReferenced
}
