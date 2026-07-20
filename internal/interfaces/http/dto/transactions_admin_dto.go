package dto

// TransactionsHubSummaryResponse combines payment and wallet ledger KPIs for the
// admin transactions hub landing page.
type TransactionsHubSummaryResponse struct {
	PaymentsTotal   int64   `json:"payments_total"`
	WalletTxTotal   int64   `json:"wallet_tx_total"`
	PaymentsVolume  float64 `json:"payments_volume"`
	WalletVolume    float64 `json:"wallet_volume"`
	PendingPayments int64   `json:"pending_payments"`
	FailedPayments  int64   `json:"failed_payments"`
}
