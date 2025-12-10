package ntptime

import (
	"errors"
	"log"
	"time"

	"github.com/beevik/ntp"
)

type NTPTime struct {
	CurrentTime       time.Time
	CurrentTimeString string
}

var DefaultNTPServers = []string{
	"pool.ntp.org",
	"time.google.com",
	"time.cloudflare.com",
	"ntp.ubuntu.com",
}

// GetTime получает текущее время с первого доступного сервера NTP.
//
// Параметры:
//   - format: Шаблон форматирования даты/времени (например, "2006-01-02 15:04:05").
//
// Возвращает:
//   - NTPTime: Структура с локальным временем и его строковым представлением.
//   - error: Ошибка, если ни один сервер NTP не смог быть достигнут.
func GetTime(address string, format string) (NTPTime, error) {
	servers := []string{address}
	if address == "" {
		servers = DefaultNTPServers
	}

	for _, server := range servers {
		log.Println("Trying to reach", server)
		ntpTime, err := ntp.Time(server)
		if err == nil {
			return NTPTime{
				CurrentTime:       ntpTime.Local(),
				CurrentTimeString: ntpTime.Local().Format(format),
			}, nil
		}
		log.Printf("Failed to reach %s: %v", server, err)
	}

	return NTPTime{}, errors.New("failed to get time")
}
