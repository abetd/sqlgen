package sqlgen

import (
	_ "embed"
	"fmt"
	"go/format"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"text/template"
	"unicode"
	"unicode/utf8"
)

//go:embed templates/query_elem.go.tmpl
var queryElemGoTmpl string

type Field struct {
	Name string
	Type string
}

type Struct struct {
	SQLFilePath  string
	SQLFileEmbed string
	StructName   string
	Fields       []Field
}

type File struct {
	Package string
	Structs []Struct
}

type CodeGen struct {
	dir     string
	outFile string
}

func NewCodeGen(dir string) *CodeGen {
	return &CodeGen{
		dir:     dir,
		outFile: dir + string(filepath.Separator) + "query_elements_gen.go",
	}
}

func (g *CodeGen) CodeGen() error {
	files, err := getFiles(g.dir, ".sql")
	if err != nil {
		return err
	}

	structs := make([]Struct, len(files))
	for i, file := range files {
		sqlFile := filepath.Base(file)
		structName := toCamel(StripExt(sqlFile)) + "QueryElem"
		content, err := os.ReadFile(file)
		if err != nil {
			return err
		}

		matches := extractBlockComments(string(content))
		var fields []Field
		for _, match := range matches {
			newFields, err := getFields(match)
			if err != nil {
				return err
			}
			if len(newFields) == 0 {
				continue
			}
			fields = append(fields, newFields...)
		}
		structs[i] = Struct{
			SQLFilePath:  sqlFile,
			SQLFileEmbed: lowerFirst(structName) + "SQL",
			StructName:   structName,
			Fields:       fields,
		}
	}
	file := File{
		Package: lastDirName(g.dir),
		Structs: structs,
	}

	tmpl, err := template.New("").Parse(queryElemGoTmpl)
	if err != nil {
		return err
	}
	var sb strings.Builder
	if err := tmpl.Execute(&sb, file); err != nil {
		return err
	}
	src, err := format.Source([]byte(sb.String()))
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(g.outFile), 0o755); err != nil {
		return err
	}
	if err := os.WriteFile(g.outFile, src, 0o644); err != nil {
		return err
	}
	return nil
}

func parseFields(src string) []string {
	var fields []string
	var current strings.Builder
	inQuotes := false
	i := 0

	for i < len(src) {
		char := src[i]

		if char == '"' && (i == 0 || src[i-1] != '\\') {
			inQuotes = !inQuotes
			current.WriteByte(char)
		} else if char == ' ' && !inQuotes {
			// スペースで区切るが、クォート内でない場合のみ
			if current.Len() > 0 {
				fields = append(fields, strings.TrimSpace(current.String()))
				current.Reset()
			}
			// 連続するスペースをスキップ
			for i+1 < len(src) && src[i+1] == ' ' {
				i++
			}
		} else {
			current.WriteByte(char)
		}
		i++
	}

	// 最後のフィールドを追加
	if current.Len() > 0 {
		fields = append(fields, strings.TrimSpace(current.String()))
	}

	return fields
}

func getFields(src string) ([]Field, error) {
	blocks := parseFields(src)
	if len(blocks) < 2 {
		return nil, nil
	}
	var fields []Field
	switch blocks[0] {
	case "param":
		fields = appendFiled(fields, blocks[1], "interface{}")
	case "int":
		fields = appendFiled(fields, blocks[1], "int")
	case "float":
		fields = appendFiled(fields, blocks[1], "float64")
	case "string":
		fields = appendFiled(fields, blocks[1], "string")
	case "if":
		fields = appendFiled(fields, blocks[1], "bool")
	case "in":
		fields = appendFiled(fields, blocks[1], "[]interface{}")
	case "multi":
		if len(blocks) != 4 {
			return nil, fmt.Errorf("invalid number of fields: %d", len(blocks))
		}
		// 1: (name LIKE ? OR kana LIKE ?), 2: "AND", 3: []interface{}{"%foo%", "%bar%", "%var%"}
		// (name LIKE ? OR kana LIKE ?) AND (name LIKE ? OR kana LIKE ?) AND (name LIKE ? OR kana LIKE ?)
		fields = appendFiled(fields, blocks[1], "string")
		fields = appendFiled(fields, blocks[2], "string")
		fields = appendFiled(fields, blocks[3], "[]interface{}")
	}
	return fields, nil
}

func appendFiled(fields []Field, s, t string) []Field {
	if !strings.HasPrefix(s, ".") {
		return fields
	}
	return append(fields, Field{Name: strings.TrimPrefix(s, "."), Type: t})
}

func getFiles(dir string, ext string) ([]string, error) {
	var matchedFiles []string
	entries, err := os.ReadDir(dir)
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if strings.EqualFold(filepath.Ext(entry.Name()), ext) {
			matchedFiles = append(matchedFiles, filepath.Join(dir, entry.Name()))
		}
	}
	return matchedFiles, err
}

func extractBlockComments(content string) []string {
	re := regexp.MustCompile(`(?s)/\*\*-?(.*?)-?\*\*/`)
	matches := re.FindAllStringSubmatch(content, -1)

	var results []string
	for _, m := range matches {
		results = append(results, m[1])
	}
	return results
}

func toCamel(s string) string {
	if s == "" {
		return s
	}
	parts := strings.Split(s, "_")
	var b strings.Builder
	for _, p := range parts {
		if p == "" {
			continue
		}
		p = strings.ToLower(p)
		r := []rune(p)
		r[0] = unicode.ToUpper(r[0])
		b.WriteString(string(r))
	}
	return b.String()
}

func StripExt(name string) string {
	ext := filepath.Ext(name)
	return strings.TrimSuffix(name, ext)
}

func lastDirName(p string) string {
	p = filepath.Clean(p)
	return filepath.Base(p)
}

func lowerFirst(s string) string {
	if s == "" {
		return s
	}
	r, size := utf8.DecodeRuneInString(s)
	return string(unicode.ToLower(r)) + s[size:]
}
