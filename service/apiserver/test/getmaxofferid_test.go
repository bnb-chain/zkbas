package test

import (
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/bnb-chain/zkbas/service/apiserver/internal/types"
)

func (s *ApiServerSuite) TestGetMaxOfferId() {
	type testcase struct {
		name     string
		args     int //accountIndex
		httpCode int
	}

	tests := []testcase{
		{"not found", math.MaxInt, 400},
	}

	statusCode, accounts := GetAccounts(s, 0, 100)
	if statusCode == http.StatusOK && len(accounts.Accounts) > 0 {
		tests = append(tests, []testcase{
			{"found", int(accounts.Accounts[0].Index), 200},
		}...)
	}

	for _, tt := range tests {
		s.T().Run(tt.name, func(t *testing.T) {
			httpCode, _ := GetMaxOfferId(s, tt.args)
			assert.Equal(t, tt.httpCode, httpCode)
		})
	}

}

func GetMaxOfferId(s *ApiServerSuite, accountIndex int) (int, *types.MaxOfferId) {
	resp, err := http.Get(fmt.Sprintf("%s/api/v1/maxOfferId?account_index=%d", s.url, accountIndex))
	assert.NoError(s.T(), err)
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	assert.NoError(s.T(), err)

	if resp.StatusCode != http.StatusOK {
		return resp.StatusCode, nil
	}
	result := types.MaxOfferId{}
	//nolint: errcheck
	json.Unmarshal(body, &result)
	return resp.StatusCode, &result
}
