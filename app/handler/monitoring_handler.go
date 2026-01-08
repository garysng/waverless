package handler

import (
	"net/http"
	"time"

	"waverless/internal/service"

	"github.com/gin-gonic/gin"
)

// MonitoringHandler handles monitoring API requests
type MonitoringHandler struct {
	monitoringService *service.MonitoringService
}

// NewMonitoringHandler creates a new monitoring handler
func NewMonitoringHandler(monitoringService *service.MonitoringService) *MonitoringHandler {
	return &MonitoringHandler{monitoringService: monitoringService}
}

// GetRealtimeMetrics returns real-time metrics for an endpoint
// GET /v1/endpoints/:endpoint/metrics/realtime
func (h *MonitoringHandler) GetRealtimeMetrics(c *gin.Context) {
	endpoint := c.Param("endpoint")
	if endpoint == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "endpoint is required"})
		return
	}

	metrics, err := h.monitoringService.GetRealtimeMetrics(c.Request.Context(), endpoint)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"endpoint":    endpoint,
		"timestamp":   time.Now(),
		"workers":     metrics.Workers,
		"tasks":       metrics.Tasks,
		"performance": metrics.Performance,
	})
}

// GetStats returns statistics with auto-selected granularity based on time range
// GET /v1/endpoints/:endpoint/metrics/stats?from=xxx&to=xxx
// Granularity: ≤2h -> minute, ≤7d -> hourly, >7d -> daily
func (h *MonitoringHandler) GetStats(c *gin.Context) {
	endpoint := c.Param("endpoint")
	if endpoint == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "endpoint is required"})
		return
	}

	// Parse time range (default: last 24 hours)
	to := time.Now()
	from := to.Add(-24 * time.Hour)

	if fromStr := c.Query("from"); fromStr != "" {
		if t, err := time.Parse(time.RFC3339, fromStr); err == nil {
			from = t
		} else if t, err := time.Parse("2006-01-02", fromStr); err == nil {
			from = t
		}
	}
	if toStr := c.Query("to"); toStr != "" {
		if t, err := time.Parse(time.RFC3339, toStr); err == nil {
			to = t
		} else if t, err := time.Parse("2006-01-02", toStr); err == nil {
			to = t.Add(24*time.Hour - time.Second)
		}
	}

	duration := to.Sub(from)
	ctx := c.Request.Context()

	var stats interface{}
	var granularity string
	var err error

	switch {
	case duration <= 2*time.Hour:
		stats, err = h.monitoringService.GetMinuteStats(ctx, endpoint, from, to)
		granularity = "1m"
	case duration <= 7*24*time.Hour:
		stats, err = h.monitoringService.GetHourlyStats(ctx, endpoint, from, to)
		granularity = "1h"
	default:
		stats, err = h.monitoringService.GetDailyStats(ctx, endpoint, from, to)
		granularity = "1d"
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"endpoint": endpoint,
		"time_range": gin.H{
			"from": from,
			"to":   to,
		},
		"granularity": granularity,
		"stats":       stats,
	})
}
