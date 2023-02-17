package nft

import (
	"net/http"

	"github.com/bnb-chain/zkbnb/service/apiserver/internal/logic/nft"
	"github.com/bnb-chain/zkbnb/service/apiserver/internal/svc"
	"github.com/bnb-chain/zkbnb/service/apiserver/internal/types"
	"github.com/zeromicro/go-zero/rest/httpx"
)

func GetNftByTxHashHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.ReqGetNftIndex
		if err := httpx.Parse(r, &req); err != nil {
			httpx.Error(w, err)
			return
		}

		l := nft.NewGetNftByTxHashLogic(r.Context(), svcCtx)
		resp, err := l.GetNftByTxHash(&req)
		if err != nil {
			httpx.Error(w, err)
		} else {
			httpx.OkJson(w, resp)
		}
	}
}
