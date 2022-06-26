package main

import "strconv"

func parse_uint(str string) (uint, bool) {
	u, err := strconv.ParseUint(str, 0, 32)
	if err != nil {
		return 0, false
	}
	return uint(u), true
}