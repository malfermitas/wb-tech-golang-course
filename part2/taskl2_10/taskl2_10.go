package main

import (
	"bufio"
	"fmt"
	"hash/fnv"
	"io"
	"os"
	"sort"
	"strconv"
	"strings"
	"unicode"

	"github.com/spf13/pflag"
)

type Line struct {
	OriginalText  string
	SeparatedText []string
	hashSum       uint64
}

type Lines []Line

func (l Lines) Len() int      { return len(l) }
func (l Lines) Swap(i, j int) { l[i], l[j] = l[j], l[i] }

func (l Lines) Less(i, j int) bool {
	line1 := l[i]
	line2 := l[j]

	var text1, text2 string

	if *columnFlag > 0 && *columnFlag <= len(line1.SeparatedText) && *columnFlag <= len(line2.SeparatedText) {
		text1 = line1.SeparatedText[*columnFlag-1]
		text2 = line2.SeparatedText[*columnFlag-1]
	} else {
		text1 = line1.OriginalText
		text2 = line2.OriginalText
	}

	if *ignoreTrailingFlag {
		text1 = strings.TrimRight(text1, " ")
		text2 = strings.TrimRight(text2, " ")
	}

	if *humanFlag {
		val1 := parseHumanReadable(text1)
		val2 := parseHumanReadable(text2)

		if *reverseFlag {
			return val1 > val2
		}
		return val1 < val2
	}

	if *numericFlag {
		num1, err1 := strconv.ParseFloat(text1, 64)
		num2, err2 := strconv.ParseFloat(text2, 64)

		if err1 == nil && err2 == nil {
			if *reverseFlag {
				return num1 > num2
			}
			return num1 < num2
		}
	}

	if *monthFlag {
		monthOrder := map[string]int{
			"jan": 1, "feb": 2, "mar": 3, "apr": 4, "may": 5, "jun": 6,
			"jul": 7, "aug": 8, "sep": 9, "oct": 10, "nov": 11, "dec": 12,
		}

		mon1 := strings.ToLower(text1)[:3]
		mon2 := strings.ToLower(text2)[:3]

		order1, ok1 := monthOrder[mon1]
		order2, ok2 := monthOrder[mon2]

		if ok1 && ok2 {
			if *reverseFlag {
				return order1 > order2
			}
			return order1 < order2
		}
	}

	if *reverseFlag {
		return text1 > text2
	}
	return text1 < text2
}

var (
	columnFlag         = pflag.Int("k", 0, "sort by column number (1-based)")
	numericFlag        = pflag.Bool("n", false, "numeric sort")
	reverseFlag        = pflag.Bool("r", false, "reverse sort")
	uniqueFlag         = pflag.Bool("u", false, "output only unique lines")
	monthFlag          = pflag.Bool("M", false, "sort by month name")
	ignoreTrailingFlag = pflag.Bool("b", false, "ignore trailing blanks")
	checkFlag          = pflag.Bool("c", false, "check if input is sorted")
	humanFlag          = pflag.Bool("h", false, "compare human readable numbers")
)

func getHashSumOfString(data string) uint64 {
	h := fnv.New64a()
	h.Write([]byte(data))
	return h.Sum64()
}

func readLines(reader io.Reader) ([]Line, error) {
	var lines []Line
	scanner := bufio.NewScanner(reader)

	for scanner.Scan() {
		text := scanner.Text()
		fields := strings.Split(text, "\t")
		lines = append(lines, Line{
			OriginalText:  text,
			SeparatedText: fields,
			hashSum:       getHashSumOfString(text),
		})
	}

	return lines, scanner.Err()
}

func parseHumanReadable(s string) float64 {
	s = strings.TrimSpace(s)
	if len(s) == 0 {
		return 0
	}

	// Находим позицию, где заканчивается число
	i := 0
	for i < len(s) && (unicode.IsDigit(rune(s[i])) || s[i] == '.' || s[i] == ',') {
		i++
	}

	if i == 0 {
		return 0
	}

	// Парсим числовую часть
	numStr := strings.Replace(s[:i], ",", "", -1)
	num, err := strconv.ParseFloat(numStr, 64)
	if err != nil {
		return 0
	}

	// Обрабатываем суффиксы
	if i < len(s) {
		suffix := strings.ToUpper(string(s[i]))
		switch suffix {
		case "K":
			num *= 1000
		case "M":
			num *= 1000000
		case "G":
			num *= 1000000000
		case "T":
			num *= 1000000000000
		case "P":
			num *= 1000000000000000
		case "E":
			num *= 1000000000000000000
		case "Z":
			num *= 1e21
		case "Y":
			num *= 1e24
		}
	}

	return num
}

func removeDuplicates(lines []Line) []Line {
	if len(lines) == 0 {
		return lines
	}

	result := map[uint64]Line{}

	for _, line := range lines {
		result[line.hashSum] = line
	}

	resultArray := []Line{}

	for _, line := range result {
		resultArray = append(resultArray, line)
	}

	return resultArray
}

func main() {

	pflag.Parse()

	// Читаем строки из stdin
	lines, err := readLines(os.Stdin)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading input: %v\n", err)
		os.Exit(1)
	}

	// Если установлен флаг -c, проверяем отсортированность
	if *checkFlag {
		isSorted := sort.SliceIsSorted(lines, func(i, j int) bool {
			return Lines(lines).Less(i, j)
		})
		if isSorted {
			fmt.Println("Input is sorted")
			os.Exit(0)
		} else {
			fmt.Println("Input is not sorted")
			os.Exit(1)
		}
	}

	// Сортируем строки
	sort.Stable(Lines(lines))

	// Если установлен флаг -u, удаляем дубликаты
	if *uniqueFlag {
		lines = removeDuplicates(lines)
	}

	// Выводим результат
	for _, line := range lines {
		fmt.Println(line.OriginalText)
	}
}
