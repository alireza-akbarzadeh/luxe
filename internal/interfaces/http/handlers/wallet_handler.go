package handlers

import (
	"net/http"

	appwallet "github.com/alireza-akbarzadeh/luxe/internal/application/wallet"
	"github.com/alireza-akbarzadeh/luxe/internal/constants"
	"github.com/alireza-akbarzadeh/luxe/internal/interfaces/http/dto"
	"github.com/alireza-akbarzadeh/luxe/internal/interfaces/http/middleware"
	"github.com/alireza-akbarzadeh/luxe/internal/shared/utils"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type WalletHandler struct {
	walletService *appwallet.Service
	validate      *validator.Validate
}

func NewWalletHandler(walletService *appwallet.Service) *WalletHandler {
	return &WalletHandler{
		walletService: walletService,
		validate:      validator.New(),
	}
}

// GetWallet returns the wallet balance and transaction history.
// @Summary      Get wallet details
// @Description  Returns the balance and paginated transaction history of the authenticated user.
// @Tags         Wallet
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        limit  query int false "Items per page (default 20, max 100)" default(20)
// @Param        offset query int false "Offset" default(0)
// @Success      200 {object} utils.Response{data=dto.WalletDetailResponse}
// @Failure      401 {object} utils.Response
// @Failure      500 {object} utils.Response
// @Router       /wallet [get]
func (ctrl *WalletHandler) GetWallet(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		utils.UnauthorizedResponse(c, constants.ErrorUnauthorized)
		return
	}
	limit, offset := paginationParams(c, constants.DefaultLimit)
	filters := dto.WalletListFilters{
		Offset: offset,
		Limit:  limit,
	}
	balance, err := ctrl.walletService.GetBalance(c.Request.Context(), userID)
	if err != nil {
		utils.HandleServiceError(c, err, "failed to get wallet balance")
		return
	}
	transactions, total, err := ctrl.walletService.GetTransactions(c.Request.Context(), userID, filters)
	txResponses := make([]dto.TransactionResponse, len(transactions))
	for i, tx := range transactions {
		txResponses[i] = dto.TransactionResponse{
			ID:              tx.ID,
			CreatedAt:       tx.CreatedAt,
			Amount:          tx.Amount,
			Type:            tx.Type,
			ReferenceType:   tx.ReferenceType,
			ReferenceID:     tx.ReferenceID,
			Description:     tx.Description,
			BalanceAfter:    tx.BalanceAfter,
			Status:          tx.Status,
			StripeSessionID: tx.StripeSessionID,
		}
	}
	data := dto.WalletDetailResponse{
		Balance:      balance,
		Currency:     "USD",
		Transactions: txResponses,
		Total:        total,
		Limit:        limit,
		Offset:       offset,
	}
	utils.SuccessResponse(c, constants.MsgFetchSuccess, data)
}

// Deposit initiates a wallet deposit via Stripe Checkout (or mock confirm when Stripe is disabled).
// @Summary      Deposit funds
// @Description  Creates a pending deposit and returns a Stripe Checkout URL when Stripe is enabled.
// @Tags         Wallet
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request body dto.DepositRequest true "Deposit amount"
// @Success      200 {object} utils.Response{data=dto.DepositResponse}
// @Failure      400 {object} utils.Response
// @Failure      401 {object} utils.Response
// @Failure      500 {object} utils.Response
// @Router       /wallet/deposit [post]
func (ctrl *WalletHandler) Deposit(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		utils.UnauthorizedResponse(c, constants.ErrorUnauthorized)
		return
	}
	var req dto.DepositRequest
	if !utils.BindAndValidate(c, &req, ctrl.validate) {
		return
	}

	customerEmail := ""
	if emailVal, exists := c.Get("user_email"); exists {
		customerEmail, _ = emailVal.(string)
	}

	result, err := ctrl.walletService.InitiateDeposit(c.Request.Context(), userID, req.Amount, customerEmail)
	if err != nil {
		utils.HandleServiceError(c, err, "failed to deposit")
		return
	}

	message := "deposit completed"
	if result.Status == constants.WalletTxStatusPending {
		message = "deposit initiated — complete payment at checkout_url"
	}
	utils.SuccessResponse(c, message, result)
}

