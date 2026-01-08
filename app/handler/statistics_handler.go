package handler

import (
	"net/http"
	"strconv"

	"waverless/internal/model"
	"waverless/internal/service"
	"waverless/pkg/interfaces"
	"waverless/pkg/logger"

	"github.com/gin-gonic/gin"
)

// StatisticsHandler handles statistics-related HTTP requests
type StatisticsHandler struct {
	statsService  *service.StatisticsService
	workerService *service.WorkerService
}

// NewStatisticsHandler creates a new statistics handler
func NewStatisticsHandler(statsService *service.StatisticsService, workerService *service.WorkerService) *StatisticsHandler {
	return &StatisticsHandler{
		statsService:  statsService,
		workerService: workerService,
	}
}

// GetOverview retrieves global task statistics for dashboard
// @Summary Get global task statistics
// @Description Get aggregated task statistics across all endpoints
// @Tags statistics
// @Produce json
// @Success 200 {object} map[string]interface{} "Statistics data including total, pending, in_progress, completed, failed, cancelled counts"
// @Router /api/v1/statistics/overview [get]
func (h *StatisticsHandler) GetOverview(c *gin.Context) {
	stats, err := h.statsService.GetOverviewStatistics(c.Request.Context())
	if err != nil {
		logger.ErrorCtx(c.Request.Context(), "failed to get overview statistics: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"total":       stats.TotalCount,
		"pending":     stats.PendingCount,
		"in_progress": stats.InProgressCount,
		"completed":   stats.CompletedCount,
		"failed":      stats.FailedCount,
		"cancelled":   stats.CancelledCount,
		"updated_at":  stats.UpdatedAt,
	})
}

// GetEndpointStatistics retrieves task statistics for a specific endpoint
// @Summary Get endpoint task statistics
// @Description Get task statistics for a specific endpoint
// @Tags statistics
// @Produce json
// @Param endpoint path string true "Endpoint name"
// @Success 200 {object} map[string]interface{} "Endpoint statistics"
// @Router /api/v1/statistics/endpoints/{endpoint} [get]
func (h *StatisticsHandler) GetEndpointStatistics(c *gin.Context) {
	endpoint := c.Param("endpoint")
	if endpoint == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "endpoint required"})
		return
	}

	stats, err := h.statsService.GetEndpointStatistics(c.Request.Context(), endpoint)
	if err != nil {
		logger.ErrorCtx(c.Request.Context(), "failed to get endpoint statistics for %s: %v", endpoint, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Get worker statistics
	workers, _ := h.workerService.ListWorkers(c.Request.Context(), endpoint)
	onlineWorkers := 0
	busyWorkers := 0
	for _, w := range workers {
		switch w.Status {
		case model.WorkerStatusOnline, model.WorkerStatusDraining:
			onlineWorkers++
		case model.WorkerStatusBusy:
			onlineWorkers++
			busyWorkers++
		}
	}

	c.JSON(http.StatusOK, interfaces.EndpointStats{
		Endpoint:       endpoint,
		PendingTasks:   stats.PendingCount,
		RunningTasks:   stats.InProgressCount,
		CompletedTasks: stats.CompletedCount,
		FailedTasks:    stats.FailedCount,
		OnlineWorkers:  onlineWorkers,
		BusyWorkers:    busyWorkers,
	})
}

// GetTopEndpoints retrieves top N endpoints by task volume
// @Summary Get top endpoints by task volume
// @Description Get top endpoints sorted by total task count
// @Tags statistics
// @Produce json
// @Param limit query int false "Number of endpoints to return (default: 10, max: 50)"
// @Success 200 {object} map[string]interface{} "List of top endpoints with their statistics"
// @Router /api/v1/statistics/endpoints [get]
func (h *StatisticsHandler) GetTopEndpoints(c *gin.Context) {
	limit := 10
	if limitParam := c.Query("limit"); limitParam != "" {
		if parsedLimit, err := strconv.Atoi(limitParam); err == nil && parsedLimit > 0 {
			limit = parsedLimit
			if limit > 50 {
				limit = 50 // Cap at 50 to prevent excessive queries
			}
		}
	}

	stats, err := h.statsService.GetTopEndpointStatistics(c.Request.Context(), limit)
	if err != nil {
		logger.ErrorCtx(c.Request.Context(), "failed to get top endpoint statistics: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Transform to response format
	endpoints := make([]map[string]interface{}, 0, len(stats))
	for _, stat := range stats {
		endpoint := "unknown"
		if stat.ScopeValue != nil {
			endpoint = *stat.ScopeValue
		}

		endpoints = append(endpoints, map[string]interface{}{
			"endpoint":    endpoint,
			"total":       stat.TotalCount,
			"pending":     stat.PendingCount,
			"in_progress": stat.InProgressCount,
			"completed":   stat.CompletedCount,
			"failed":      stat.FailedCount,
			"cancelled":   stat.CancelledCount,
			"updated_at":  stat.UpdatedAt,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"endpoints": endpoints,
		"total":     len(endpoints),
	})
}
