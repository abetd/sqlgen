package sqlgen

import (
	"testing"
	"text/template"

	"github.com/stretchr/testify/assert"
)

func TestQueryBuilder_Build(t *testing.T) {
	type fields struct {
		templateSQL  string
		templateData interface{}
		inFuncMap    template.FuncMap
	}
	type want struct {
		sql    string
		params []interface{}
	}
	tests := []struct {
		name    string
		fields  fields
		want    want
		wantErr bool
	}{
		{
			name: "パラメータを使用 数値",
			fields: fields{
				templateSQL:  "SELECT * FROM user_table WHERE id = /** param .ID **/1234",
				templateData: struct{ ID int }{ID: 1234},
			},
			want: want{
				sql:    "SELECT * FROM user_table WHERE id = ?",
				params: []interface{}{1234},
			},
		},
		{
			name: "パラメータを使用 小数",
			fields: fields{
				templateSQL:  "SELECT * FROM user_table WHERE price = /** param .Price **/1.234",
				templateData: struct{ Price float64 }{Price: 1.234},
			},
			want: want{
				sql:    "SELECT * FROM user_table WHERE price = ?",
				params: []interface{}{1.234},
			},
		},
		{
			name: "パラメータを使用 文字列",
			fields: fields{
				templateSQL:  "SELECT * FROM user_table WHERE name = /** param .Name **/'foo bar'",
				templateData: struct{ Name string }{Name: "foo bar"},
				inFuncMap:    nil,
			},
			want: want{
				sql:    "SELECT * FROM user_table WHERE name = ?",
				params: []interface{}{"foo bar"},
			},
		},
		{
			name: "INTを使用 クエリはパラメータ",
			fields: fields{
				templateSQL:  "SELECT * FROM user_table WHERE id = /** int .ID **/?",
				templateData: struct{ ID int }{ID: 1234},
				inFuncMap:    nil,
			},
			want: want{
				sql:    "SELECT * FROM user_table WHERE id = ?",
				params: []interface{}{1234},
			},
		},
		{
			name: "INを使用 数値",
			fields: fields{
				templateSQL:  "SELECT * FROM user_table WHERE id IN /** in .IDs **/(1, 2, 3)",
				templateData: struct{ IDs []interface{} }{IDs: []interface{}{1, 2, 3}},
			},
			want: want{
				sql:    "SELECT * FROM user_table WHERE id IN (?, ?, ?)",
				params: []interface{}{1, 2, 3},
			},
		},
		{
			name: "INを使用 文字列",
			fields: fields{
				templateSQL:  "SELECT * FROM user_table WHERE name IN /** in .Names **/('foo', 'bar', 'var')",
				templateData: struct{ Names []interface{} }{Names: []interface{}{"foo", "bar", "var"}},
			},
			want: want{
				sql:    "SELECT * FROM user_table WHERE name IN (?, ?, ?)",
				params: []interface{}{"foo", "bar", "var"},
			},
		},
		{
			name: "検索条件を繰り返す検索",
			fields: fields{
				templateSQL: "SELECT * FROM user_table WHERE ( /** multi .Where .Sep .Names **/(name LIKE '%foo%' OR kana LIKE '%foo%') )",
				templateData: struct {
					Where string
					Sep   string
					Names []interface{}
				}{
					Where: "(name LIKE ? OR kana LIKE ?)",
					Sep:   "AND",
					Names: []interface{}{"%foo%", "%bar%", "%var%"},
				},
			},
			want: want{
				sql:    "SELECT * FROM user_table WHERE ( (name LIKE ? OR kana LIKE ?) AND (name LIKE ? OR kana LIKE ?) AND (name LIKE ? OR kana LIKE ?) )",
				params: []interface{}{"%foo%", "%foo%", "%bar%", "%bar%", "%var%", "%var%"},
			},
		},
		{
			name: "検索条件を繰り返す検索 検索条件と接続部を指定",
			fields: fields{
				templateSQL: "SELECT * FROM user_table WHERE ( /** multi \"(name LIKE ? OR kana LIKE ?)\" \"AND\" .Names **/(name LIKE '%foo%' OR kana LIKE '%foo%') )",
				templateData: struct {
					Names []interface{}
				}{
					Names: []interface{}{"%foo%", "%bar%", "%var%"},
				},
			},
			want: want{
				sql:    "SELECT * FROM user_table WHERE ( (name LIKE ? OR kana LIKE ?) AND (name LIKE ? OR kana LIKE ?) AND (name LIKE ? OR kana LIKE ?) )",
				params: []interface{}{"%foo%", "%foo%", "%bar%", "%bar%", "%var%", "%var%"},
			},
		},
		{
			name: "複数のパラメータ",
			fields: fields{
				templateSQL: "SELECT * FROM user_table WHERE id = /** param .ID **/1234 OR name IN /** in .Names **/('foo', 'bar', 'var')",
				templateData: struct {
					ID    int
					Names []interface{}
				}{
					ID:    1,
					Names: []interface{}{"foo", "bar", "var"},
				},
			},
			want: want{
				sql:    "SELECT * FROM user_table WHERE id = ? OR name IN (?, ?, ?)",
				params: []interface{}{1, "foo", "bar", "var"},
			},
		},
		{
			name: "使用する条件を動的に変更する",
			fields: fields{
				templateSQL: "SELECT * FROM user_table WHERE /** if .IsSelectID -**/ id = /** param .ID **/1234 /**- end **/ /** if false -**/ OR /**- end **/ /**- if .IsSelectName -**/ name IN /** in .Names **/('foo', 'bar', 'var') /**- end **/",
				templateData: struct {
					IsSelectID   bool
					ID           int
					IsSelectName bool
					Names        []interface{}
				}{
					IsSelectID:   false,
					ID:           1,
					IsSelectName: true,
					Names:        []interface{}{"foo", "bar", "var"},
				},
			},
			want: want{
				sql:    "SELECT * FROM user_table WHERE  name IN (?, ?, ?)",
				params: []interface{}{"foo", "bar", "var"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := NewQueryBuilder(tt.fields.templateSQL, tt.fields.templateData, tt.fields.inFuncMap)
			q, err := b.Build()
			if err != nil {
				t.Errorf("Build() error = %v, wantErr %v", err, tt.wantErr)
			}
			assert.Equal(t, q.SQL, tt.want.sql)
			assert.Equal(t, q.Args, tt.want.params)
		})
	}
}
