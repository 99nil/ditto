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
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"strconv"

	xmle "github.com/99nil/ditto/xml"
	"github.com/pelletier/go-toml/v2"
	"gopkg.in/yaml.v2"
)

var set = make(map[string]*Engine)

const (
	FormatJSON = "json"
	FormatYaml = "yaml"
	FormatXML  = "xml"
	FormatTOML = "toml"
)

func init() {
	Register(FormatJSON, json.Marshal, json.Unmarshal)
	Register(FormatYaml, yaml.Marshal, yaml.Unmarshal)
	Register(FormatXML, xml.Marshal, xml.Unmarshal)
	Register(FormatTOML, toml.Marshal, toml.Unmarshal)

	RegisterED(FormatJSON, func(w io.Writer) Encoder {
		return json.NewEncoder(w)
	}, func(r io.Reader) Decoder {
		return json.NewDecoder(r)
	})
	RegisterED(FormatYaml, func(w io.Writer) Encoder {
		return yaml.NewEncoder(w)
	}, func(r io.Reader) Decoder {
		return yaml.NewDecoder(r)
	})
	RegisterED(FormatXML, func(w io.Writer) Encoder {
		return xml.NewEncoder(w)
	}, func(r io.Reader) Decoder {
		return xml.NewDecoder(r)
	})
	RegisterED(FormatTOML, func(w io.Writer) Encoder {
		return toml.NewEncoder(w)
	}, func(r io.Reader) Decoder {
		return toml.NewDecoder(r)
	})
}

type (
	MarshalFunc   func(v interface{}) ([]byte, error)
	UnmarshalFunc func(data []byte, v interface{}) error

	EncoderFunc func(w io.Writer) Encoder
	DecoderFunc func(r io.Reader) Decoder
)

type (
	Encoder interface {
		Encode(v interface{}) error
	}
	Decoder interface {
		Decode(v interface{}) error
	}
)

type Engine struct {
	marshal   MarshalFunc
	unmarshal UnmarshalFunc
	eFunc     EncoderFunc
	dFunc     DecoderFunc
}

func Register(name string, m MarshalFunc, um UnmarshalFunc) {
	e, ok := set[name]
	if !ok {
		e = &Engine{}
		set[name] = e
	}
	e.marshal = m
	e.unmarshal = um
}

func RegisterED(name string, eFunc EncoderFunc, dFunc DecoderFunc) {
	e, ok := set[name]
	if !ok {
		e = &Engine{}
		set[name] = e
	}
	e.eFunc = eFunc
	e.dFunc = dFunc
}

func Marshal(name string, v interface{}) ([]byte, error) {
	e, ok := set[name]
	if !ok {
		return nil, fmt.Errorf("failed to find %s engine", name)
	}
	return e.marshal(v)
}

func Unmarshal(name string, data []byte, v interface{}) error {
	e, ok := set[name]
	if !ok {
		return fmt.Errorf("failed to find %s engine", name)
	}
	return e.unmarshal(data, v)
}

type Transfer struct {
	in  string
	out string
}

func NewTransfer(in, out string) *Transfer {
	return &Transfer{in: in, out: out}
}

func (t *Transfer) Exchange(data []byte) ([]byte, error) {
	ipr, ok := set[t.in]
	if !ok {
		return nil, errors.New("failed to find input engine")
	}
	opr, ok := set[t.out]
	if !ok {
		return nil, errors.New("failed to find output engine")
	}
	var spec interface{}
	if t.in == FormatXML {
		xmlSpec := make(xmle.Map)
		if err := ipr.unmarshal(data, &xmlSpec); err != nil {
			return nil, err
		}
		spec = xmlSpec
	} else {
		if err := ipr.unmarshal(data, &spec); err != nil {
			return nil, err
		}
	}
	if err := transformData(&spec); err != nil {
		return nil, err
	}
	if t.out == FormatXML {
		spec = xmle.Map(spec.(map[string]interface{}))
	}
	return opr.marshal(spec)
}

func (t *Transfer) ExchangeED(r io.Reader, w io.Writer) error {
	iParser, ok := set[t.in]
	if !ok {
		return errors.New("failed to find reader engine")
	}
	oParser, ok := set[t.out]
	if !ok {
		return errors.New("failed to find writer engine")
	}
	var spec interface{}
	if err := iParser.dFunc(r).Decode(&spec); err != nil {
		return err
	}
	if err := transformData(&spec); err != nil {
		return err
	}
	return oParser.eFunc(w).Encode(&spec)
}

func transformData(pIn *interface{}) (err error) {
	switch in := (*pIn).(type) {
	case map[interface{}]interface{}:
		m := make(map[string]interface{}, len(in))
		for k, v := range in {
			if err = transformData(&v); err != nil {
				return err
			}
			var sk string
			switch val := k.(type) {
			case string:
				sk = val
			case int:
				sk = strconv.Itoa(val)
			case bool:
				sk = strconv.FormatBool(val)
			case nil:
				sk = "null"
			case float64:
				sk = strconv.FormatFloat(val, 'f', -1, 64)
			default:
				return fmt.Errorf("type mismatch: expect map key string or int; got: %T", k)
			}
			m[sk] = v
		}
		*pIn = m
	case []interface{}:
		for i := len(in) - 1; i >= 0; i-- {
			if err = transformData(&in[i]); err != nil {
				return err
			}
		}
	}
	return nil
}
