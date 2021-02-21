package aiptest

import (
	"strings"
	"unicode"
)

func camelToSnake(str string) string {
	if str == "" {
		return ""
	}
	i := strings.LastIndexFunc(str, unicode.IsUpper)
	for i != -1 {
		if i == 0 {
			str = string(unicode.ToLower(rune(str[i]))) + str[1:]
		} else {
			str = str[:i] + "_" + string(unicode.ToLower(rune(str[i]))) + str[i+1:]
		}
		i = strings.LastIndexFunc(str, unicode.IsUpper)
	}
	return str
}
