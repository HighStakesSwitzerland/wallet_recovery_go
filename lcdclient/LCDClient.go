package lcdclient

import (
	"encoding/json"
	"fmt"
	"github.com/HighStakesSwitzerland/wallet_recovery_go/config"
	"github.com/cosmos/cosmos-sdk/x/staking/types"
	"io"
	"net/http"
	"time"
)

func GetPendingUndelegations() (*types.QueryDelegatorUnbondingDelegationsResponse, error) {
	var client = http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(config.LcdClientUrl + fmt.Sprintf("/cosmos/staking/v1beta1/delegators/%s/unbonding_delegations?pagination.limit=1000", config.DestinationWalletBech32))
	if err != nil {
		return nil, err
	}

	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	out, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != 200 {
		return nil, err
	}

	var response types.QueryDelegatorUnbondingDelegationsResponse
	err = json.Unmarshal(out, &response)

	n := 0
	for _, res := range response.UnbondingResponses {
		n += len(res.Entries)
	}

	config.Logger.Info(fmt.Sprintf("Got %d pending undelegations from LCD", n))
	return &response, err
}
