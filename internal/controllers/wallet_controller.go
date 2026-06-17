package controllers

import (
	"github.com/alireza-akbarzadeh/luxe/internal/constants"
	"github.com/alireza-akbarzadeh/luxe/internal/dto"
	"github.com/alireza-akbarzadeh/luxe/internal/middleware"
	"github.com/alireza-akbarzadeh/luxe/internal/services"
	"github.com/alireza-akbarzadeh/luxe/internal/utils"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type WalletController struct {
	walletService services.WalletServiceInterface
	validate      *validator.Validate
}

func NewWallerController(walletService services.WalletServiceInterface) *WalletController {
	return &WalletController{
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
func (ctrl *WalletController) GetWallet(c *gin.Context) {
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
	balance, err := ctrl.walletService.GetBalance(userID)
	if err != nil {
		utils.HandleServiceError(c, err, "failed to get wallet balance")
		return
	}
	transactions, total, err := ctrl.walletService.GetTransactions(userID, filters)
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
func (ctrl *WalletController) Deposit(c *gin.Context) {
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

	result, err := ctrl.walletService.InitiateDeposit(userID, req.Amount, customerEmail)
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
func (ctrl *WalletController) AdminAdjust(c *gin.Context) {
	var req dto.AdminAdjustRequest
	if !utils.BindAndValidate(c, &req, ctrl.validate) {
		return
	}
	if err := ctrl.walletService.AdminAdjust(req.UserID, req.Amount, req.Description); err != nil {
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
func (ctrl *WalletController) Withdraw(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		utils.UnauthorizedResponse(c, constants.ErrorUnauthorized)
		return
	}
	var req dto.WithdrawRequest
	if !utils.BindAndValidate(c, &req, ctrl.validate) {
		return
	}
	err := ctrl.walletService.Withdraw(userID, req.Amount, "user_withdrawal", nil, req.Description)
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
func (ctrl *WalletController) GetTransaction(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		utils.UnauthorizedResponse(c, constants.ErrorUnauthorized)
		return
	}
	txID, ok := parseUintParam(c, "id")
	if !ok {
		return
	}
	tx, err := ctrl.walletService.GetTransaction(userID, txID)
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
func (ctrl *WalletController) CancelPendingDeposit(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		utils.UnauthorizedResponse(c, constants.ErrUnauthorized)
		return
	}

	txID, ok := parseUintParam(c, "id")
	if !ok {
		return
	}

	if err := ctrl.walletService.CancelPendingDeposit(userID, txID); err != nil {
		utils.HandleServiceError(c, err, "failed to cancel deposit")
		return
	}

	utils.SuccessResponse(c, "deposit cancelled", nil)
}
