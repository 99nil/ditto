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
	"fmt"
	"io"
	"sort"
	"strings"
)

type Map = xml

type xml map[string]interface{}

type xmlData struct {
	XMLName xmle.Name
	Attr    []xmle.Attr `xml:",attr"`
	Value   interface{} `xml:",chardata"`
}

var arrayAttr = xmle.Attr{
	Name: xmle.Name{
		Local: "type",
	},
	Value: "array",
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
			err = encodeSlice(e, key, val)
		default:
			err = e.Encode(xmlData{
				XMLName: xmle.Name{
					Local: key,
				},
				Value: val,
			})
		}
		if err != nil {
			return err
		}
	}
	return err
}

func encodeSlice(e *xmle.Encoder, key string, data []interface{}) error {
	var err error
	for _, v := range data {
		switch value := v.(type) {
		case map[string]interface{}:
			err = e.EncodeElement(xml(value), xmle.StartElement{
				Name: xmle.Name{
					Local: key,
				},
				Attr: []xmle.Attr{arrayAttr},
			})
		case []interface{}:
		default:
			err = e.Encode(xmlData{
				XMLName: xmle.Name{
					Local: key,
				},
				Attr:  []xmle.Attr{arrayAttr},
				Value: v,
			})
		}
		if err != nil {
			return err
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
		akc      = make(map[string]int)
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
			isArray := false
			for _, attr := range tv.Attr {
				if attr == arrayAttr {
					isArray = true
					break
				}
			}
			indexArr = append(indexArr, tv.Name.Local)
			if isArray {
				akc[strings.Join(indexArr, ".")]++
			}
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
				loop(m, string(charData), indexArr, akc, 0)
				charData = nil
			}
			indexArr = indexArr[:indexArrLen-1]
		}
	}
	return nil
}

// 自动合并数组
func loop(data interface{}, value string, keys []string, akc map[string]int, index int) {
	// 判断数据类型
	switch dv := data.(type) {
	case map[string]interface{}:
		loopMap(dv, value, keys, akc, index)
	case xml:
		loopMap(dv, value, keys, akc, index)
	case *[]interface{}:
		loopSlice(dv, value, keys, akc, index)
	default:
		fmt.Printf("%T\n", data)
	}
}

func loopMap(data xml, value string, keys []string, akc map[string]int, index int) {
	// 获取key剩余个数
	extraLen := len(keys) - index
	// 获取当前key
	key := keys[index]
	// 获取当前key路径
	index++
	pres := keys[:index]
	switch extraLen {
	case 0:
		return
	case 1:
		// 判断是否为数组
		if akc[strings.Join(pres, ".")] > 0 {
			// 判断值是否存在，不存在初始化
			if dataVal, ok := data[key]; !ok {
				data[key] = make([]interface{}, 0)
			} else if _, ok = dataVal.([]interface{}); !ok {
				data[key] = make([]interface{}, 0)
			}
			// 判断第一个值类型，不同则跳过赋值
			if len(data[key].([]interface{})) > 0 {
				switch data[key].([]interface{})[0].(type) {
				case map[string]interface{}, []interface{}:
					return
				}
			}
			data[key] = append(data[key].([]interface{}), value)
			return
		}

		if val, ok := data[key]; ok {
			switch val.(type) {
			case map[string]interface{}, []interface{}:
				return
			}
		}
		data[key] = value
	default:
		// 判断value是否初始化
		if dataVal, ok := data[key]; !ok {
			// 未初始化value，判断该key是否为数组
			if akc[strings.Join(pres, ".")] > 0 {
				arr := make([]interface{}, 0)
				loop(&arr, value, keys, akc, index)
				data[key] = arr
				return
			}
			data[key] = make(map[string]interface{})
		} else if _, ok = dataVal.([]interface{}); ok {
			// 已初始化数组，判断该key是否为数组
			if akc[strings.Join(pres, ".")] > 0 {
				arr := data[key].([]interface{})
				loop(&arr, value, keys, akc, index)
				data[key] = arr
				return
			}
			data[key] = make(map[string]interface{})
		} else if _, ok = dataVal.(map[string]interface{}); ok {
			// 已初始化map，判断该key是否为数组
			if akc[strings.Join(pres, ".")] > 0 {
				arr := make([]interface{}, 0)
				loop(&arr, value, keys, akc, index)
				data[key] = arr
				return
			}
		}
		loop(data[key], value, keys, akc, index)
	}
}

func loopSlice(data *[]interface{}, value string, keys []string, akc map[string]int, index int) {
	// 获取key剩余个数
	extraLen := len(keys) - index
	// 获取当前key路径
	pres := keys[:index]
	preArrNum := akc[strings.Join(pres, ".")]
	pres = keys[:index+1]
	switch extraLen {
	case 0:
		*data = append(*data, value)
	default:
		// 获取上次数组数量推演下标，并初始化补全
		if preArrNum == 0 {
			return
		}
		dataExtraLen := preArrNum - len(*data)
		for i := 0; i < dataExtraLen; i++ {
			*data = append(*data, struct{}{})
		}

		switch (*data)[preArrNum-1].(type) {
		case map[string]interface{}:
		default:
			// 重赋值为map
			(*data)[preArrNum-1] = make(map[string]interface{})
		}
		loop((*data)[preArrNum-1], value, keys, akc, index)
	}
}
