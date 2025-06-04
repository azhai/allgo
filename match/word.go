package match

import (
	"strings"
)

// Word 字符串逐个字符匹配
type Word string

// SkipAnyChar 忽略开头字符
func (w Word) SkipAnyChar(chars string) Word {
	i := 0
	for i < len(w) && strings.ContainsRune(chars, rune(w[i])) {
		i++
	}
	return w[i:]
}

// MatchSubString 寻找字符第一次出现的位置，类似IndexAny
func (w Word) MatchSubString(subs string) (string, Word) {
	i := 0
	for i < len(w) && w[i] >= 0x20 && w[i] != 0x7f && w[i] != subs[0] {
		if w[i] == '\\' { // 不要匹配同名转义字符，例如目标是"就要跳过\"
			i++
		}
		i++
	}
	if i == 0 || i >= len(w) || !strings.HasPrefix(string(w[i:]), subs) {
		return "", ""
	}
	return string(w[:i]), w[i+len(subs):]
}

// MatchFirstID 在头部寻找标识符，例如URL中的Schema
func (w Word) MatchFirstID() string {
	var i, offset int
	matched := false
	for i = 0; i < len(w); i++ {
		c := w[i]
		if c == '_' || 'a' <= c && c <= 'z' || 'A' <= c && c <= 'Z' {
			matched = true
			continue
		}
		if matched && '0' <= c && c <= '9' {
			matched = true
			continue
		}
		if !matched && (c == ' ' || c == '\t' || c == '\r' || c == '\n') {
			offset++
		}
		break
	}
	return string(w[offset:i])
}
