// Package json
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
package json

func Check(payload string) (line, pos int, ok bool) {
	return CheckBytes([]byte(payload))
}

func CheckBytes(payload []byte) (line, pos int, ok bool) {
	return validPayload(payload, 1, 0)
}

type position struct {
	line    int
	linePos int
	pos     int
	ok      bool
}

func (p *position) buildTrue(poss ...int) *position {
	if len(poss) > 0 {
		p.pos = poss[0]
	}
	p.ok = true
	return p
}

func (p *position) buildFalse(poss ...int) *position {
	if len(poss) > 0 {
		p.pos = poss[0]
	}
	p.ok = false
	return p
}

func (p *position) AddLine() {
	p.line++
	p.linePos = p.pos
}

func (p *position) out() (line, pos int, ok bool) {
	return p.line, p.pos - p.linePos, p.ok
}

func validPayload(data []byte, l, i int) (line, outi int, ok bool) {
	in := &position{line: l, pos: i}
	for ; in.pos < len(data); in.pos++ {
		switch data[in.pos] {
		default:
			res := validAny(data, in)
			if !res.ok {
				return res.out()
			}
			for ; res.pos < len(data); res.pos++ {
				switch data[res.pos] {
				default:
					return res.out()
				case ' ', '\t', '\r':
					continue
				case '\n':
					res.AddLine()
					continue
				}
			}
			return res.buildTrue().out()
		case ' ', '\t', '\r':
			continue
		case '\n':
			in.AddLine()
			continue
		}
	}
	return in.out()
}

func validAny(data []byte, in *position) *position {
	for ; in.pos < len(data); in.pos++ {
		switch data[in.pos] {
		default:
			return in
		case ' ', '\t', '\r':
			continue
		case '\n':
			in.AddLine()
			continue
		case '{':
			in.pos++
			return validObject(data, in)
		case '[':
			in.pos++
			return validArray(data, in)
		case '"':
			in.pos, in.ok = validString(data, in.pos+1)
			return in
		case '-', '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
			in.pos, in.ok = validNumber(data, in.pos+1)
			return in
		case 't':
			in.pos, in.ok = validTrue(data, in.pos+1)
			return in
		case 'f':
			in.pos, in.ok = validFalse(data, in.pos+1)
			return in
		case 'n':
			in.pos, in.ok = validNull(data, in.pos+1)
			return in
		}
	}
	return in
}

// 校验对象
func validObject(data []byte, in *position) *position {
	for ; in.pos < len(data); in.pos++ {
		switch data[in.pos] {
		default:
			return in.buildFalse()
		case ' ', '\t', '\r':
			continue
		case '\n':
			in.AddLine()
			continue
		case '}':
			return in.buildTrue(in.pos + 1)
		case '"':
		key:
			if in.pos, in.ok = validString(data, in.pos+1); !in.ok {
				return in
			}
			if in = validColon(data, in); !in.ok {
				return in
			}
			if in = validAny(data, in); !in.ok {
				return in
			}
			if in = validComma(data, in, '}'); !in.ok {
				return in
			}
			if data[in.pos] == '}' {
				return in.buildTrue(in.pos + 1)
			}
			in.pos++
			for ; in.pos < len(data); in.pos++ {
				switch data[in.pos] {
				default:
					return in.buildFalse()
				case ' ', '\t', '\r':
					continue
				case '\n':
					in.AddLine()
					continue
				case '"':
					goto key
				}
			}
			return in.buildFalse()
		}
	}
	return in.buildFalse()
}

// 校验冒号
func validColon(data []byte, in *position) *position {
	for ; in.pos < len(data); in.pos++ {
		switch data[in.pos] {
		default:
			return in.buildFalse()
		case ' ', '\t', '\r':
			continue
		case '\n':
			in.AddLine()
			continue
		case ':':
			return in.buildTrue(in.pos + 1)
		}
	}
	return in
}