// ConfirmStripeDeposit confirms a wallet deposit after returning from Stripe Checkout.
// @Summary      Confirm Stripe wallet deposit
// @Description  Idempotent wallet credit using checkout session_id from the Stripe success redirect.
// @Tags         Wallet
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request body dto.ConfirmWalletDepositRequest true "Stripe session ID"
// @Success      200 {object} utils.Response{data=dto.ConfirmWalletDepositResponse}
// @Failure      400 {object} utils.Response
// @Failure      401 {object} utils.Response
// @Failure      403 {object} utils.Response
// @Router       /wallet/deposit/confirm-stripe [post]
func (ctrl *WalletHandler) ConfirmStripeDeposit(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		utils.UnauthorizedResponse(c, constants.ErrorUnauthorized)
		return
	}

	var req dto.ConfirmWalletDepositRequest
	if !utils.BindAndValidate(c, &req, ctrl.validate) {
		return
	}

	result, err := ctrl.walletService.ConfirmDepositBySessionID(c.Request.Context(), userID, req.SessionID)
	if err != nil {
		utils.HandleServiceError(c, err, "failed to confirm wallet deposit")
		return
	}

	utils.SuccessResponse(c, "deposit added to your wallet", result)
}

// AdminAdjust adjusts a user's wallet balance (admin only).
// @Summary      Admin adjust wallet
// @Description  Increase or decrease any user's wallet balance. Requires admin role.
// @Tags         Admin Wallet
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request body dto.AdminAdjustRequest true "Adjustment data"
// @Success      200 {object} utils.Response
// @Failure      400 {object} utils.Response
// @Failure      401 {object} utils.Response
// @Failure      403 {object} utils.Response
// @Failure      500 {object} utils.Response
// @Router       /admin/wallet/adjust [post]
func (ctrl *WalletHandler) AdminAdjust(c *gin.Context) {
	var req dto.AdminAdjustRequest
	if !utils.BindAndValidate(c, &req, ctrl.validate) {
		return
	}
	if err := ctrl.walletService.AdminAdjust(c.Request.Context(), req.UserID, req.Amount, req.Description); err != nil {
		utils.HandleServiceError(c, err, "adjustment failed")
		return
	}
	utils.SuccessResponse(c, "wallet adjusted successfully", nil)
}

// Withdraw requests a withdrawal from the wallet.
// @Summary      Withdraw funds
// @Description  Withdraw money from wallet. In production, this would require admin approval or integration with payout gateway.
// @Tags         Wallet
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request body dto.WithdrawRequest true "Withdrawal amount"
// @Success      200 {object} utils.Response{data=object{transaction_id=uint}}
// @Failure      400 {object} utils.Response
// @Failure      401 {object} utils.Response
// @Failure      500 {object} utils.Response
// @Router       /wallet/withdraw [post]
func (ctrl *WalletHandler) Withdraw(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		utils.UnauthorizedResponse(c, constants.ErrorUnauthorized)
		return
	}
	var req dto.WithdrawRequest
	if !utils.BindAndValidate(c, &req, ctrl.validate) {
		return
	}
	err := ctrl.walletService.Withdraw(c.Request.Context(), userID, req.Amount, "user_withdrawal", nil, req.Description)
	if err != nil {
		utils.HandleServiceError(c, err, "withdrawal failed")
		return
	}
	utils.SuccessResponse(c, "withdrawal successful", nil)
}

// GetTransaction returns a single wallet transaction by ID.
// @Summary      Get transaction details
// @Description  Fetch details of a specific wallet transaction.
// @Tags         Wallet
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      int  true  "Transaction ID"
// @Success      200  {object}  utils.Response{data=dto.TransactionResponse}
// @Failure      400  {object}  utils.Response
// @Failure      401  {object}  utils.Response
// @Failure      404  {object}  utils.Response
// @Router       /wallet/transactions/{id} [get]
func (ctrl *WalletHandler) GetTransaction(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		utils.UnauthorizedResponse(c, constants.ErrorUnauthorized)
		return
	}
	txID, ok := parseUintParam(c, "id")
	if !ok {
		return
	}
	tx, err := ctrl.walletService.GetTransaction(c.Request.Context(), userID, txID)
	if err != nil {
		utils.HandleServiceError(c, err, "failed to fetch transaction")
		return
	}
	resp := dto.TransactionResponse{
		ID:              tx.ID,
		CreatedAt:       tx.CreatedAt,
		Amount:          tx.Amount,
		Type:            tx.Type,
		ReferenceType:   tx.ReferenceType,
		ReferenceID:     tx.ReferenceID,
		Description:     tx.Description,
		BalanceAfter:    tx.BalanceAfter,
		Status:          tx.Status,
		StripeSessionID: tx.StripeSessionID,
	}
	utils.SuccessResponse(c, constants.MsgFetchSuccess, resp)
}

