package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/alireza-akbarzadeh/luxe/internal/services"
	"github.com/alireza-akbarzadeh/luxe/internal/shared/utils"
	"github.com/gin-gonic/gin"
)

type ImportHandler struct {
	importService services.ImportServiceInterface
}

func NewImportHandler(svc services.ImportServiceInterface) *ImportHandler {
	return &ImportHandler{importService: svc}
}

// ImportProducts imports products from an uploaded Excel file (admin only).
// @Summary      Import products from Excel (admin)
// @Description  Upload an .xlsx file. Row 1 must be the header (name,sku,price,stock,description,status,category_id,brand_id,store_id,compare_at_price,barcode,low_stock_threshold,weight). Returns per-row result.
// @Tags         Admin, Import
// @Accept       multipart/form-data
// @Produce      json
// @Security     BearerAuth
// @Param        file     formData  file  true  "Excel file (.xlsx)"
// @Param        store_id query     int   false "Default store ID for products without one in the file"
// @Success      200 {object} utils.Response{data=dto.ImportSummary}
// @Failure      400 {object} utils.Response
// @Failure      401 {object} utils.Response
// @Failure      403 {object} utils.Response
// @Failure      500 {object} utils.Response
// @Router       /admin/import/products [post]
func (ctrl *ImportHandler) ImportProducts(c *gin.Context) {
	fh, err := c.FormFile("file")
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "file is required (form field: file)")
		return
	}
	if fh.Size > 10<<20 { // 10 MB
		utils.ErrorResponse(c, http.StatusBadRequest, "file too large (max 10 MB)")
		return
	}

	f, err := fh.Open()
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "cannot open file")
		return
	}
	defer f.Close()

	var storeID uint
	if raw := c.Query("store_id"); raw != "" {
		if v, err := strconv.ParseUint(raw, 10, 64); err == nil {
			storeID = uint(v)
		}
	}

	summary, err := ctrl.importService.ImportProductsFromExcel(f, storeID)
	if err != nil {
		utils.HandleServiceError(c, err, "product import failed")
		return
	}
	utils.SuccessResponse(c, fmt.Sprintf("import complete: %d created, %d failed, %d skipped",
		summary.Created, summary.Failed, summary.Skipped), summary)
}

// ImportCategories imports categories from an uploaded Excel file (admin only).
// @Summary      Import categories from Excel (admin)
// @Description  Upload an .xlsx file. Row 1 must be the header (name,slug,description,parent_id,is_active). Returns per-row result.
// @Tags         Admin, Import
// @Accept       multipart/form-data
// @Produce      json
// @Security     BearerAuth
// @Param        file  formData  file  true  "Excel file (.xlsx)"
// @Success      200 {object} utils.Response{data=dto.ImportSummary}
// @Failure      400 {object} utils.Response
// @Failure      401 {object} utils.Response
// @Failure      403 {object} utils.Response
// @Failure      500 {object} utils.Response
// @Router       /admin/import/categories [post]
func (ctrl *ImportHandler) ImportCategories(c *gin.Context) {
	fh, err := c.FormFile("file")
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "file is required (form field: file)")
		return
	}
	if fh.Size > 10<<20 {
		utils.ErrorResponse(c, http.StatusBadRequest, "file too large (max 10 MB)")
		return
	}

	f, err := fh.Open()
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "cannot open file")
		return
	}
	defer f.Close()

	summary, err := ctrl.importService.ImportCategoriesFromExcel(f)
	if err != nil {
		utils.HandleServiceError(c, err, "category import failed")
		return
	}
	utils.SuccessResponse(c, fmt.Sprintf("import complete: %d created, %d failed, %d skipped",
		summary.Created, summary.Failed, summary.Skipped), summary)
}

// DownloadTemplate streams a blank Excel template (admin only).
// @Summary      Download import template (admin)
// @Description  Returns a pre-formatted .xlsx template for the given entity (products | categories).
// @Tags         Admin, Import
// @Produce      application/vnd.openxmlformats-officedocument.spreadsheetml.sheet
// @Security     BearerAuth
// @Param        entity  path  string  true  "Entity type: products | categories"
// @Success      200  {file}   binary
// @Failure      400  {object} utils.Response
// @Failure      401  {object} utils.Response
// @Failure      403  {object} utils.Response
// @Router       /admin/import/template/{entity} [get]
func (ctrl *ImportHandler) DownloadTemplate(c *gin.Context) {
	entity := c.Param("entity")
	var (
		data []byte
		err  error
	)
	switch entity {
	case "products":
		data, err = ctrl.importService.ProductTemplate()
	case "categories":
		data, err = ctrl.importService.CategoryTemplate()
	default:
		utils.ErrorResponse(c, http.StatusBadRequest, "unknown entity; use 'products' or 'categories'")
		return
	}
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "failed to generate template")
		return
	}

	filename := fmt.Sprintf("luxe_%s_template_%s.xlsx", entity, time.Now().UTC().Format("20060102"))
	c.Header("Content-Disposition", "attachment; filename="+filename)
	c.Data(http.StatusOK,
		"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
		data)
}
