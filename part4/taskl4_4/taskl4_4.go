package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	_ "net/http/pprof"
	"os/signal"
	"runtime"
	"runtime/metrics"
	"syscall"
	"time"
)

func metricsHandler(w http.ResponseWriter, r *http.Request) {
	var stats runtime.MemStats
	runtime.ReadMemStats(&stats)

	w.Header().Set("Content-Type", "text/plain; version=0.0.4; charset=utf-8")

	// Вспомогательная функция для вывода метрик в формате Prometheus
	writeMetric := func(name, help, typ string, value interface{}) {
		fmt.Fprintf(w, "# HELP %s %s\n", name, help)
		fmt.Fprintf(w, "# TYPE %s %s\n", name, typ)
		fmt.Fprintf(w, "%s %v\n", name, value)
	}

	// Основные метрики памяти и GC
	writeMetric("go_memstats_alloc_bytes", "Количество байт, выделенных и всё ещё используемых.", "gauge", stats.Alloc)
	writeMetric("go_memstats_total_alloc_bytes", "Суммарное количество байт, выделенных под объекты в куче.", "counter", stats.TotalAlloc)
	writeMetric("go_memstats_mallocs_total", "Общее количество выделенных объектов в куче.", "counter", stats.Mallocs)
	writeMetric("go_memstats_frees_total", "Общее количество освобождённых объектов в куче.", "counter", stats.Frees)
	writeMetric("go_memstats_heap_alloc_bytes", "Байты кучи, которые сейчас используются.", "gauge", stats.HeapAlloc)
	writeMetric("go_memstats_heap_sys_bytes", "Байты кучи, полученные от ОС.", "gauge", stats.HeapSys)
	writeMetric("go_memstats_heap_inuse_bytes", "Байты кучи, находящиеся в активных спанах.", "gauge", stats.HeapInuse)
	writeMetric("go_memstats_heap_idle_bytes", "Байты кучи в неиспользуемых спанах (готовы к переиспользованию или возврату ОС).", "gauge", stats.HeapIdle)
	writeMetric("go_memstats_heap_objects", "Количество объектов в куче.", "gauge", stats.HeapObjects)
	writeMetric("go_memstats_gc_sys_bytes", "Байты, используемые рантаймом для метаданных GC.", "gauge", stats.GCSys)
	writeMetric("go_memstats_next_gc_bytes", "Порог аллокаций памяти, при котором будет запущен следующий GC.", "gauge", stats.NextGC)
	writeMetric("go_memstats_last_gc_time_seconds", "Время (Unix timestamp в секундах) завершения последнего цикла GC.", "gauge", float64(stats.LastGC)/1e9)
	writeMetric("go_memstats_num_gc", "Количество завершённых циклов GC.", "counter", stats.NumGC)
	writeMetric("go_memstats_pause_total_ns", "Суммарное время пауз GC в наносекундах.", "counter", stats.PauseTotalNs)
	writeMetric("go_goroutines", "Текущее количество горутин.", "gauge", runtime.NumGoroutine())
	writeMetric("go_memstats_gc_percentage", "Процент увеличения объёма памяти кучи до следующего цикла GC", "gauge", getGCPercentage())
}

func getGCPercentage() float64 {
	gcPercentMetric := "/gc/gogc:percent"
	sample := make([]metrics.Sample, 1)
	sample[0].Name = gcPercentMetric
	metrics.Read(sample)

	if sample[0].Value.Kind() == metrics.KindUint64 {
		return float64(sample[0].Value.Uint64())
	}
	return 100
}

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	mux := http.NewServeMux()
	mux.HandleFunc("/metrics", metricsHandler)
	mux.Handle("/debug/pprof/", http.DefaultServeMux)

	server := &http.Server{Addr: ":8080", Handler: mux}

	fmt.Printf("Server started on http://localhost%s\n", server.Addr)
	fmt.Println("Metrics:      http://localhost:8080/metrics")
	fmt.Println("Profiling:    http://localhost:8080/debug/pprof/")

	go func() {
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			fmt.Printf("Server failed: %v\n", err)
		}
	}()

	<-ctx.Done()
	fmt.Println("Shutting down server...")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		fmt.Printf("Server forced to shutdown: %v", err)
	}
}
