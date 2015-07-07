package core

import "strings"

type Replacer struct {
	Map map[string]string
}

func (r Replacer) Replace(str string) string {
	for oldPart, newPart := range r.Map {
		str = strings.Replace(str, oldPart, newPart, -1)
	}
	return str
}
