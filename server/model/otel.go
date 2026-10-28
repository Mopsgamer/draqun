package model

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"runtime"
	"sync"
	"sync/atomic"
	"time"
)

type OtelSpan struct {
	TraceID    string            `json:"trace_id"`
	SpanID     string            `json:"span_id"`
	Name       string            `json:"name"`
	Kind       string            `json:"kind"`
	Status     string            `json:"status"`
	StatusCode int               `json:"status_code"`
	DurationMs float64           `json:"duration_ms"`
	Timestamp  string            `json:"timestamp"`
	Attributes map[string]string `json:"attributes"`
}

type OtelStats struct {
	UptimeSeconds    uint64     `json:"uptime_seconds"`
	TotalRequests    uint64     `json:"total_requests"`
	ErrorCount       uint64     `json:"error_count"`
	AvgLatencyMs     float64    `json:"avg_latency_ms"`
	ActiveGoroutines int        `json:"active_goroutines"`
	MemAllocMB       float64    `json:"mem_alloc_mb"`
	MemSysMB         float64    `json:"mem_sys_mb"`
	RecentSpans      []OtelSpan `json:"recent_spans"`
}

type otelCollector struct {
	mu            sync.RWMutex
	startTime     time.Time
	spans         []OtelSpan
	maxSpans      int
	totalRequests uint64
	totalLatency  uint64 // in microseconds
	errorCount    uint64
}

var GlobalOtelCollector = &otelCollector{
	startTime: time.Now(),
	spans:     make([]OtelSpan, 0, 100),
	maxSpans:  100,
}

func generateRandomHex(bytesLen int) string {
	b := make([]byte, bytesLen)
	_, err := rand.Read(b)
	if err != nil {
		return fmt.Sprintf("%d", time.Now().UnixNano())
	}
	return hex.EncodeToString(b)
}

func (c *otelCollector) RecordSpan(name, kind string, statusCode int, duration time.Duration, attributes map[string]string) {
	atomic.AddUint64(&c.totalRequests, 1)
	atomic.AddUint64(&c.totalLatency, uint64(duration.Microseconds()))

	status := "OK"
	if statusCode >= 400 {
		status = "ERROR"
		atomic.AddUint64(&c.errorCount, 1)
	}

	span := OtelSpan{
		TraceID:    generateRandomHex(16),
		SpanID:     generateRandomHex(8),
		Name:       name,
		Kind:       kind,
		Status:     status,
		StatusCode: statusCode,
		DurationMs: float64(duration.Microseconds()) / 1000.0,
		Timestamp:  time.Now().Format("15:04:05.000"),
		Attributes: attributes,
	}

	c.mu.Lock()
	defer c.mu.Unlock()
	if len(c.spans) >= c.maxSpans {
		c.spans = c.spans[1:]
	}
	c.spans = append(c.spans, span)
}

func (c *otelCollector) GetStats() OtelStats {
	c.mu.RLock()
	defer c.mu.RUnlock()

	reqs := atomic.LoadUint64(&c.totalRequests)
	totLat := atomic.LoadUint64(&c.totalLatency)
	errs := atomic.LoadUint64(&c.errorCount)

	var avgLat float64
	if reqs > 0 {
		avgLat = (float64(totLat) / 1000.0) / float64(reqs)
	}

	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	spansCopy := make([]OtelSpan, 0, len(c.spans))
	if len(c.spans) > 0 {
		spansCopy = make([]OtelSpan, len(c.spans))
		// Copy in reverse order so latest spans come first
		for i, s := range c.spans {
			spansCopy[len(c.spans)-1-i] = s
		}
	}

	return OtelStats{
		UptimeSeconds:    uint64(time.Since(c.startTime).Seconds()),
		TotalRequests:    reqs,
		ErrorCount:       errs,
		AvgLatencyMs:     avgLat,
		ActiveGoroutines: runtime.NumGoroutine(),
		MemAllocMB:       float64(m.Alloc) / (1024 * 1024),
		MemSysMB:         float64(m.Sys) / (1024 * 1024),
		RecentSpans:      spansCopy,
	}
}

func RecordOtelSpan(name, kind string, statusCode int, duration time.Duration, attributes map[string]string) {
	GlobalOtelCollector.RecordSpan(name, kind, statusCode, duration, attributes)
}

func GetOtelStats() OtelStats {
	return GlobalOtelCollector.GetStats()
}
