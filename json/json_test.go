// Package json

// Created by zc on 2021/7/12.
package json

import (
	"testing"
)

const errJsonStr = `{
"a": 1, "b": [2,3,4],
"c": 22,
"d": false
}`

func Test_validPayload(t *testing.T) {
	type args struct {
		data []byte
		l    int
		i    int
	}
	tests := []struct {
		name     string
		args     args
		wantLine int
		wantOuti int
		wantOk   bool
	}{
		{
			name: "err",
			args: args{
				data: []byte(errJsonStr),
			},
			wantLine: 2,
			wantOuti: 0,
			wantOk:   false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotLine, gotOuti, gotOk := validPayload(tt.args.data, tt.args.l, tt.args.i)
			if gotLine != tt.wantLine {
				t.Errorf("validPayload() gotLine = %v, want %v", gotLine, tt.wantLine)
			}
			if gotOuti != tt.wantOuti {
				t.Errorf("validPayload() gotOuti = %v, want %v", gotOuti, tt.wantOuti)
			}
			if gotOk != tt.wantOk {
				t.Errorf("validPayload() gotOk = %v, want %v", gotOk, tt.wantOk)
			}
		})
	}
}
