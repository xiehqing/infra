package jsonx

import (
	"bytes"
	"encoding/json"
	"strconv"
	"strings"

	"github.com/pkg/errors"
)

type JSONFixer struct{}

func NewJSONFixer() *JSONFixer {
	return &JSONFixer{}
}

func FixJSON(input string) (string, error) {
	return NewJSONFixer().Fix(input)
}

func MustFixJSON(input string) string {
	fixed, err := FixJSON(input)
	if err != nil {
		return ""
	}
	return fixed
}

func (f *JSONFixer) Fix(input string) (string, error) {
	_ = f

	input = preprocessJSON(input)
	if input == "" {
		return "", errors.New("json input is empty")
	}

	if json.Valid([]byte(input)) {
		return compactJSON(input)
	}

	parser := newLooseJSONParser(input)
	repaired := parser.parse()
	if repaired == "" {
		return "", errors.New("repair json result is empty")
	}
	if !json.Valid([]byte(repaired)) {
		return "", errors.New("repair json result is invalid")
	}
	return compactJSON(repaired)
}

func (f *JSONFixer) FixBytes(input []byte) ([]byte, error) {
	fixed, err := f.Fix(string(input))
	if err != nil {
		return nil, err
	}
	return []byte(fixed), nil
}

func compactJSON(input string) (string, error) {
	var compact bytes.Buffer
	if err := json.Compact(&compact, []byte(input)); err != nil {
		return "", errors.WithMessage(err, "compact repaired json failed")
	}
	return compact.String(), nil
}

func preprocessJSON(input string) string {
	input = strings.TrimSpace(input)
	input = strings.TrimPrefix(input, "```json")
	input = strings.TrimPrefix(input, "```JSON")
	input = strings.TrimSuffix(input, "```")
	input = strings.TrimSpace(input)

	replacer := strings.NewReplacer(
		"\u201c", `"`,
		"\u201d", `"`,
		"\u2018", `'`,
		"\u2019", `'`,
	)
	return replacer.Replace(input)
}

type looseJSONParser struct {
	src string
	pos int
}

func newLooseJSONParser(src string) *looseJSONParser {
	return &looseJSONParser{src: src}
}

func (p *looseJSONParser) parse() string {
	p.skipSpaces()
	if p.eof() {
		return ""
	}

	var result string
	switch p.peek() {
	case '{':
		result = p.parseObject()
	case '[':
		result = p.parseArray()
	default:
		result = p.parseValue()
	}
	return result
}

func (p *looseJSONParser) parseObject() string {
	var buf bytes.Buffer
	buf.WriteByte('{')
	p.consumeIf('{')

	wroteAny := false
	for !p.eof() {
		p.skipSpaces()
		p.consumeIf(',')
		p.skipSpaces()

		if p.eof() {
			break
		}
		if p.peek() == '}' {
			p.pos++
			break
		}

		key, ok := p.parseKey()
		if !ok {
			p.pos++
			continue
		}

		p.skipSpaces()
		p.consumeIf(':')
		p.skipSpaces()

		value := p.parseValue()
		if value == "" {
			value = `""`
		}

		if wroteAny {
			buf.WriteByte(',')
		}
		buf.WriteString(key)
		buf.WriteByte(':')
		buf.WriteString(value)
		wroteAny = true
	}

	buf.WriteByte('}')
	return buf.String()
}

func (p *looseJSONParser) parseArray() string {
	var buf bytes.Buffer
	buf.WriteByte('[')
	p.consumeIf('[')

	wroteAny := false
	for !p.eof() {
		p.skipSpaces()
		p.consumeIf(',')
		p.skipSpaces()

		if p.eof() {
			break
		}
		if p.peek() == ']' {
			p.pos++
			break
		}

		value := p.parseValue()
		if value == "" {
			p.pos++
			continue
		}

		if wroteAny {
			buf.WriteByte(',')
		}
		buf.WriteString(value)
		wroteAny = true
	}

	buf.WriteByte(']')
	return buf.String()
}

func (p *looseJSONParser) parseValue() string {
	p.skipSpaces()
	if p.eof() {
		return ""
	}

	switch p.peek() {
	case '{':
		return p.parseObject()
	case '[':
		return p.parseArray()
	case '"', '\'':
		return p.parseQuotedString(false)
	default:
		if isNumberStart(p.peek()) {
			token := p.readNumberToken()
			if number, ok := normalizeNumberToken(token); ok {
				return number
			}
		}

		token := p.readIdentifierToken()
		if token == "" {
			return ""
		}

		lower := strings.ToLower(token)
		switch lower {
		case "true", "false", "null":
			return lower
		default:
			return marshalJSONString(token)
		}
	}
}

func (p *looseJSONParser) parseKey() (string, bool) {
	p.skipSpaces()
	if p.eof() {
		return "", false
	}

	switch p.peek() {
	case '"', '\'':
		return p.parseQuotedString(true), true
	default:
		token := p.readKeyToken()
		if token == "" {
			return "", false
		}
		return marshalJSONString(token), true
	}
}

