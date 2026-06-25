package handlers

import (
	"github.com/alireza-akbarzadeh/luxe/internal/constants"
	"github.com/alireza-akbarzadeh/luxe/internal/interfaces/http/dto"
	"github.com/alireza-akbarzadeh/luxe/internal/interfaces/http/middleware"
	"github.com/alireza-akbarzadeh/luxe/internal/services"
	"github.com/alireza-akbarzadeh/luxe/internal/shared/utils"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type UploadHandler struct {
	uploadService services.UploadServiceInterface
	validate      *validator.Validate
}

func NewUploadHandler(uploadService services.UploadServiceInterface) *UploadHandler {
	return &UploadHandler{
		uploadService: uploadService,
		validate:      validator.New(),
	}
}

// GetUploadConfig returns whether R2 presigned uploads are enabled.
// @Summary      Upload configuration
// @Description  Returns direct-upload settings for the storefront (R2 presigned PUT).
// @Tags         Uploads
// @Produce      json
// @Success      200 {object} utils.Response{data=dto.UploadConfigResponse}
// @Router       /uploads/config [get]
func (ctrl *UploadHandler) GetUploadConfig(c *gin.Context) {
	utils.SuccessResponse(c, constants.MsgFetchSuccess, ctrl.uploadService.GetConfig())
}

// PresignUpload creates a presigned URL for uploading a file directly to R2.
// @Summary      Presign upload
// @Description  Returns a short-lived PUT URL; client uploads file bytes to upload_url with Content-Type header.
// @Tags         Uploads
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request body dto.PresignUploadRequest true "Upload metadata"
// @Success      200 {object} utils.Response{data=dto.PresignUploadResponse}
// @Failure      400 {object} utils.Response
// @Failure      401 {object} utils.Response
// @Router       /uploads/presign [post]
func (ctrl *UploadHandler) PresignUpload(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		utils.UnauthorizedResponse(c, constants.ErrUnauthorized)
		return
	}

	var req dto.PresignUploadRequest
	if !utils.BindAndValidate(c, &req, ctrl.validate) {
		return
	}

	result, err := ctrl.uploadService.CreatePresignedUpload(c.Request.Context(), userID, req)
	if err != nil {
		utils.HandleServiceError(c, err, "failed to create upload URL")
		return
	}

	utils.SuccessResponse(c, "upload URL created", result)
}
