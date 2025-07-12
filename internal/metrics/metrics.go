package metrics

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	MessagesProcessed = promauto.NewCounter(prometheus.CounterOpts{
		Name: "bot_messages_total",
		Help: "Total number of processed messages",
	})

	APIErrors = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "bot_api_errors_total",
		Help: "Total API errors by type",
	}, []string{"type"})

	ResponseTime = promauto.NewHistogram(prometheus.HistogramOpts{
		Name:    "bot_response_time_ms",
		Help:    "Bot response time in milliseconds",
		Buckets: []float64{50, 100, 200, 500, 1000},
	})

	ActiveUsers = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "bot_active_users",
		Help: "Number of active users",
	})
)

func Init() {
	ActiveUsers.Set(0)
	MessagesProcessed.Add(0)
	prometheus.MustRegister(ActiveUsers, MessagesProcessed)
}

func TrackResponseTime(start time.Time) {
	ResponseTime.Observe(float64(time.Since(start).Milliseconds()))
}

func StartMetricsServer() {
	http.Handle("/metrics", promhttp.Handler())
	go func() {
		if err := http.ListenAndServe(":2112", nil); err != nil {
			panic(fmt.Sprintf("Metrics server failed: %v", err))
		}
	}()
	log.Println("Metrics server started on :2112")
}
