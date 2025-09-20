package sqlgen

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_getFields(t *testing.T) {
	type args struct {
		src string
	}
	tests := []struct {
		name    string
		args    args
		want    []Field
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name:    "param filed",
			args:    args{src: "param .Filed"},
			want:    []Field{{Name: "Filed", Type: "interface{}"}},
			wantErr: assert.NoError,
		},
		{
			name:    "param value",
			args:    args{src: "param 12345"},
			want:    nil,
			wantErr: assert.NoError,
		},
		{
			name:    "int filed",
			args:    args{src: "int .Filed"},
			want:    []Field{{Name: "Filed", Type: "int"}},
			wantErr: assert.NoError,
		},
		{
			name:    "float filed",
			args:    args{src: "float .Filed"},
			want:    []Field{{Name: "Filed", Type: "float64"}},
			wantErr: assert.NoError,
		},
		{
			name:    "string filed",
			args:    args{src: "string .Filed"},
			want:    []Field{{Name: "Filed", Type: "string"}},
			wantErr: assert.NoError,
		},
		{
			name:    "if condition",
			args:    args{src: "if .IsFoo"},
			want:    []Field{{Name: "IsFoo", Type: "bool"}},
			wantErr: assert.NoError,
		},
		{
			name:    "in array",
			args:    args{src: "in .Names"},
			want:    []Field{{Name: "Names", Type: "[]interface{}"}},
			wantErr: assert.NoError,
		},
		{
			name: "multi ",
			args: args{src: "multi .Where .Sep .Slice"},
			want: []Field{
				{Name: "Where", Type: "string"},
				{Name: "Sep", Type: "string"},
				{Name: "Slice", Type: "[]interface{}"},
			},
			wantErr: assert.NoError,
		},
		{
			name: "multi クエリー、セパレートを指定",
			args: args{src: "multi \"(name LIKE ? OR kana LIKE ?)\" \" AND \" .Slice"},
			want: []Field{
				{Name: "Slice", Type: "[]interface{}"},
			},
			wantErr: assert.NoError,
		},
		{
			name:    "multi パラメータ不足",
			args:    args{src: "multi .Where .Sep"},
			want:    nil,
			wantErr: assert.Error,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getFields(tt.args.src)
			if !tt.wantErr(t, err, fmt.Sprintf("getFields(%v)", tt.args.src)) {
				return
			}
			assert.Equalf(t, tt.want, got, "getFields(%v)", tt.args.src)
		})
	}
}
