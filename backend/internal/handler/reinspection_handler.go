package handler

import (
	"log/slog"
	"net/http"

	"groundclearance/internal/constants"
	"groundclearance/internal/dto"
	"groundclearance/internal/middleware"
	"groundclearance/internal/service"

	"github.com/gin-gonic/gin"
)

type ReinspectionHandler struct {
	svc    *service.ReinspectionService
	logger *slog.Logger
}

func NewReinspectionHandler(svc *service.ReinspectionService, logger *slog.Logger) *ReinspectionHandler {
	return &ReinspectionHandler{svc: svc, logger: logger}
}

func (h *ReinspectionHandler) Register(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	var request dto.ReinspectionRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		Fail(c, http.StatusBadRequest, constants.CodeBadRequest, err.Error())
		return
	}
	record, err := h.svc.Register(id, middleware.GetUserID(c), request.Result, request.Evidence,
		request.Remark, requestAuditContext(c))
	if err != nil {
		handleServiceError(c, h.logger, err, "ground unit reinspection")
		return
	}
	c.Set("audit_persisted", true)
	OKWithMessage(c, "复检记录已登记", record)
}

func (h *ReinspectionHandler) ListByGroundUnit(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	records, err := h.svc.ListByGroundUnit(id)
	if err != nil {
		handleServiceError(c, h.logger, err, "ground unit reinspection list")
		return
	}
	OK(c, records)
}
