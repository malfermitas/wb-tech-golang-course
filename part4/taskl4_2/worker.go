//go:build ignore

package main

// Worker запускается отдельно: go run -tags=worker worker.go
// Или компилируется: go build -tags=worker -o worker worker.go

import (
	"bufio"
	"bytes"
	"encoding/json"
	"flag"
	"log"
	"net/http"
	"regexp"
	"strings"
)

var (
	fixedStringFlag = flag.Bool("F", false, "Interpret pattern as fixed string")
	ignoreCaseFlag  = flag.Bool("i", false, "Ignore case")
	inverseFlag     = flag.Bool("v", false, "Invert match")
	countFlag       = flag.Bool("c", false, "Count only")
	beforeFlag      = flag.Int("B", 0, "Number of context lines before")
	afterFlag       = flag.Int("A", 0, "Number of context lines after")
	contextFlag     = flag.Int("C", 0, "Number of context lines around")
	coordinatorURL  = flag.String("callback", "", "URL to send results back to")
	port            = flag.String("port", ":6000", "Port to listen on")
)

type Line struct {
	lineNumber int
	lineText   string
}

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
	flag.Parse()

	if *coordinatorURL == "" {
		log.Fatal("--callback is required")
	}

	mux := http.NewServeMux()
	server := http.Server{Addr: *port, Handler: mux}

	mux.HandleFunc("/process", func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			ID          string `json:"id"`
			Data        string `json:"data"`
			Pattern     string `json:"pattern"`
			IgnoreCase  bool   `json:"ignore_case"`
			Inverse     bool   `json:"inverse"`
			CountOnly   bool   `json:"count_only"`
			FixedString bool   `json:"fixed_string"`
			Before      int    `json:"before"`
			After       int    `json:"after"`
			Context     int    `json:"context"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
			return
		}

		// Переопределяем флаги из запроса
		*ignoreCaseFlag = req.IgnoreCase
		*inverseFlag = req.Inverse
		*countFlag = req.CountOnly
		*fixedStringFlag = req.FixedString
		*beforeFlag = req.Before
		*afterFlag = req.After
		*contextFlag = req.Context

		reader := bytes.NewReader([]byte(req.Data))
		results, err := searchInReader(bufio.NewReader(reader), req.Pattern)

		var resp struct {
			ID      string   `json:"id"`
			Matches []string `json:"matches"`
			Error   string   `json:"error,omitempty"`
		}
		resp.ID = req.ID

		if err != nil {
			resp.Error = err.Error()
		} else {
			for _, line := range results {
				resp.Matches = append(resp.Matches, line.lineText)
			}
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)

		if *coordinatorURL != "" {
			callbackReq, _ := json.Marshal(map[string]interface{}{
				"id":      req.ID,
				"matches": resp.Matches,
				"error":   resp.Error,
			})
			http.Post(*coordinatorURL+"/sendResults", "application/json", bytes.NewBuffer(callbackReq))
		}
	})

	log.Printf("[WORKER] Starting on %s", *port)
	if err := server.ListenAndServe(); err != nil {
		log.Printf("server: %s", err)
	}
}
