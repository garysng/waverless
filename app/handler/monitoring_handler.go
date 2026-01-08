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

// GetMinuteStats returns minute-level statistics
// GET /v1/endpoints/:endpoint/stats/minute
func (h *MonitoringHandler) GetMinuteStats(c *gin.Context) {
	endpoint := c.Param("endpoint")
	if endpoint == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "endpoint is required"})
		return
	}

	// Parse time range (default: last 1 hour)
	to := time.Now()
	from := to.Add(-time.Hour)

	if fromStr := c.Query("from"); fromStr != "" {
		if t, err := time.Parse(time.RFC3339, fromStr); err == nil {
			from = t
		}
	}
	if toStr := c.Query("to"); toStr != "" {
		if t, err := time.Parse(time.RFC3339, toStr); err == nil {
			to = t
		}
	}

	stats, err := h.monitoringService.GetMinuteStats(c.Request.Context(), endpoint, from, to)
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
		"interval": "1m",
		"stats":    stats,
	})
}

// GetHourlyStats returns hourly statistics
// GET /v1/endpoints/:endpoint/stats/hourly
func (h *MonitoringHandler) GetHourlyStats(c *gin.Context) {
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
		}
	}
	if toStr := c.Query("to"); toStr != "" {
		if t, err := time.Parse(time.RFC3339, toStr); err == nil {
			to = t
		}
	}

	stats, err := h.monitoringService.GetHourlyStats(c.Request.Context(), endpoint, from, to)
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
		"interval": "1h",
		"stats":    stats,
	})
}

// GetDailyStats returns daily statistics
// GET /v1/endpoints/:endpoint/stats/daily
func (h *MonitoringHandler) GetDailyStats(c *gin.Context) {
	endpoint := c.Param("endpoint")
	if endpoint == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "endpoint is required"})
		return
	}

	// Parse time range (default: last 30 days)
	to := time.Now()
	from := to.AddDate(0, 0, -30)

	if fromStr := c.Query("from"); fromStr != "" {
		if t, err := time.Parse("2006-01-02", fromStr); err == nil {
			from = t
		}
	}
	if toStr := c.Query("to"); toStr != "" {
		if t, err := time.Parse("2006-01-02", toStr); err == nil {
			to = t
		}
	}

	stats, err := h.monitoringService.GetDailyStats(c.Request.Context(), endpoint, from, to)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"endpoint": endpoint,
		"time_range": gin.H{
			"from": from.Format("2006-01-02"),
			"to":   to.Format("2006-01-02"),
		},
		"interval": "1d",
		"stats":    stats,
	})
}
