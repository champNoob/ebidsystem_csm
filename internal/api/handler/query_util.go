package handler

import "strconv"

func parseIntDefault(raw string, def int) int {
	if raw == "" {
		return def
	}

	n, err := strconv.Atoi(raw)
	if err != nil {
		return def
	}

	return n
}
