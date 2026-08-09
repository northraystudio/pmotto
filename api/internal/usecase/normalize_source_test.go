package usecase

import "testing"

// TestNormalizeSource は取得元の指定を「種別と識別子はセット」として正規化することを固定する。
// 片方だけを独立に空文字判定すると、source_type=NULL なのに source_value だけ残る行ができ、
// どのワークフローも収集対象にできない（抽出条件が source_type で絞るため）。
func TestNormalizeSource(t *testing.T) {
	deref := func(p *string) string {
		if p == nil {
			return "<nil>"
		}
		return *p
	}

	tests := []struct {
		name      string
		inType    string
		inValue   string
		wantType  string
		wantValue string
	}{
		{"両方指定なら両方保持", "spreadsheet", "1AbC", "spreadsheet", "1AbC"},
		{"前後空白は落とす", "  spreadsheet  ", "  1AbC  ", "spreadsheet", "1AbC"},
		{"種別が空なら識別子も捨てる", "", "1AbC", "<nil>", "<nil>"},
		{"種別が空白のみでも識別子を捨てる", "   ", "1AbC", "<nil>", "<nil>"},
		{"種別だけ指定は許す（識別子は後から入れる）", "backlog", "", "backlog", "<nil>"},
		{"両方空なら両方nil", "", "", "<nil>", "<nil>"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotType, gotValue := normalizeSource(tt.inType, tt.inValue)
			if deref(gotType) != tt.wantType {
				t.Errorf("source_type = %s, want %s", deref(gotType), tt.wantType)
			}
			if deref(gotValue) != tt.wantValue {
				t.Errorf("source_value = %s, want %s", deref(gotValue), tt.wantValue)
			}
		})
	}
}
