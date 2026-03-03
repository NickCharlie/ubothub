package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/NickCharlie/ubothub/backend/internal/dto/response"
	"github.com/NickCharlie/ubothub/backend/internal/model"
	"github.com/NickCharlie/ubothub/backend/internal/service"
	"github.com/NickCharlie/ubothub/backend/pkg/errcode"
	resp "github.com/NickCharlie/ubothub/backend/pkg/response"
)

// Ensure dto/response import for Swagger type resolution.
var _ response.CommonResponse

// LegalHandler handles legal agreement HTTP endpoints.
type LegalHandler struct {
	legalSvc *service.LegalService
}

// NewLegalHandler creates a new legal handler.
func NewLegalHandler(legalSvc *service.LegalService) *LegalHandler {
	return &LegalHandler{legalSvc: legalSvc}
}

// GetTermsOfService handles GET /api/v1/legal/terms.
// @Summary Get terms of service
// @Description Returns the active terms of service for the given locale.
// @Tags Legal
// @Produce json
// @Param locale query string false "Locale" default(en)
// @Success 200 {object} response.CommonResponse
// @Failure 404 {object} response.CommonResponse
// @Router /legal/terms [get]
func (h *LegalHandler) GetTermsOfService(c *gin.Context) {
	locale := c.DefaultQuery("locale", "en")
	agreement, err := h.legalSvc.GetActiveAgreement(c.Request.Context(), model.AgreementTypeTerms, locale)
	if err != nil {
		resp.Error(c, errcode.ErrNotFound)
		return
	}
	resp.OK(c, toAgreementResponse(agreement))
}

// GetPrivacyPolicy handles GET /api/v1/legal/privacy.
// @Summary Get privacy policy
// @Description Returns the active privacy policy for the given locale.
// @Tags Legal
// @Produce json
// @Param locale query string false "Locale" default(en)
// @Success 200 {object} response.CommonResponse
// @Failure 404 {object} response.CommonResponse
// @Router /legal/privacy [get]
func (h *LegalHandler) GetPrivacyPolicy(c *gin.Context) {
	locale := c.DefaultQuery("locale", "en")
	agreement, err := h.legalSvc.GetActiveAgreement(c.Request.Context(), model.AgreementTypePrivacy, locale)
	if err != nil {
		resp.Error(c, errcode.ErrNotFound)
		return
	}
	resp.OK(c, toAgreementResponse(agreement))
}

// GetAllAgreements handles GET /api/v1/legal/agreements.
// @Summary Get all active agreements
// @Description Returns all active agreements grouped by type.
// @Tags Legal
// @Produce json
// @Success 200 {object} response.CommonResponse
// @Router /legal/agreements [get]
func (h *LegalHandler) GetAllAgreements(c *gin.Context) {
	terms, err := h.legalSvc.GetAllActiveAgreements(c.Request.Context(), model.AgreementTypeTerms)
	if err != nil {
		resp.Error(c, errcode.ErrInternalServer)
		return
	}
	privacy, err := h.legalSvc.GetAllActiveAgreements(c.Request.Context(), model.AgreementTypePrivacy)
	if err != nil {
		resp.Error(c, errcode.ErrInternalServer)
		return
	}

	result := gin.H{
		"terms_of_service": toAgreementList(terms),
		"privacy_policy":   toAgreementList(privacy),
	}
	resp.OK(c, result)
}

func toAgreementResponse(a *model.LegalAgreement) gin.H {
	return gin.H{
		"id":         a.ID,
		"type":       a.Type,
		"version":    a.Version,
		"locale":     a.Locale,
		"title":      a.Title,
		"content":    a.Content,
		"updated_at": a.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}
}

func toAgreementList(agreements []model.LegalAgreement) []gin.H {
	result := make([]gin.H, 0, len(agreements))
	for i := range agreements {
		result = append(result, toAgreementResponse(&agreements[i]))
	}
	return result
}
