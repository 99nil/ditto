// Package ditto
// Copyright © 2021 zc2638 <zc2638@qq.com>.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package ditto

import (
	"bytes"
	"io"
	"reflect"
	"strings"
	"testing"

	jsoniter "github.com/json-iterator/go"
)

func TestNewTransfer(t *testing.T) {
	type args struct {
		in  string
		out string
	}
	tests := []struct {
		name string
		args args
		want *Transfer
	}{
		{
			name: "json-to-yaml",
			args: args{
				in:  FormatJSON,
				out: FormatYaml,
			},
			want: &Transfer{in: FormatJSON, out: FormatYaml},
		},
		{
			name: "json-to-toml",
			args: args{
				in:  FormatJSON,
				out: FormatTOML,
			},
			want: &Transfer{in: FormatJSON, out: FormatTOML},
		},
		{
			name: "xml-to-toml",
			args: args{
				in:  FormatXML,
				out: FormatTOML,
			},
			want: &Transfer{in: FormatXML, out: FormatTOML},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewTransfer(tt.args.in, tt.args.out); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewTransfer() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRegister(t *testing.T) {
	type args struct {
		name string
		m    MarshalFunc
		um   UnmarshalFunc
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "register-jsoniter",
			args: args{
				name: "jsoniter",
				m:    jsoniter.Marshal,
				um:   jsoniter.Unmarshal,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			Register(tt.args.name, tt.args.m, tt.args.um)
		})
	}
}

func TestRegisterED(t *testing.T) {
	type args struct {
		name string
		ne   EncoderFunc
		nd   DecoderFunc
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "registerED-jsoniter",
			args: args{
				name: "jsoniter",
				ne: func(w io.Writer) Encoder {
					return jsoniter.NewEncoder(w)
				},
				nd: func(r io.Reader) Decoder {
					return jsoniter.NewDecoder(r)
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			RegisterED(tt.args.name, tt.args.ne, tt.args.nd)
		})
	}
}

const (
	jsonStr = `{"database":{"connection_max":5000,"ports":[8001,8002,8003],"server":"127.0.0.1"},"owner":{"name":"zc"},"title":"demo"}`

	yamlStr = `
database:
  connection_max: 5000
  ports:
  - 8001
  - 8002
  - 8003
  server: 127.0.0.1
owner:
  name: zc
title: demo
`

	tomlStr = `
title = 'demo'
[database]
connection_max = 5000
ports = [8001, 8002, 8003]
server = '127.0.0.1'

[owner]
name = 'zc'
`

	xmlStr = `
<xml>
    <database>
        <connection_max>5000</connection_max>
        <ports type="array">8001</ports>
        <ports type="array">8002</ports>
        <ports type="array">8003</ports>
        <server>127.0.0.1</server>
    </database>
    <owner>
        <name>zc</name>
    </owner>
	<title>demo</title>
</xml>
`
	yamlStr2 = `
database:
  connection_max: "5000"
  ports:
  - "8001"
  - "8002"
  - "8003"
  server: 127.0.0.1
owner:
  name: zc
title: demo
`
)

func TestTransfer_Exchange(t1 *testing.T) {
	type fields struct {
		in  string
		out string
	}
	type args struct {
		data []byte
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []byte
		wantErr bool
	}{
		{
			name: "json-to-yaml",
			fields: fields{
				in:  FormatJSON,
				out: FormatYaml,
			},
			args: args{
				data: []byte(jsonStr),
			},
			want:    []byte(yamlStr),
			wantErr: false,
		},
		{
			name: "yaml-to-json",
			fields: fields{
				in:  FormatYaml,
				out: FormatJSON,
			},
			args: args{
				data: []byte(yamlStr),
			},
			want:    []byte(jsonStr),
			wantErr: false,
		},
		{
			name: "toml-to-json",
			fields: fields{
				in:  FormatTOML,
				out: FormatJSON,
			},
			args: args{
				data: []byte(tomlStr),
			},
			want:    []byte(jsonStr),
			wantErr: false,
		},
		{
			name: "yaml-to-toml",
			fields: fields{
				in:  FormatYaml,
				out: FormatTOML,
			},
			args: args{
				data: []byte(yamlStr),
			},
			want:    []byte(tomlStr),
			wantErr: false,
		},
		{
			name: "yaml-to-xml",
			fields: fields{
				in:  FormatYaml,
				out: FormatXML,
			},
			args: args{
				data: []byte(yamlStr),
			},
			want:    []byte(xmlStr),
			wantErr: false,
		},
		{
			name: "xml-to-yaml",
			fields: fields{
				in:  FormatXML,
				out: FormatYaml,
			},
			args: args{
				data: []byte(xmlStr),
			},
			want:    []byte(yamlStr2),
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t1.Run(tt.name, func(t1 *testing.T) {
			t := &Transfer{
				in:  tt.fields.in,
				out: tt.fields.out,
			}
			got, err := t.Exchange(tt.args.data)
			if (err != nil) != tt.wantErr {
				t1.Errorf("Exchange() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			got = UnifiedTreatment(got)
			tt.want = UnifiedTreatment(tt.want)
			if !reflect.DeepEqual(got, tt.want) {
				t1.Errorf("Exchange() got = %s\n want = %s\n", got, tt.want)
			}
		})
	}
}

func UnifiedTreatment(data []byte) []byte {
	data = bytes.TrimLeft(data, "\n")
	data = bytes.TrimRight(data, "\n")
	data = bytes.ReplaceAll(data, []byte("\t"), []byte("    "))
	return data
}

func TestTransfer_ExchangeED(t1 *testing.T) {
	type fields struct {
		in  string
		out string
	}
	type args struct {
		r io.Reader
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantW   string
		wantErr bool
	}{
		{
			name: "json-to-yaml",
			fields: fields{
				in:  FormatJSON,
				out: FormatYaml,
			},
			args: args{
				r: strings.NewReader(jsonStr),
			},
			wantW:   yamlStr,
			wantErr: false,
		},
		{
			name: "yaml-to-json",
			fields: fields{
				in:  FormatYaml,
				out: FormatJSON,
			},
			args: args{
				r: strings.NewReader(yamlStr),
			},
			wantW:   jsonStr,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t1.Run(tt.name, func(t1 *testing.T) {
			t := &Transfer{
				in:  tt.fields.in,
				out: tt.fields.out,
			}
			w := &bytes.Buffer{}
			err := t.ExchangeED(tt.args.r, w)
			if (err != nil) != tt.wantErr {
				t1.Errorf("ExchangeED() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotW := w.String(); strings.TrimSuffix(gotW, "\n") != tt.wantW {
				t1.Errorf("ExchangeED() gotW = %v, want %v", gotW, tt.wantW)
			}
		})
	}
}