// 校验逗号
func validComma(data []byte, in *position, end byte) *position {
	for ; in.pos < len(data); in.pos++ {
		switch data[in.pos] {
		default:
			return in.buildFalse()
		case ' ', '\t', '\r':
			continue
		case '\n':
			in.AddLine()
			continue
		case ',':
			return in.buildTrue()
		case end:
			return in.buildTrue()
		}
	}
	return in
}

// 校验数组
func validArray(data []byte, in *position) *position {
	for ; in.pos < len(data); in.pos++ {
		switch data[in.pos] {
		default:
			for ; in.pos < len(data); in.pos++ {
				if in = validAny(data, in); !in.ok {
					return in
				}
				if in = validComma(data, in, ']'); !in.ok {
					return in
				}
				if data[in.pos] == ']' {
					return in.buildTrue(in.pos + 1)
				}
			}
		case ' ', '\t', '\r':
			continue
		case '\n':
			in.AddLine()
			continue
		case ']':
			return in.buildTrue(in.pos + 1)
		}
	}
	return in.buildFalse()
}

func validString(data []byte, i int) (outi int, ok bool) {
	for ; i < len(data); i++ {
		if data[i] < ' ' {
			return i, false
		} else if data[i] == '\\' {
			i++
			if i == len(data) {
				return i, false
			}
			switch data[i] {
			default:
				return i, false
			case '"', '\\', '/', 'b', 'f', 'n', 'r', 't':
			case 'u':
				for j := 0; j < 4; j++ {
					i++
					if i >= len(data) {
						return i, false
					}
					if !((data[i] >= '0' && data[i] <= '9') ||
						(data[i] >= 'a' && data[i] <= 'f') ||
						(data[i] >= 'A' && data[i] <= 'F')) {
						return i, false
					}
				}
			}
		} else if data[i] == '"' {
			return i + 1, true
		}
	}
	return i, false
}

func validNumber(data []byte, i int) (outi int, ok bool) {
	i--
	// sign
	if data[i] == '-' {
		i++
		if i == len(data) {
			return i, false
		}
		if data[i] < '0' || data[i] > '9' {
			return i, false
		}
	}
	// int
	if i == len(data) {
		return i, false
	}
	if data[i] == '0' {
		i++
	} else {
		for ; i < len(data); i++ {
			if data[i] >= '0' && data[i] <= '9' {
				continue
			}
			break
		}
	}
	// frac
	if i == len(data) {
		return i, true
	}
	if data[i] == '.' {
		i++
		if i == len(data) {
			return i, false
		}
		if data[i] < '0' || data[i] > '9' {
			return i, false
		}
		i++
		for ; i < len(data); i++ {
			if data[i] >= '0' && data[i] <= '9' {
				continue
			}
			break
		}
	}
	// exp
	if i == len(data) {
		return i, true
	}
	if data[i] == 'e' || data[i] == 'E' {
		i++
		if i == len(data) {
			return i, false
		}
		if data[i] == '+' || data[i] == '-' {
			i++
		}
		if i == len(data) {
			return i, false
		}
		if data[i] < '0' || data[i] > '9' {
			return i, false
		}
		i++
		for ; i < len(data); i++ {
			if data[i] >= '0' && data[i] <= '9' {
				continue
			}
			break
		}
	}
	return i, true
}

func validTrue(data []byte, i int) (outi int, ok bool) {
	if i+3 <= len(data) && data[i] == 'r' && data[i+1] == 'u' &&
		data[i+2] == 'e' {
		return i + 3, true
	}
	return i, false
}

func validFalse(data []byte, i int) (outi int, ok bool) {
	if i+4 <= len(data) && data[i] == 'a' && data[i+1] == 'l' &&
		data[i+2] == 's' && data[i+3] == 'e' {
		return i + 4, true
	}
	return i, false
}

func validNull(data []byte, i int) (outi int, ok bool) {
	if i+3 <= len(data) && data[i] == 'u' && data[i+1] == 'l' &&
		data[i+2] == 'l' {
		return i + 3, true
	}
	return i, false
}
