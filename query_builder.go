package sqlgen

import (
	"bytes"
	"regexp"
	"strings"
	"text/template"
)

const (
	leftDelimiter  = "/**"
	rightDelimiter = "**/"
)

type QueryBuilder struct {
	templateSQL  string
	templateData interface{}
	inFuncMap    template.FuncMap
	params       []interface{}
}

type Query struct {
	SQL  string
	Args []interface{}
}

func (q *Query) addParam(arg interface{}) {
	q.Args = append(q.Args, arg)
}

func NewQueryBuilder(sql string, data interface{}, funcMap template.FuncMap) *QueryBuilder {
	return &QueryBuilder{
		templateSQL:  sql,
		templateData: data,
		inFuncMap:    funcMap,
	}
}

type QueryElem interface {
	SQLTemplate() string
}

func CreateQuery(e QueryElem) (*Query, error) {
	b := NewQueryBuilder(e.SQLTemplate(), e, nil)
	return b.Build()
}

func (q *QueryBuilder) Build() (*Query, error) {
	cleaner := NewSQLQueryCleaner()
	cleanedSQL := cleaner.CleanQuery(q.templateSQL)
	query := Query{}
	t := template.Must(template.New("").Delims(leftDelimiter, rightDelimiter).Funcs(q.funcMap(query.addParam)).Parse(cleanedSQL))
	var buf bytes.Buffer
	if err := t.Execute(&buf, q.templateData); err != nil {
		return nil, err
	}
	query.SQL = buf.String()
	return &query, nil
}

func (q *QueryBuilder) funcMap(addParam func(arg interface{})) template.FuncMap {
	funcMap := template.FuncMap{
		"param": func(arg interface{}) string {
			addParam(arg)
			return "?"
		},
		"int": func(arg interface{}) string {
			addParam(arg)
			return "?"
		},
		"string": func(arg interface{}) string {
			addParam(arg)
			return "?"
		},
		"in": func(args []interface{}) string {
			p := make([]string, len(args))
			for i, arg := range args {
				p[i] = "?"
				addParam(arg)
			}
			return "(" + strings.Join(p, ", ") + ")"
		},
		"multi": func(query, sep string, args []interface{}) string {
			paramCount := strings.Count(query, "?")
			queries := make([]string, len(args))
			for i, arg := range args {
				queries[i] = query
				for j := 0; j < paramCount; j++ {
					addParam(arg)
				}
			}
			return strings.Join(queries, sep)
		},
	}
	for k, v := range q.inFuncMap {
		funcMap[k] = v
	}
	return funcMap
}

// SQLQueryCleaner はSQLクエリからコメント後の値を削除する
type SQLQueryCleaner struct {
	// 各パターンに対応する正規表現
	numberPattern *regexp.Regexp
	stringPattern *regexp.Regexp
	inPattern     *regexp.Regexp
}

// NewSQLQueryCleaner は新しいSQLQueryCleanerを作成
func NewSQLQueryCleaner() *SQLQueryCleaner {
	return &SQLQueryCleaner{
		// パターン1: 数字・小数 - コメント後の数字や小数を削除（その後のスペースは保持）
		numberPattern: regexp.MustCompile(`(/\*\*[^*]*\*\*/)\d+(?:\.\d+)?(\s|$)`),
		// パターン2: 文字列 - コメント後の'...'を削除（その後のスペースは保持）
		stringPattern: regexp.MustCompile(`(/\*\*[^*]*\*\*/)'[^']*'(\s|$)`),
		// パターン3: IN句 - コメント後の(...)を削除（その後のスペースは保持）
		inPattern: regexp.MustCompile(`(/\*\*[^*]*\*\*/)\([^)]*\)(\s|$)`),
	}
}

// CleanQuery はSQLクエリからコメント後の値を削除
func (c *SQLQueryCleaner) CleanQuery(query string) string {
	// 順番が重要：より具体的なパターンから先に処理
	// パターン3: IN句の処理
	query = c.inPattern.ReplaceAllString(query, "$1$2")
	// パターン2: 文字列の処理
	query = c.stringPattern.ReplaceAllString(query, "$1$2")
	// パターン1: 数字の処理
	query = c.numberPattern.ReplaceAllString(query, "$1$2")
	return query
}
