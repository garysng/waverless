package handler

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	endpointsvc "waverless/internal/service/endpoint"
	"waverless/pkg/config"
	"waverless/pkg/image"
	"waverless/pkg/logger"
	"waverless/pkg/notification"
)

// ImageHandler handles image-related operations
type ImageHandler struct {
	endpointService *endpointsvc.Service
	imageChecker    *image.Checker
	notifier        *notification.FeishuNotifier
}

// NewImageHandler creates a new image handler
func NewImageHandler(endpointService *endpointsvc.Service, dockerConfig *config.DockerConfig) *ImageHandler {
	return &ImageHandler{
		endpointService: endpointService,
		imageChecker:    image.NewChecker(dockerConfig),
		notifier:        notification.NewFeishuNotifier(),
	}
}

// DockerHubWebhookPayload represents the payload from DockerHub webhook
type DockerHubWebhookPayload struct {
	PushData struct {
		Tag      string `json:"tag"`
		PushedAt int64  `json:"pushed_at"`
		Pusher   string `json:"pusher"`
	} `json:"push_data"`
	Repository struct {
		Name      string `json:"name"`
		Namespace string `json:"namespace"`
		RepoName  string `json:"repo_name"`
		Status    string `json:"status"`
	} `json:"repository"`
}

// DockerHubWebhook handles DockerHub webhook notifications
// @Summary DockerHub webhook
// @Description Receive DockerHub webhook notifications for image updates
// @Tags Webhooks
// @Accept json
// @Produce json
// @Param payload body DockerHubWebhookPayload true "DockerHub webhook payload"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/webhooks/dockerhub [post]
func (h *ImageHandler) DockerHubWebhook(c *gin.Context) {
	ctx := c.Request.Context()

	var payload DockerHubWebhookPayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		logger.ErrorCtx(ctx, "[Image Webhook] Failed to parse DockerHub webhook payload: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
		return
	}

	logger.InfoCtx(ctx, "[Image Webhook] Received DockerHub webhook: repo=%s, tag=%s",
		payload.Repository.RepoName, payload.PushData.Tag)

	// Construct full image name
	fullImageName := fmt.Sprintf("%s:%s", payload.Repository.RepoName, payload.PushData.Tag)

	// Get all endpoints
	endpoints, err := h.endpointService.ListEndpoints(ctx)
	if err != nil {
		logger.ErrorCtx(ctx, "[Image Webhook] Failed to list endpoints: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list endpoints"})
		return
	}

	affectedEndpoints := []string{}
	notifiedCount := 0

	// Check which endpoints are affected by this image update
	for _, endpoint := range endpoints {
		// Check if the endpoint's image prefix matches the pushed image
		if endpoint.ImagePrefix != "" && strings.HasPrefix(fullImageName, endpoint.ImagePrefix) {
			// Extract tag from fullImageName
			_, newTag := image.ParseImageName(fullImageName)
			// Extract prefix from endpoint.ImagePrefix to get tagPrefix
			prefixParts := strings.Split(endpoint.ImagePrefix, ":")
			if len(prefixParts) != 2 {
				logger.WarnCtx(ctx, "[Image Webhook] Invalid image prefix format for endpoint %s: %s", endpoint.Name, endpoint.ImagePrefix)
				continue
			}
			tagPrefix := prefixParts[1]

			// Validate that the suffix after prefix is exactly 12 digits (YYYYMMDDHHMM)
			suffix := strings.TrimPrefix(newTag, tagPrefix)
			if len(suffix) != 12 {
				logger.InfoCtx(ctx, "[Image Webhook] Skipping endpoint %s: tag suffix length is %d, expected 12 (YYYYMMDDHHMM). Tag: %s",
					endpoint.Name, len(suffix), newTag)
				continue
			}

			// Validate all characters are digits
			allDigits := true
			for _, c := range suffix {
				if c < '0' || c > '9' {
					allDigits = false
					break
				}
			}
			if !allDigits {
				logger.InfoCtx(ctx, "[Image Webhook] Skipping endpoint %s: tag suffix contains non-digit characters. Tag: %s",
					endpoint.Name, newTag)
				continue
			}

			logger.InfoCtx(ctx, "[Image Webhook] Endpoint %s matches image prefix with valid date format: prefix=%s, newImage=%s, currentImage=%s",
				endpoint.Name, endpoint.ImagePrefix, fullImageName, endpoint.Image)

			// Get the new image digest
			repository, tag := image.ParseImageName(fullImageName)
			newDigest, err := h.imageChecker.GetImageDigest(ctx, repository, tag)
			if err != nil {
				logger.ErrorCtx(ctx, "[Image Webhook] Failed to get digest for image %s: %v", fullImageName, err)
				continue
			}

			// Get current image digest if not set
			currentDigest := endpoint.ImageDigest
			if currentDigest == "" {
				currentRepo, currentTag := image.ParseImageName(endpoint.Image)
				digest, digestErr := h.imageChecker.GetImageDigest(ctx, currentRepo, currentTag)
				if digestErr != nil {
					logger.ErrorCtx(ctx, "[Image Webhook] Failed to get current image digest for %s: %v", endpoint.Image, digestErr)
					continue
				}
				currentDigest = digest
			}

			// Check if this is actually a newer image (compare by name or digest)
			hasUpdate := (fullImageName != endpoint.Image) || (newDigest != currentDigest)

			if !hasUpdate {
				logger.InfoCtx(ctx, "[Image Webhook] Endpoint %s - No update needed, new image is same as current (digest match)",
					endpoint.Name)
				continue
			}

			// If using imagePrefix, verify this is the latest by comparing lexicographically
			if endpoint.ImagePrefix != "" && fullImageName < endpoint.Image {
				logger.InfoCtx(ctx, "[Image Webhook] Endpoint %s - Skipping older image: new=%s < current=%s",
					endpoint.Name, fullImageName, endpoint.Image)
				continue
			}

			logger.InfoCtx(ctx, "[Image Webhook] Endpoint %s - Update detected: current=%s, new=%s",
				endpoint.Name, endpoint.Image, fullImageName)

			affectedEndpoints = append(affectedEndpoints, endpoint.Name)

			// Mark the endpoint as having an update available
			now := time.Now()
			endpoint.ImageDigest = newDigest
			endpoint.ImageLastChecked = &now
			endpoint.LatestImage = fullImageName

			if err := h.endpointService.SaveEndpoint(ctx, endpoint); err != nil {
				logger.ErrorCtx(ctx, "[Image Webhook] Failed to update endpoint %s: %v", endpoint.Name, err)
				continue
			}

			// Send Feishu notification
			notification := &notification.ImageUpdateNotification{
				Endpoint:      endpoint.Name,
				CurrentImage:  endpoint.Image,
				NewImageTag:   fullImageName,
				ImagePrefix:   endpoint.ImagePrefix,
				DetectedAt:    time.Now(),
				DetectionType: "webhook",
			}

			if err := h.notifier.SendImageUpdateNotification(ctx, notification); err != nil {
				logger.ErrorCtx(ctx, "[Image Webhook] Failed to send Feishu notification for endpoint %s: %v", endpoint.Name, err)
			} else {
				notifiedCount++
			}
		}
	}

	logger.InfoCtx(ctx, "[Image Webhook] DockerHub webhook processed: affectedEndpoints=%d, notified=%d",
		len(affectedEndpoints), notifiedCount)

	c.JSON(http.StatusOK, gin.H{
		"message":           "webhook processed",
		"affectedEndpoints": affectedEndpoints,
		"notifiedCount":     notifiedCount,
	})
}

