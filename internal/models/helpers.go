package models

import "strconv"

func itoa(i int) string {
	if i == 0 {
		return ""
	}
	return strconv.Itoa(i)
}
