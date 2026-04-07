package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/spf13/pflag"
)

var (
	patternFlag     = pflag.String("pattern", "", "regex pattern to search for")
	serversListFlag = pflag.StringSlice("servers", nil, "list of servers")
	quorumFlag      = pflag.Int("quorum", 0, "number of quorum nodes (n+2/1 by default)")
	chunkSize       = pflag.Int("chunksize", 1024, "size of each chunk")
	filePath        = pflag.String("file", "", "file path (stdin is read if not specified)")
	coordinatorPort = pflag.String("port", ":6000", "coordinator listen port for callbacks")
	ignoreCaseFlag  = pflag.Bool("i", false, "Ignore case")
	inverseFlag     = pflag.Bool("v", false, "Invert match")
	countFlag       = pflag.Bool("c", false, "Count only")
	fixedStringFlag = pflag.Bool("F", false, "Interpret pattern as fixed string")
	beforeFlag      = pflag.Int("B", 0, "Number of context lines before")
	afterFlag       = pflag.Int("A", 0, "Number of context lines after")
	contextFlag     = pflag.Int("C", 0, "Number of context lines around")
)

type processRequest struct {
	ID       string `json:"id"`
	Data     string `json:"data"`
	Pattern  string `json:"pattern"`
	Callback string `json:"callback,omitempty"`
	// Добавляем флаги для передачи воркеру
	IgnoreCase  bool `json:"ignore_case"`
	Inverse     bool `json:"inverse"`
	CountOnly   bool `json:"count_only"`
	FixedString bool `json:"fixed_string"`
	Before      int  `json:"before"`
	After       int  `json:"after"`
	Context     int  `json:"context"`
}

type processResponse struct {
	ID      string   `json:"id"`
	Matches []string `json:"matches"`
	Error   string   `json:"error,omitempty"`
}

type chunkTask struct {
	seq  int
	data string
}

func launchWorkerClient(ctx context.Context, servers []string, dispatched *int32, callbackURL string) chan<- chunkTask {
	chunkChan := make(chan chunkTask, 10)
	serverChannels := make([]chan chunkTask, len(servers), len(servers))

	for n, server := range servers {
		serverChannels[n] = make(chan chunkTask, 10)
		go func(serverURL string, serverChannel chan chunkTask) {
			client := http.Client{Timeout: time.Second * 10}
			for {
				select {
				case chunk, ok := <-serverChannel:
					if !ok {
						return
					}
					request := processRequest{
						ID:          fmt.Sprintf("chunk-%d", chunk.seq),
						Data:        chunk.data,
						Pattern:     *patternFlag,
						Callback:    callbackURL,
						IgnoreCase:  *ignoreCaseFlag,
						Inverse:     *inverseFlag,
						CountOnly:   *countFlag,
						FixedString: *fixedStringFlag,
						Before:      *beforeFlag,
						After:       *afterFlag,
						Context:     *contextFlag,
					}
					requestJson, _ := json.Marshal(request)
					resp, err := client.Post(serverURL+"/process", "application/json", bytes.NewBuffer(requestJson))
					if err != nil {
						log.Printf("Error posting to %s: %s", serverURL, err)
						continue
					}
					log.Printf("Posted to %s", serverURL)
					if resp != nil && resp.Body != nil {
						resp.Body.Close()
					}

					atomic.AddInt32(dispatched, 1)
					log.Printf("Task %d accepted by %s", chunk.seq, serverURL)
				case <-ctx.Done():
					client.CloseIdleConnections()
					return
				}
			}
		}(server, serverChannels[n])
	}

	go func() {
		for {
			select {
			case chunk, ok := <-chunkChan:
				if !ok {
					for n := range serverChannels {
						close(serverChannels[n])
					}
					return
				}
				serverChannels[chunk.seq%len(serverChannels)] <- chunk
			case <-ctx.Done():
				return
			}
		}
	}()

	return chunkChan
}

type processResponseResult struct {
	Value processResponse
	Err   error
}

