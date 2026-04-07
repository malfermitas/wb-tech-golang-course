package main

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/spf13/pflag"
)

type Line struct {
	lineNumber int
	lineText   string
}

// flags declaration
var (
	afterFlag       = pflag.IntP("A", "A", 0, "print next N lines after the line matching regular expression")
	beforeFlag      = pflag.IntP("B", "B", 0, "print next N lines before the line matching regular expression")
	countFlag       = pflag.BoolP("count", "c", false, "print only a count of matching lines per FILE")
	contextFlag     = pflag.IntP("context", "C", 0, "print N lines of leading and trailing context surrounding each match")
	ignoreCaseFlag  = pflag.BoolP("ignore-case", "i", false, "ignore case distinctions")
	inverseFlag     = pflag.BoolP("invert-match", "v", false, "select non-matching lines")
	fixedStringFlag = pflag.BoolP("fixed-string", "F", false, "interpret pattern as fixed string, not regular expression")
	printNumberFlag = pflag.BoolP("line-number", "n", false, "print line number with output lines")
	filePathFlag    = pflag.StringP("file", "f", "", "File to read. If not specified, will use standard input")
)

func searchInReader(reader *bufio.Reader, pattern string) ([]Line, error) {
	var patternRegularExpression *regexp.Regexp
	var regexpError error
	addLineIfMissing := func(lines []Line, line Line) []Line {
		for _, addedLine := range lines {
			if addedLine.lineNumber == line.lineNumber {
				return lines
			}
		}
		return append(lines, line)
	}

	if *fixedStringFlag {
		// Если флаг -F установлен, интерпретируем как фиксированную строку
		escapedPattern := regexp.QuoteMeta(pattern)
		if *ignoreCaseFlag {
			patternRegularExpression, regexpError = regexp.Compile("(?i)" + escapedPattern)
		} else {
			patternRegularExpression, regexpError = regexp.Compile(escapedPattern)
		}
	} else {
		// Обычная обработка регулярного выражения
		if *ignoreCaseFlag {
			patternRegularExpression, regexpError = regexp.Compile("(?i)" + pattern)
		} else {
			patternRegularExpression, regexpError = regexp.Compile(pattern)
		}
	}

	if regexpError != nil {
		return nil, regexpError
	}

	resultingLines := make([]Line, 0)

	linesBuffer := make([]Line, 0)
	afterLinesRemaining := 0

	lineNumber := 1

	scanner := bufio.NewScanner(reader)

	for scanner.Scan() {
		line := scanner.Text()
		originalLine := line

		checkLine := line
		if *ignoreCaseFlag && !*fixedStringFlag {
			checkLine = strings.ToLower(checkLine)
		}

		matchFound := patternRegularExpression.MatchString(checkLine)

		if *inverseFlag {
			matchFound = !matchFound
		}

		linesBuffer = append(linesBuffer, Line{lineNumber: lineNumber, lineText: originalLine})

		if matchFound {
			if *countFlag {
				resultingLines = append(resultingLines, Line{lineNumber: lineNumber, lineText: originalLine})
			} else {
				contextLines := *beforeFlag
				if *contextFlag > 0 {
					contextLines = *contextFlag
				}

				startIndex := len(linesBuffer) - 1 - contextLines
				if startIndex < 0 {
					startIndex = 0
				}

				for i := startIndex; i < len(linesBuffer)-1; i++ {
					resultingLines = addLineIfMissing(resultingLines, linesBuffer[i])
				}

				resultingLines = addLineIfMissing(resultingLines, linesBuffer[len(linesBuffer)-1])

				afterLines := *afterFlag
				if *contextFlag > 0 {
					afterLines = *contextFlag
				}
				if afterLines > afterLinesRemaining {
					afterLinesRemaining = afterLines
				}
			}
		} else if !*countFlag && afterLinesRemaining > 0 {
			resultingLines = addLineIfMissing(resultingLines, Line{lineNumber: lineNumber, lineText: originalLine})
			afterLinesRemaining--
		}

		contextSize := *beforeFlag
		if *contextFlag > 0 {
			contextSize = *contextFlag
		}
		if len(linesBuffer) > contextSize+1 {
			linesBuffer = linesBuffer[1:]
		}

		lineNumber++
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return resultingLines, nil
}

func main() {
	pflag.Parse()

	args := pflag.Args()

	// Проверяем наличие паттерна
	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "Error: pattern is required")
		os.Exit(1)
	}

	pattern := args[0]

	var lines []Line
	var err error

	if *filePathFlag != "" {
		// Читаем из файла
		file, fileErr := os.Open(*filePathFlag)
		if fileErr != nil {
			fmt.Fprintf(os.Stderr, "Error opening file: %v\n", fileErr)
			os.Exit(1)
		}
		defer file.Close()

		reader := bufio.NewReader(file)
		lines, err = searchInReader(reader, pattern)
	} else if len(args) >= 2 {
		// Читаем из указанного файла в аргументах
		file, fileErr := os.Open(args[1])
		if fileErr != nil {
			fmt.Fprintf(os.Stderr, "Error opening file: %v\n", fileErr)
			os.Exit(1)
		}
		defer file.Close()

		reader := bufio.NewReader(file)
		lines, err = searchInReader(reader, pattern)
	} else {
		// Читаем из stdin
		reader := bufio.NewReader(os.Stdin)
		lines, err = searchInReader(reader, pattern)
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	if *countFlag {
		fmt.Println(len(lines))
	} else {
		for _, line := range lines {
			if *printNumberFlag {
				fmt.Printf("%d:%s\n", line.lineNumber, line.lineText)
			} else {
				fmt.Println(line.lineText)
			}
		}
	}
}