// CancelPendingDeposit cancels a pending deposit transaction.
// @Summary      Cancel pending deposit
// @Description  Cancels a deposit that is still in pending status.
// @Tags         Wallet
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      int  true  "Transaction ID"
// @Success      200  {object}  utils.Response
// @Failure      400  {object}  utils.Response
// @Failure      401  {object}  utils.Response
// @Failure      404  {object}  utils.Response
// @Router       /wallet/deposit/{id}/cancel [post]
func (ctrl *WalletHandler) CancelPendingDeposit(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		utils.UnauthorizedResponse(c, constants.ErrUnauthorized)
		return
	}

	txID, ok := parseUintParam(c, "id")
	if !ok {
		return
	}

	if err := ctrl.walletService.CancelPendingDeposit(c.Request.Context(), userID, txID); err != nil {
		utils.HandleServiceError(c, err, "failed to cancel deposit")
		return
	}

	utils.SuccessResponse(c, "deposit cancelled", nil)
}

// ListWalletTransactionsAdmin lists wallet ledger transactions for admin review.
// @Summary      List wallet transactions (admin)
// @Description  Paginated, filterable list of wallet ledger transactions for admin transaction management.
// @Tags         Admin Transactions
// @Produce      json
// @Security     BearerAuth
// @Param        page      query int    false "Page number" default(1)
// @Param        limit     query int    false "Items per page" default(20)
// @Param        search    query string false "Search description, Stripe session ID, or customer"
// @Param        type      query string false "Filter by type (deposit|payment|refund|adjustment|membership)"
// @Param        status    query string false "Filter by status"
// @Param        user_id   query int    false "Filter by user ID"
// @Param        date_from query string false "Start date (YYYY-MM-DD)"
// @Param        date_to   query string false "End date (YYYY-MM-DD)"
// @Success      200 {object} utils.Response{data=dto.AdminWalletTxListData}
// @Failure      401 {object} utils.Response
// @Failure      403 {object} utils.Response
// @Router       /admin/wallet/transactions [get]
func (ctrl *WalletHandler) ListWalletTransactionsAdmin(c *gin.Context) {
	var filters dto.AdminWalletTxListFilters
	if err := c.ShouldBindQuery(&filters); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "invalid query parameters")
		return
	}
	filters.Page, filters.Limit = pageLimitParams(c, constants.DefaultLimit)

	transactions, total, err := ctrl.walletService.ListAdminTransactions(c.Request.Context(), filters)
	if err != nil {
		utils.HandleServiceError(c, err, "failed to list wallet transactions")
		return
	}

	items := make([]dto.AdminWalletTxListItem, 0, len(transactions))
	for i := range transactions {
		items = append(items, dto.ToAdminWalletTxListItem(&transactions[i]))
	}
	utils.SuccessResponse(c, constants.MsgFetchSuccess, dto.AdminWalletTxListData{
		Transactions: items,
		Total:        total,
		Page:         filters.Page,
		Limit:        filters.Limit,
	})
}

// GetWalletTransactionAdmin returns a single wallet ledger transaction for admin review.
// @Summary      Get wallet transaction by ID (admin)
// @Description  Returns full wallet ledger transaction details for admin review.
// @Tags         Admin Transactions
// @Produce      json
// @Security     BearerAuth
// @Param        id path int true "Transaction ID"
// @Success      200 {object} utils.Response{data=dto.AdminWalletTxDetailResponse}
// @Failure      401 {object} utils.Response
// @Failure      403 {object} utils.Response
// @Failure      404 {object} utils.Response
// @Router       /admin/wallet/transactions/{id} [get]
func (ctrl *WalletHandler) GetWalletTransactionAdmin(c *gin.Context) {
	txID, ok := parseUintParam(c, "id")
	if !ok {
		return
	}
	record, err := ctrl.walletService.GetTransactionByIDAdmin(c.Request.Context(), txID)
	if err != nil {
		utils.HandleServiceError(c, err, "failed to load wallet transaction")
		return
	}
	utils.SuccessResponse(c, constants.MsgFetchSuccess, dto.ToAdminWalletTxDetail(record))
}

// GetWalletTransactionsSummaryAdmin returns KPI counters for the admin wallet ledger view.
// @Summary      Get wallet transactions summary (admin)
// @Description  Returns aggregate counters (by status and type) and net completed volume for the wallet ledger.
// @Tags         Admin Transactions
// @Produce      json
// @Security     BearerAuth
// @Success      200 {object} utils.Response{data=dto.WalletTxSummaryResponse}
// @Failure      401 {object} utils.Response
// @Failure      403 {object} utils.Response
// @Router       /admin/wallet/transactions/summary [get]
func (ctrl *WalletHandler) GetWalletTransactionsSummaryAdmin(c *gin.Context) {
	summary, err := ctrl.walletService.GetTransactionsSummaryAdmin(c.Request.Context())
	if err != nil {
		utils.HandleServiceError(c, err, "failed to load wallet summary")
		return
	}
	utils.SuccessResponse(c, constants.MsgFetchSuccess, summary)
}
