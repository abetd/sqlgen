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
	sql := q.removeComment(q.templateSQL)
	query := Query{}
	t := template.Must(template.New("").Delims(leftDelimiter, rightDelimiter).Funcs(q.funcMap(query.addParam)).Parse(sql))
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

var removeCommentPatterns = []*regexp.Regexp{
	// コメントのあとの?を削除 その後のスペースは保持
	regexp.MustCompile(`(/\*\*[^*]*\*\*/)\?(\s|$)`),
	// コメントのあとの(...)を削除 その後のスペースは保持
	regexp.MustCompile(`(/\*\*[^*]*\*\*/)\([^)]*\)(\s|$)`),
	// コメントのあとの文字列'str'を削除 その後のスペースは保持
	regexp.MustCompile(`(/\*\*[^*]*\*\*/)'[^']*'(\s|$)`),
	// コメントのあとの数字や小数を削除 その後のスペースは保持
	regexp.MustCompile(`(/\*\*[^*]*\*\*/)\d+(?:\.\d+)?(\s|$)`),
}

func (q *QueryBuilder) removeComment(query string) string {
	for _, pattern := range removeCommentPatterns {
		query = pattern.ReplaceAllString(query, "$1$2")
	}
	return query
}