func (p *looseJSONParser) parseQuotedString(isKey bool) string {
	quote := p.peek()
	p.pos++

	var raw bytes.Buffer
	for !p.eof() {
		ch := p.peek()

		if ch == '\\' {
			if p.pos+1 < len(p.src) {
				raw.WriteByte(ch)
				p.pos++
				raw.WriteByte(p.peek())
				p.pos++
				continue
			}
			p.pos++
			continue
		}

		if ch == quote {
			next := p.peekNextNonSpace(1)
			if isKey {
				if next == ':' {
					p.pos++
					break
				}
			} else if next == ',' || next == ']' || next == '}' || next == 0 || p.looksLikeObjectKeyAhead(1) {
				p.pos++
				break
			}

			raw.WriteByte(ch)
			p.pos++
			continue
		}

		if ch == '"' && quote == '"' {
			next := p.peekNextNonSpace(1)
			if next == ',' || next == ']' || next == '}' || next == 0 || p.looksLikeObjectKeyAhead(1) {
				p.pos++
				break
			}
			raw.WriteByte('"')
			p.pos++
			continue
		}

		if ch == '\n' || ch == '\r' {
			raw.WriteByte(' ')
			p.pos++
			continue
		}

		raw.WriteByte(ch)
		p.pos++
	}

	return marshalJSONString(raw.String())
}

func (p *looseJSONParser) readKeyToken() string {
	start := p.pos
	for !p.eof() {
		ch := p.peek()
		if ch == ':' || ch == ',' || ch == '}' || ch == ']' || isSpace(ch) {
			break
		}
		p.pos++
	}
	return strings.TrimSpace(p.src[start:p.pos])
}

func (p *looseJSONParser) readIdentifierToken() string {
	start := p.pos
	for !p.eof() {
		ch := p.peek()
		if ch == ',' || ch == '}' || ch == ']' || isSpace(ch) {
			break
		}
		if ch == ':' {
			break
		}
		p.pos++
	}
	return strings.TrimSpace(p.src[start:p.pos])
}

func (p *looseJSONParser) readNumberToken() string {
	start := p.pos
	for !p.eof() {
		ch := p.peek()
		if (ch < '0' || ch > '9') && ch != '-' && ch != '+' && ch != '.' && ch != 'e' && ch != 'E' {
			break
		}
		p.pos++
	}
	return strings.TrimSpace(p.src[start:p.pos])
}

func normalizeNumberToken(token string) (string, bool) {
	if token == "" {
		return "", false
	}
	if strings.HasPrefix(token, ".") {
		token = "0" + token
	}
	if strings.HasPrefix(token, "-.") {
		token = "-0" + token[1:]
	}
	if strings.HasPrefix(token, "+.") {
		token = "+0" + token[1:]
	}

	if strings.ContainsAny(token, ".eE") {
		if _, err := strconv.ParseFloat(token, 64); err == nil {
			return token, true
		}
		return "", false
	}
	if _, err := strconv.ParseInt(token, 10, 64); err == nil {
		return token, true
	}
	return "", false
}

func marshalJSONString(input string) string {
	bytes, _ := json.Marshal(input)
	return string(bytes)
}

func (p *looseJSONParser) skipSpaces() {
	for !p.eof() && isSpace(p.peek()) {
		p.pos++
	}
}

func (p *looseJSONParser) consumeIf(ch byte) bool {
	if !p.eof() && p.peek() == ch {
		p.pos++
		return true
	}
	return false
}

func (p *looseJSONParser) peek() byte {
	if p.eof() {
		return 0
	}
	return p.src[p.pos]
}

func (p *looseJSONParser) peekNextNonSpace(offset int) byte {
	index := p.pos + offset
	for index < len(p.src) {
		if !isSpace(p.src[index]) {
			return p.src[index]
		}
		index++
	}
	return 0
}

func (p *looseJSONParser) looksLikeObjectKeyAhead(offset int) bool {
	index := p.pos + offset
	for index < len(p.src) && isSpace(p.src[index]) {
		index++
	}
	if index >= len(p.src) {
		return false
	}

	if p.src[index] == '"' || p.src[index] == '\'' {
		quote := p.src[index]
		index++
		for index < len(p.src) && p.src[index] != quote {
			if p.src[index] == '\\' && index+1 < len(p.src) {
				index += 2
				continue
			}
			index++
		}
		if index < len(p.src) {
			index++
		}
		for index < len(p.src) && isSpace(p.src[index]) {
			index++
		}
		return index < len(p.src) && p.src[index] == ':'
	}

	start := index
	for index < len(p.src) {
		ch := p.src[index]
		if ch == ':' {
			return index > start
		}
		if ch == ',' || ch == '}' || ch == ']' {
			return false
		}
		if isSpace(ch) {
			for index < len(p.src) && isSpace(p.src[index]) {
				index++
			}
			return index < len(p.src) && p.src[index] == ':'
		}
		index++
	}
	return false
}

func (p *looseJSONParser) eof() bool {
	return p.pos >= len(p.src)
}

func isSpace(ch byte) bool {
	switch ch {
	case ' ', '\t', '\n', '\r':
		return true
	default:
		return false
	}
}

func isNumberStart(ch byte) bool {
	return (ch >= '0' && ch <= '9') || ch == '-' || ch == '+' || ch == '.'
}
