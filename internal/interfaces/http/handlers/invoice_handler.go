package handlers

import (
	"fmt"
	"net/http"

	appinvoice "github.com/alireza-akbarzadeh/luxe/internal/application/invoice"
	"github.com/alireza-akbarzadeh/luxe/internal/constants"
	"github.com/alireza-akbarzadeh/luxe/internal/interfaces/http/dto"
	"github.com/alireza-akbarzadeh/luxe/internal/shared/utils"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type InvoiceHandler struct {
	commands *appinvoice.Commands
	queries  *appinvoice.Queries
	validate *validator.Validate
}

func NewInvoiceHandler(commands *appinvoice.Commands, queries *appinvoice.Queries) *InvoiceHandler {
	return &InvoiceHandler{
		commands: commands,
		queries:  queries,
		validate: validator.New(),
	}
}

// ListInvoicesAdmin lists all invoices (admin only).
// @Summary      List invoices (admin)
// @Tags         Invoices
// @Produce      json
// @Security     BearerAuth
// @Param        status    query string false "Filter by status"
// @Param        user_id   query int    false "Filter by user ID"
// @Param        order_id  query int    false "Filter by order ID"
// @Param        search    query string false "Search invoice #, order #, or customer"
// @Param        from_date query string false "Start date (RFC3339)"
// @Param        to_date   query string false "End date (RFC3339)"
// @Param        limit     query int    false "Items per page"
// @Param        offset    query int    false "Offset"
// @Success      200 {object} utils.Response{data=dto.AdminInvoiceListData}
// @Router       /admin/invoices [get]
func (ctrl *InvoiceHandler) ListInvoicesAdmin(c *gin.Context) {
	var filters dto.AdminInvoiceListFilters
	if err := c.ShouldBindQuery(&filters); err != nil {
		utils.ErrorResponse(c, 400, "invalid query parameters")
		return
	}
	filters.Limit, filters.Offset = paginationParams(c, constants.DefaultLimit)

	invoices, total, err := ctrl.queries.ListAdmin(c.Request.Context(), filters)
	if err != nil {
		RespondServiceError(c, err, "failed to list invoices")
		return
	}

	items := make([]dto.AdminInvoiceListItem, 0, len(invoices))
	for i := range invoices {
		items = append(items, dto.ToAdminInvoiceListItem(&invoices[i]))
	}
	utils.SuccessResponse(c, constants.MsgFetchSuccess, dto.AdminInvoiceListData{
		Invoices: items,
		Total:    total,
		Limit:    filters.Limit,
		Offset:   filters.Offset,
	})
}

// GetInvoiceAdmin returns a single invoice (admin only).
// @Summary      Get invoice by ID (admin)
// @Tags         Invoices
// @Produce      json
// @Security     BearerAuth
// @Param        id path int true "Invoice ID"
// @Success      200 {object} utils.Response{data=dto.InvoiceDetailResponse}
// @Router       /admin/invoices/{id} [get]
func (ctrl *InvoiceHandler) GetInvoiceAdmin(c *gin.Context) {
	invoiceID, ok := parseUintParam(c, "id")
	if !ok {
		return
	}

	invoice, err := ctrl.queries.GetByIDAdmin(c.Request.Context(), invoiceID)
	if err != nil {
		RespondServiceError(c, err, "failed to load invoice")
		return
	}
	utils.SuccessResponse(c, constants.MsgFetchSuccess, dto.ToInvoiceDetail(invoice))
}

// UpdateInvoiceStatus updates an invoice status (admin only).
// @Summary      Update invoice status (admin)
// @Tags         Invoices
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id      path int true "Invoice ID"
// @Param        request body dto.UpdateInvoiceStatusRequest true "Status update"
// @Success      200 {object} utils.Response
// @Router       /admin/invoices/{id}/status [put]
func (ctrl *InvoiceHandler) UpdateInvoiceStatus(c *gin.Context) {
	invoiceID, ok := parseUintParam(c, "id")
	if !ok {
		return
	}

	var req dto.UpdateInvoiceStatusRequest
	if !utils.BindAndValidate(c, &req, ctrl.validate) {
		return
	}

	if err := ctrl.commands.UpdateStatus(c.Request.Context(), invoiceID, req.Status); err != nil {
		RespondServiceError(c, err, "failed to update invoice status")
		return
	}
	utils.SuccessResponse(c, constants.MsgUpdateSuccess, nil)
}

// DownloadInvoicePDF returns a PDF document for an invoice (admin only).
// @Summary      Download invoice PDF (admin)
// @Tags         Invoices
// @Produce      application/pdf
// @Security     BearerAuth
// @Param        id path int true "Invoice ID"
// @Success      200 {file} binary
// @Router       /admin/invoices/{id}/pdf [get]
func (ctrl *InvoiceHandler) DownloadInvoicePDF(c *gin.Context) {
	invoiceID, ok := parseUintParam(c, "id")
	if !ok {
		return
	}

	data, filename, err := ctrl.queries.GeneratePDF(c.Request.Context(), invoiceID)
	if err != nil {
		RespondServiceError(c, err, "failed to generate invoice pdf")
		return
	}

	c.Header("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filename))
	c.Data(http.StatusOK, "application/pdf", data)
}

// SendInvoiceEmail emails the invoice summary to the customer (admin only).
// @Summary      Email invoice to customer (admin)
// @Tags         Invoices
// @Produce      json
// @Security     BearerAuth
// @Param        id path int true "Invoice ID"
// @Success      200 {object} utils.Response
// @Router       /admin/invoices/{id}/send [post]
func (ctrl *InvoiceHandler) SendInvoiceEmail(c *gin.Context) {
	invoiceID, ok := parseUintParam(c, "id")
	if !ok {
		return
	}

	if err := ctrl.commands.SendToCustomer(c.Request.Context(), invoiceID); err != nil {
		RespondServiceError(c, err, "failed to send invoice email")
		return
	}
	utils.SuccessResponse(c, "invoice email queued", nil)
}
