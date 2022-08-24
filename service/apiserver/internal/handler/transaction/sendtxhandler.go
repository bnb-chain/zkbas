package transaction

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"

	"github.com/bnb-chain/zkbas/service/apiserver/internal/logic/transaction"
	"github.com/bnb-chain/zkbas/service/apiserver/internal/svc"
	"github.com/bnb-chain/zkbas/service/apiserver/internal/types"
)

func SendTxHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.ReqSendTx
		if err := httpx.Parse(r, &req); err != nil {
			httpx.Error(w, err)
			return
		}

		l := transaction.NewSendTxLogic(r.Context(), svcCtx)
		resp, err := l.SendTx(&req)
		if err != nil {
			httpx.Error(w, err)
		} else {
			httpx.OkJson(w, resp)
		}
	}
}