// CheckImageUpdate checks for image updates for a specific endpoint
// @Summary Check image update for endpoint
// @Description Manually check if there's a new image version available for an endpoint
// @Tags Endpoints
// @Produce json
// @Param name path string true "Endpoint name"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/endpoints/{name}/check-image [post]
func (h *ImageHandler) CheckImageUpdate(c *gin.Context) {
	ctx := c.Request.Context()
	name := c.Param("name")

	endpoint, err := h.endpointService.GetEndpoint(ctx, name)
	if err != nil || endpoint == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "endpoint not found"})
		return
	}

	// Get current image digest if not set
	currentDigest := endpoint.ImageDigest
	currentImage := endpoint.Image
	if currentDigest == "" {
		repository, tag := image.ParseImageName(endpoint.Image)
		digest, digestErr := h.imageChecker.GetImageDigest(ctx, repository, tag)
		if digestErr != nil {
			logger.ErrorCtx(ctx, "Failed to get current image digest for %s: %v", endpoint.Image, digestErr)
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": fmt.Sprintf("failed to get current image digest: %v", digestErr),
			})
			return
		}
		currentDigest = digest

		// Update the endpoint with the current digest
		endpoint.ImageDigest = currentDigest
		now := time.Now()
		endpoint.ImageLastChecked = &now
		if saveErr := h.endpointService.SaveEndpoint(ctx, endpoint); saveErr != nil {
			logger.WarnCtx(ctx, "Failed to save current digest for endpoint %s: %v", name, saveErr)
		}
	}

	// Check for updates using imagePrefix if configured
	var hasUpdate bool
	var newImage string
	var newDigest string
	var checkErr error

	if endpoint.ImagePrefix != "" {
		// Use imagePrefix to find the latest matching image
		logger.InfoCtx(ctx, "Checking for latest image with prefix: %s", endpoint.ImagePrefix)
		newImage, newDigest, checkErr = h.imageChecker.GetLatestImageByPrefix(ctx, endpoint.ImagePrefix)
		if checkErr != nil {
			logger.ErrorCtx(ctx, "Failed to get latest image by prefix for endpoint %s: %v", name, checkErr)
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": fmt.Sprintf("failed to get latest image by prefix: %v", checkErr),
			})
			return
		}

		// Check if the latest image is different from current
		hasUpdate = (newImage != currentImage) || (newDigest != currentDigest)
		logger.InfoCtx(ctx, "Latest image: %s (digest: %s), Current: %s (digest: %s), hasUpdate: %v",
			newImage, newDigest[:19], currentImage, currentDigest[:19], hasUpdate)
	} else {
		// Fallback: check if current image tag has been updated (digest changed)
		logger.InfoCtx(ctx, "No imagePrefix configured, checking current image digest")
		hasUpdate, newDigest, checkErr = h.imageChecker.CheckImageUpdate(ctx, endpoint.Image, currentDigest)
		newImage = endpoint.Image // Same image name
		if checkErr != nil {
			logger.ErrorCtx(ctx, "Failed to check image update for endpoint %s: %v", name, checkErr)
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": fmt.Sprintf("failed to check image update: %v", checkErr),
			})
			return
		}
	}

	// Update endpoint metadata
	now := time.Now()
	endpoint.ImageDigest = newDigest
	endpoint.ImageLastChecked = &now
	if hasUpdate {
		endpoint.LatestImage = newImage
	} else {
		endpoint.LatestImage = "" // Clear if no update
	}

	if err := h.endpointService.SaveEndpoint(ctx, endpoint); err != nil {
		logger.ErrorCtx(ctx, "Failed to update endpoint %s: %v", name, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update endpoint"})
		return
	}

	// If there's an update, send notification
	if hasUpdate {
		notif := &notification.ImageUpdateNotification{
			Endpoint:      endpoint.Name,
			CurrentImage:  currentImage,
			NewImageTag:   newImage,
			ImagePrefix:   endpoint.ImagePrefix,
			DetectedAt:    time.Now(),
			DetectionType: "manual",
		}

		if err := h.notifier.SendImageUpdateNotification(ctx, notif); err != nil {
			logger.ErrorCtx(ctx, "Failed to send Feishu notification for endpoint %s: %v", name, err)
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"endpoint":        name,
		"currentImage":    currentImage,
		"currentDigest":   currentDigest,
		"newImage":        newImage,
		"newDigest":       newDigest,
		"updateAvailable": hasUpdate,
		"lastChecked":     endpoint.ImageLastChecked,
	})
}

