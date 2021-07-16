// Package xml
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
package xml

import (
	xmle "encoding/xml"
	"errors"
	"io"
	"sort"
)

type Map = xml

type xml map[string]interface{}

type xmlData struct {
	XMLName xmle.Name
	Value   interface{} `xml:",chardata"`
}

func sortXML(e *xmle.Encoder, data xml) error {
	keys := make([]string, 0, len(data))
	for k, _ := range data {
		keys = append(keys, k)
	}
	sort.Sort(sort.StringSlice(keys))
	var err error
	for _, key := range keys {
		v := data[key]
		switch val := v.(type) {
		case map[string]interface{}:
			err = e.EncodeElement(xml(val), xmle.StartElement{
				Name: xmle.Name{
					Local: key,
				},
			})
		case []interface{}:
			for _, vv := range val {
				if err := e.Encode(xmlData{
					XMLName: xmle.Name{
						Local: key,
					},
					Value: vv,
				}); err != nil {
					return err
				}
			}
		default:
			err = e.Encode(xmlData{
				XMLName: xmle.Name{
					Local: key,
				},
				Value: val,
			})
		}
	}
	return err
}

func (m xml) MarshalXML(e *xmle.Encoder, start xmle.StartElement) error {
	if len(m) == 0 {
		return nil
	}
	e.Indent("", "    ")

	err := e.EncodeToken(start)
	if err != nil {
		return err
	}

	if err := sortXML(e, m); err != nil {
		return err
	}
	return e.EncodeToken(start.End())
}

func (m xml) UnmarshalXML(d *xmle.Decoder, start xmle.StartElement) error {
	var (
		indexArr []string
		charData []byte
	)
	for {
		token, err := d.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		switch tv := token.(type) {
		case xmle.StartElement:
			indexArr = append(indexArr, tv.Name.Local)
		case xmle.CharData:
			charData = tv.Copy()
		case xmle.EndElement:
			if tv.Name.Local == start.Name.Local {
				break
			}
			if len(indexArr) == 0 {
				return errors.New("format error, expect startElement")
			}
			indexArrLen := len(indexArr)
			end := indexArr[indexArrLen-1]
			if end != tv.Name.Local {
				return errors.New("format error, expect current endElement")
			}

			if charData != nil {
				loopMap(m, string(charData), indexArr...)
				charData = nil
			}
			indexArr = indexArr[:indexArrLen-1]
		}
	}
	return nil
}

// 自动合并数组
func loopMap(data map[string]interface{}, v string, keys ...string) {
	keysLen := len(keys)
	switch keysLen {
	case 0:
		return
	case 1:
		key := keys[0]
		if val, ok := data[key]; ok {
			switch val.(type) {
			case map[string]interface{}:
			case []interface{}:
				data[key] = append(data[key].([]interface{}), v)
			default:
				data[key] = make([]interface{}, 0)
				data[key] = append(data[key].([]interface{}), val, v)
			}
		} else {
			data[key] = v
		}
		return
	default:
		if _, ok := data[keys[0]].(map[string]interface{}); !ok {
			data[keys[0]] = make(map[string]interface{})
		}
		loopMap(data[keys[0]].(map[string]interface{}), v, keys[1:]...)
	}
}