func launchWorkerServer(ctx context.Context, success *int32) <-chan processResponseResult {
	responseChan := make(chan processResponseResult, 10)
	mux := http.NewServeMux()
	server := http.Server{Addr: *coordinatorPort, Handler: mux}

	go func() {
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	go func() {
		<-ctx.Done()
		if err := server.Shutdown(ctx); err != nil {
			log.Fatalf("server shutdown: %s\n", err)
		}
	}()

	mux.HandleFunc("/sendResults", func(w http.ResponseWriter, r *http.Request) {
		var response processResponse
		err := json.NewDecoder(r.Body).Decode(&response)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			log.Printf("Error decoding JSON: %s\n", err)
			responseChan <- processResponseResult{Err: err}
		}
		w.WriteHeader(http.StatusOK) // ← Подтверждаем приём колбэка

		if response.Error == "" {
			atomic.AddInt32(success, 1) // ← Считаем успешные завершения
		}

		log.Printf("Received response from server: %s\n", response.ID)
		responseChan <- processResponseResult{Value: response, Err: nil}
	})

	return responseChan
}

func main() {
	pflag.Parse()

	if *patternFlag == "" {
		log.Fatal("-pattern is required")
	}
	if *serversListFlag == nil {
		log.Fatal("-servers is required")
	}
	if *chunkSize <= 0 {
		*chunkSize = 1000
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	ctxTimeout, _ := context.WithTimeout(ctx, 1*time.Minute)
	defer stop()

	var input *os.File
	var err error
	if *filePath != "" {
		input, err = os.Open(*filePath)
		if err != nil {
			log.Fatalf("error reading file %s: %v", *filePath, err)
		}
		defer input.Close()
		fmt.Printf("[COORDINATOR] Reading file: %s\n", *filePath)
	} else {
		input = os.Stdin
		fmt.Println("[COORDINATOR] Reading stdin...")
	}

	n := len(*serversListFlag)
	requiredPercent := float64(*quorumFlag) / float64(n) * 100.0
	var dispatched int32
	var success int32

	allMatches := make([]string, 0, n)

	responseChan := launchWorkerServer(ctx, &success)
	go func() {
		for {
			select {
			case response := <-responseChan:
				if response.Err != nil {
					log.Printf("error reading response: %v", response.Err)
				}
				allMatches = append(allMatches, response.Value.Matches...)
			case <-ctx.Done():
				return
			}
		}
	}()

	requestChan := launchWorkerClient(ctx, *serversListFlag, &dispatched, "http://localhost"+*coordinatorPort)
	scanner := bufio.NewScanner(input)
	lines := make([]string, 0, *chunkSize)
	seq := 0
	go func() {
		defer close(requestChan)
		for scanner.Scan() {
			line := scanner.Text()
			if strings.TrimSpace(line) == "" {
				continue
			}
			lines = append(lines, line)

			if len(lines) >= *chunkSize {
				chunkLines := make([]string, len(lines))
				copy(chunkLines, lines)
				requestChan <- chunkTask{
					seq:  seq,
					data: strings.Join(chunkLines, "\n"),
				}
				seq++
				lines = make([]string, 0, *chunkSize)
			}
		}
		if len(lines) > 0 {
			requestChan <- chunkTask{
				seq:  seq,
				data: strings.Join(lines, "\n"),
			}
		}
	}()

	finished := false
	for !finished {
		select {
		case <-ctxTimeout.Done():
			finished = true
		case <-ctx.Done():
			finished = true
		default:
			sent := atomic.LoadInt32(&dispatched)
			recv := atomic.LoadInt32(&success)
			if sent > 0 && recv == sent {
				finished = true
			} else {
				time.Sleep(50 * time.Millisecond)
			}
		}
	}

	// Код проверки кворума и вывода (выполняется строго после завершения цикла)
	sent := atomic.LoadInt32(&dispatched)
	recv := atomic.LoadInt32(&success)

	if sent == 0 {
		log.Println("[COORDINATOR] No chunks dispatched.")
		return
	}

	actualPercent := float64(recv) / float64(sent) * 100.0
	if actualPercent < requiredPercent {
		log.Fatalf("[QUORUM FAIL] Success: %d/%d (%.1f%%). Required: %d servers (%.1f%%)",
			recv, sent, actualPercent, *quorumFlag, requiredPercent)
	}

	for _, m := range allMatches {
		fmt.Println(m)
	}
	log.Printf("[COORDINATOR] Done. Matches: %d. Quorum: %d/%d servers (%.1f%%)",
		len(allMatches), recv, sent, actualPercent)
}
