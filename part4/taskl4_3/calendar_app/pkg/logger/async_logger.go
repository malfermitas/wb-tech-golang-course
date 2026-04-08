package logger

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"
)

type Logger struct {
	msgChan chan string
	wg      sync.WaitGroup
	ctx     context.Context
	stopped bool
	stopMu  sync.Mutex
}

func New(ctx context.Context) *Logger {
	return &Logger{
		msgChan: make(chan string, 1000),
		ctx:     ctx,
	}
}

func (l *Logger) Start() {
	l.wg.Add(1)
	go func() {
		defer l.wg.Done()
		for {
			select {
			case msg := <-l.msgChan:
				log.Println(msg)
			case <-l.ctx.Done():
				l.flush()
				return
			}
		}
	}()
}

func (l *Logger) flush() {
	l.stopMu.Lock()
	if l.stopped {
		l.stopMu.Unlock()
		return
	}
	l.stopped = true
	l.stopMu.Unlock()

	close(l.msgChan)
	for msg := range l.msgChan {
		log.Println(msg)
	}
}

func (l *Logger) Stop() {
	l.stopMu.Lock()
	if l.stopped {
		l.stopMu.Unlock()
		return
	}
	l.stopped = true
	l.stopMu.Unlock()

	close(l.msgChan)
	l.wg.Wait()
}

func (l *Logger) Info(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	select {
	case l.msgChan <- msg:
	default:
		log.Println("logger: channel full, dropping message")
	}
}

func (l *Logger) WithFields(fields map[string]interface{}) *Logged {
	return &Logged{logger: l, fields: fields}
}

type Logged struct {
	logger *Logger
	fields map[string]interface{}
}

func (l *Logged) Info(msg string) {
	l.logger.Info("%s %v", msg, l.fields)
}

type LogEntry struct {
	Time   time.Time
	Level  string
	Msg    string
	Fields map[string]interface{}
}
