/*
 * Copyright (c) 2019 Guangzhou DOYU Information Technology Co., Ltd. All right reserved.
 */

package stringutils

import "strings"

func IsBlank(str string) bool {
	return str == "" || strings.TrimSpace(str) == ""
}

func DefaultIfBlank(str string, defaultStr string) string {
	if IsBlank(str) {
		return defaultStr
	}
	return strings.TrimSpace(str)
}

func OmittedString(str string, needLen int, appendOmitted string) string {
	runeStr := []rune(str)
	if len(runeStr) <= needLen {
		return str
	}
	return string(runeStr[:needLen]) + appendOmitted
}