// CheckAllImagesUpdate checks for image updates for all endpoints
// @Summary Check image updates for all endpoints
// @Description Manually check if there are new image versions available for all endpoints
// @Tags Endpoints
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/endpoints/check-images [post]
func (h *ImageHandler) CheckAllImagesUpdate(c *gin.Context) {
	ctx := c.Request.Context()

	endpoints, err := h.endpointService.ListEndpoints(ctx)
	if err != nil {
		logger.ErrorCtx(ctx, "Failed to list endpoints: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list endpoints"})
		return
	}

	results := []map[string]interface{}{}
	updatesFound := 0
	notifiedCount := 0

	for _, endpoint := range endpoints {
		// Get current image digest if not set
		currentDigest := endpoint.ImageDigest
		currentImage := endpoint.Image
		if currentDigest == "" {
			repository, tag := image.ParseImageName(endpoint.Image)
			digest, err := h.imageChecker.GetImageDigest(ctx, repository, tag)
			if err != nil {
				logger.ErrorCtx(ctx, "Failed to get current image digest for %s: %v", endpoint.Image, err)
				results = append(results, map[string]interface{}{
					"endpoint": endpoint.Name,
					"error":    fmt.Sprintf("failed to get current digest: %v", err),
				})
				continue
			}
			currentDigest = digest
		}

		// Check for updates using imagePrefix if configured
		var hasUpdate bool
		var newImage string
		var newDigest string
		var checkErr error

		if endpoint.ImagePrefix != "" {
			// Use imagePrefix to find the latest matching image
			logger.InfoCtx(ctx, "Checking for latest image with prefix: %s for endpoint: %s", endpoint.ImagePrefix, endpoint.Name)
			newImage, newDigest, checkErr = h.imageChecker.GetLatestImageByPrefix(ctx, endpoint.ImagePrefix)
			if checkErr != nil {
				logger.ErrorCtx(ctx, "Failed to get latest image by prefix for endpoint %s: %v", endpoint.Name, checkErr)
				results = append(results, map[string]interface{}{
					"endpoint": endpoint.Name,
					"error":    fmt.Sprintf("failed to get latest image by prefix: %v", checkErr),
				})
				continue
			}

			// Check if the latest image is different from current
			hasUpdate = (newImage != currentImage) || (newDigest != currentDigest)
			logger.InfoCtx(ctx, "Endpoint %s - Latest image: %s, Current: %s, hasUpdate: %v",
				endpoint.Name, newImage, currentImage, hasUpdate)
		} else {
			// Fallback: check if current image tag has been updated (digest changed)
			logger.InfoCtx(ctx, "No imagePrefix configured for endpoint %s, checking current image digest", endpoint.Name)
			hasUpdate, newDigest, checkErr = h.imageChecker.CheckImageUpdate(ctx, endpoint.Image, currentDigest)
			newImage = endpoint.Image // Same image name
			if checkErr != nil {
				logger.ErrorCtx(ctx, "Failed to check image update for endpoint %s: %v", endpoint.Name, checkErr)
				results = append(results, map[string]interface{}{
					"endpoint": endpoint.Name,
					"error":    fmt.Sprintf("failed to check update: %v", checkErr),
				})
				continue
			}
		}

		// Update endpoint metadata
		now := time.Now()
		endpoint.ImageDigest = newDigest
		endpoint.ImageLastChecked = &now
		if hasUpdate {
			endpoint.LatestImage = newImage
		} else {
			endpoint.LatestImage = ""
		}

		if err := h.endpointService.SaveEndpoint(ctx, endpoint); err != nil {
			logger.ErrorCtx(ctx, "Failed to update endpoint %s: %v", endpoint.Name, err)
		}

		results = append(results, map[string]interface{}{
			"endpoint":        endpoint.Name,
			"currentImage":    currentImage,
			"newImage":        newImage,
			"updateAvailable": hasUpdate,
			"currentDigest":   currentDigest,
			"newDigest":       newDigest,
		})

		if hasUpdate {
			updatesFound++

			// Send notification
			notification := &notification.ImageUpdateNotification{
				Endpoint:      endpoint.Name,
				CurrentImage:  currentImage,
				NewImageTag:   newImage,
				ImagePrefix:   endpoint.ImagePrefix,
				DetectedAt:    time.Now(),
				DetectionType: "manual",
			}

			if err := h.notifier.SendImageUpdateNotification(ctx, notification); err != nil {
				logger.ErrorCtx(ctx, "Failed to send Feishu notification for endpoint %s: %v", endpoint.Name, err)
			} else {
				notifiedCount++
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"message":       "image check completed",
		"totalChecked":  len(endpoints),
		"updatesFound":  updatesFound,
		"notifiedCount": notifiedCount,
		"results":       results,
	})
}
