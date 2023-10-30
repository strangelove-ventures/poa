package helpers

import "time"

// From stakingtypes.Validator
type Vals struct {
	Validators []struct {
		OperatorAddress string `json:"operator_address"`
		ConsensusPubkey struct {
			Type  string `json:"type"`
			Value string `json:"value"`
		} `json:"consensus_pubkey"`
		Status          int    `json:"status"`
		Tokens          string `json:"tokens"`
		DelegatorShares string `json:"delegator_shares"`
		Description     struct {
			Moniker string `json:"moniker"`
		} `json:"description"`
		UnbondingTime time.Time `json:"unbonding_time"`
		Commission    struct {
			CommissionRates struct {
				Rate          string `json:"rate"`
				MaxRate       string `json:"max_rate"`
				MaxChangeRate string `json:"max_change_rate"`
			} `json:"commission_rates"`
			UpdateTime time.Time `json:"update_time"`
		} `json:"commission"`
		MinSelfDelegation string `json:"min_self_delegation"`
	} `json:"validators"`
	Pagination struct {
		Total string `json:"total"`
	} `json:"pagination"`
}

// sdk v50 RPC/blocks endpoint
type BlockData struct {
	Header struct {
		Version struct {
			Block string `json:"block"`
			App   string `json:"app"`
		} `json:"version"`
		ChainID     string    `json:"chain_id"`
		Height      string    `json:"height"`
		Time        time.Time `json:"time"`
		LastBlockID struct {
			Hash          string `json:"hash"`
			PartSetHeader struct {
				Total int    `json:"total"`
				Hash  string `json:"hash"`
			} `json:"part_set_header"`
		} `json:"last_block_id"`
		LastCommitHash     string `json:"last_commit_hash"`
		DataHash           string `json:"data_hash"`
		ValidatorsHash     string `json:"validators_hash"`
		NextValidatorsHash string `json:"next_validators_hash"`
		ConsensusHash      string `json:"consensus_hash"`
		AppHash            string `json:"app_hash"`
		LastResultsHash    string `json:"last_results_hash"`
		EvidenceHash       string `json:"evidence_hash"`
		ProposerAddress    string `json:"proposer_address"`
	} `json:"header"`
	Data struct {
		Txs []string `json:"txs"`
	} `json:"data"`
	Evidence struct {
		Evidence []any `json:"evidence"`
	} `json:"evidence"`
	LastCommit struct {
		Height  string `json:"height"`
		Round   int    `json:"round"`
		BlockID struct {
			Hash          string `json:"hash"`
			PartSetHeader struct {
				Total int    `json:"total"`
				Hash  string `json:"hash"`
			} `json:"part_set_header"`
		} `json:"block_id"`
		Signatures []struct {
			BlockIDFlag      string    `json:"block_id_flag"`
			ValidatorAddress string    `json:"validator_address"`
			Timestamp        time.Time `json:"timestamp"`
			Signature        string    `json:"signature"`
		} `json:"signatures"`
	} `json:"last_commit"`
}

// sdk v50
type TxResponse struct {
	Height    string `json:"height"`
	Txhash    string `json:"txhash"`
	Codespace string `json:"codespace"`
	Code      int    `json:"code"`
	Data      string `json:"data"`
	RawLog    string `json:"raw_log"`
	Logs      []any  `json:"logs"`
	Info      string `json:"info"`
	GasWanted string `json:"gas_wanted"`
	GasUsed   string `json:"gas_used"`
	Tx        struct {
		Type string `json:"@type"`
		Body struct {
			Messages []struct {
				Type             string `json:"@type"`
				FromAddress      string `json:"from_address"`
				ValidatorAddress string `json:"validator_address"`
				Power            string `json:"power"`
			} `json:"messages"`
			Memo                        string `json:"memo"`
			TimeoutHeight               string `json:"timeout_height"`
			ExtensionOptions            []any  `json:"extension_options"`
			NonCriticalExtensionOptions []any  `json:"non_critical_extension_options"`
		} `json:"body"`
		AuthInfo struct {
			SignerInfos []struct {
				PublicKey struct {
					Type string `json:"@type"`
					Key  string `json:"key"`
				} `json:"public_key"`
				ModeInfo struct {
					Single struct {
						Mode string `json:"mode"`
					} `json:"single"`
				} `json:"mode_info"`
				Sequence string `json:"sequence"`
			} `json:"signer_infos"`
			Fee struct {
				Amount   []any  `json:"amount"`
				GasLimit string `json:"gas_limit"`
				Payer    string `json:"payer"`
				Granter  string `json:"granter"`
			} `json:"fee"`
			Tip any `json:"tip"`
		} `json:"auth_info"`
		Signatures []string `json:"signatures"`
	} `json:"tx"`
	Timestamp string `json:"timestamp"`
	Events    []struct {
		Type       string `json:"type"`
		Attributes []struct {
			Key   string `json:"key"`
			Value string `json:"value"`
			Index bool   `json:"index"`
		} `json:"attributes"`
	} `json:"events"`
}

type POAParams struct {
	Params struct {
		Admins []string `json:"admins"`
	} `json:"params"`
}
