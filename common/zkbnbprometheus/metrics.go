package zkbnbprometheus

import "github.com/prometheus/client_golang/prometheus"

type Metrics struct {
	TxPrepareMetrics           prometheus.Gauge
	TxVerifyInputsMetrics      prometheus.Gauge
	TxGenerateTxDetailsMetrics prometheus.Gauge
	TxApplyTransactionMetrics  prometheus.Gauge
	TxGeneratePubDataMetrics   prometheus.Gauge
	TxGetExecutedTxMetrics     prometheus.Gauge
}
