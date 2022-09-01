package test

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/bnb-chain/zkbas/service/apiserver/internal/types"
)

func (s *ApiServerSuite) TestGetBlock() {

	type args struct {
		by    string
		value string
	}
	tests := []struct {
		name     string
		args     args
		httpCode int
	}{
		{"found by height", args{"height", "1"}, 200},
		{"found by commitment", args{"commitment", "0000000000000000000000000000000000000000000000000000000000000000"}, 200},
		{"invalidby", args{"invalidby", ""}, 400},
	}

	for _, tt := range tests {
		s.T().Run(tt.name, func(t *testing.T) {
			httpCode, result := GetBlock(s, tt.args.by, tt.args.value)
			assert.Equal(t, tt.httpCode, httpCode)
			if httpCode == http.StatusOK {
				assert.NotNil(t, result.Height)
				assert.NotNil(t, result.Commitment)
				assert.NotNil(t, result.Status)
				assert.NotNil(t, result.StateRoot)
				fmt.Printf("result: %+v \n", result)
			}
		})
	}

}

func GetBlock(s *ApiServerSuite, by, value string) (int, *types.Block) {
	resp, err := http.Get(fmt.Sprintf("%s/api/v1/block?by=%s&value=%s", s.url, by, value))
	assert.NoError(s.T(), err)
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	assert.NoError(s.T(), err)

	if resp.StatusCode != http.StatusOK {
		return resp.StatusCode, nil
	}
	result := types.Block{}
	//nolint:errcheck
	json.Unmarshal(body, &result)
	return resp.StatusCode, &result
}
