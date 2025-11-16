package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"waverless/internal/service"
	"waverless/pkg/interfaces"
	"waverless/pkg/logger"
)

// SpecHandler handles spec CRUD APIs
type SpecHandler struct {
	specService *service.SpecService
}

// NewSpecHandler creates a new spec handler
func NewSpecHandler(specService *service.SpecService) *SpecHandler {
	return &SpecHandler{
		specService: specService,
	}
}

// CreateSpec creates a new spec
// @Summary Create spec
// @Description Create a new resource specification
// @Tags Specs
// @Accept json
// @Produce json
// @Param request body interfaces.CreateSpecRequest true "Spec creation request"
// @Success 200 {object} interfaces.SpecInfo
// @Router /api/v1/k8s/specs [post]
func (h *SpecHandler) CreateSpec(c *gin.Context) {
	var req interfaces.CreateSpecRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.ErrorCtx(c.Request.Context(), "Failed to bind create spec request: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	logger.InfoCtx(c.Request.Context(), "Creating spec: name=%s, category=%s", req.Name, req.Category)

	spec, err := h.specService.CreateSpec(c.Request.Context(), &req)
	if err != nil {
		logger.ErrorCtx(c.Request.Context(), "Failed to create spec: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	logger.InfoCtx(c.Request.Context(), "Successfully created spec: %s", req.Name)
	c.JSON(http.StatusOK, spec)
}

// GetSpec gets spec details
// @Summary Get spec details
// @Description Get detailed information for specified spec
// @Tags Specs
// @Produce json
// @Param name path string true "Spec name"
// @Success 200 {object} interfaces.SpecInfo
// @Router /api/v1/k8s/specs/{name} [get]
func (h *SpecHandler) GetSpec(c *gin.Context) {
	name := c.Param("name")

	spec, err := h.specService.GetSpec(c.Request.Context(), name)
	if err != nil {
		logger.ErrorCtx(c.Request.Context(), "Failed to get spec: %v", err)
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, spec)
}

// ListSpecs lists all specs
// @Summary List all specs
// @Description Get all available resource specs
// @Tags Specs
// @Produce json
// @Param category query string false "Filter by category (cpu, gpu)"
// @Success 200 {array} interfaces.SpecInfo
// @Router /api/v1/k8s/specs [get]
func (h *SpecHandler) ListSpecs(c *gin.Context) {
	category := c.Query("category")

	var specs []*interfaces.SpecInfo
	var err error

	if category != "" {
		specs, err = h.specService.ListSpecsByCategory(c.Request.Context(), category)
	} else {
		specs, err = h.specService.ListSpecs(c.Request.Context())
	}

	if err != nil {
		logger.ErrorCtx(c.Request.Context(), "Failed to list specs: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, specs)
}

// UpdateSpec updates a spec
// @Summary Update spec
// @Description Update an existing resource specification
// @Tags Specs
// @Accept json
// @Produce json
// @Param name path string true "Spec name"
// @Param request body interfaces.UpdateSpecRequest true "Spec update request"
// @Success 200 {object} interfaces.SpecInfo
// @Router /api/v1/k8s/specs/{name} [put]
func (h *SpecHandler) UpdateSpec(c *gin.Context) {
	name := c.Param("name")

	var req interfaces.UpdateSpecRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.ErrorCtx(c.Request.Context(), "Failed to bind update spec request: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	logger.InfoCtx(c.Request.Context(), "Updating spec: name=%s", name)

	spec, err := h.specService.UpdateSpec(c.Request.Context(), name, &req)
	if err != nil {
		logger.ErrorCtx(c.Request.Context(), "Failed to update spec: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	logger.InfoCtx(c.Request.Context(), "Successfully updated spec: %s", name)
	c.JSON(http.StatusOK, spec)
}

// DeleteSpec deletes a spec
// @Summary Delete spec
// @Description Delete a resource specification (soft delete)
// @Tags Specs
// @Param name path string true "Spec name"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/k8s/specs/{name} [delete]
func (h *SpecHandler) DeleteSpec(c *gin.Context) {
	name := c.Param("name")

	logger.InfoCtx(c.Request.Context(), "Deleting spec: name=%s", name)

	if err := h.specService.DeleteSpec(c.Request.Context(), name); err != nil {
		logger.ErrorCtx(c.Request.Context(), "Failed to delete spec: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	logger.InfoCtx(c.Request.Context(), "Successfully deleted spec: %s", name)
	c.JSON(http.StatusOK, gin.H{
		"message": "Spec deleted successfully",
		"name":    name,
	})
}
