package main

import (
	"errors"
	"strconv"
	"strings"
	"unicode"
)

func UnpackString(s string) (string, error) {
	if s == "" {
		return "", nil
	}

	var result strings.Builder
	var prevChar rune
	var escaped bool

	for _, char := range s {
		switch {
		case escaped:
			result.WriteRune(char)
			prevChar = char
			escaped = false

		case char == '\\':
			escaped = true

		case unicode.IsDigit(char):
			if prevChar == 0 {
				return "", errors.New("digit at the beginning of string")
			}

			count, err := strconv.Atoi(string(char))
			if err != nil {
				return "", errors.New("invalid digit")
			}

			for i := 1; i < count; i++ {
				result.WriteRune(prevChar)
			}

		default:
			result.WriteRune(char)
			prevChar = char
		}
	}

	if escaped {
		return "", errors.New("unfinished escape sequence")
	}

	return result.String(), nil
}

func main() {
	var tests = []string{
		"a22bc2d5e",
		"abcd",
		"45",
		"",
		"a1b2c3",
		"qwe\\4\\5",
		"qwe\\45",
		"\\\\",
		"a\\",
		"\\3",
	}

	for _, test := range tests {
		result, err := UnpackString(test)
		if err != nil {
			println("Вход:", test, "-> Ошибка:", err.Error())
		} else {
			println("Вход:", test, "-> Выход:", result)
		}
	}
}
