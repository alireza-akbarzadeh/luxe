package order

// Status constants used by domain rules (mirror application constants).
const (
	StatusPending   = "pending"
	StatusPaid      = "paid"
	StatusShipped   = "shipped"
	StatusDelivered = "delivered"
	StatusCancelled = "cancelled"
	StatusRefunded  = "refunded"
	StatusDelayed   = "delayed"
)

// PaymentStatusSucceeded is the succeeded payment status string.
const PaymentStatusSucceeded = "succeeded"

// PaymentMethodWallet is the wallet payment method identifier.
const PaymentMethodWallet = "wallet"

// CanCancel reports whether a customer may cancel an order.
func CanCancel(o Order) error {
	switch o.Status {
	case StatusDelivered, "completed", StatusRefunded, StatusCancelled:
		return ErrCannotCancel
	default:
		return nil
	}
}

// IsValidBulkAdminStatus reports whether an admin bulk status update is allowed.
func IsValidBulkAdminStatus(status string) bool {
	switch status {
	case StatusPaid, StatusShipped, StatusDelivered, StatusCancelled:
		return true
	default:
		return false
	}
}

// WorkflowStateCode maps an order status to a workflow state code.
func WorkflowStateCode(status string) string {
	switch status {
	case StatusPending:
		return "pending_payment"
	case StatusPaid:
		return "paid"
	case "processing":
		return "processing"
	case StatusShipped:
		return "shipped"
	case StatusDelivered:
		return "delivered"
	case "completed":
		return "completed"
	case StatusCancelled:
		return "cancelled"
	case StatusRefunded:
		return "refunded"
	default:
		return ""
	}
}

// ShouldRefundWalletOnCancel reports whether wallet balance should be restored on cancel.
func ShouldRefundWalletOnCancel(paymentMethod, paymentStatus string) bool {
	return paymentMethod == PaymentMethodWallet && paymentStatus == PaymentStatusSucceeded
}
