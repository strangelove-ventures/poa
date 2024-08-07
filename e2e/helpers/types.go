package helpers

import (
	"time"

	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

type Validators []struct {
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
	UnbondingTime string `json:"unbonding_time"`
	Commission    struct {
		CommissionRates struct {
			Rate          string `json:"rate"`
			MaxRate       string `json:"max_rate"`
			MaxChangeRate string `json:"max_change_rate"`
		} `json:"commission_rates"`
		UpdateTime string `json:"update_time"`
	} `json:"commission"`
	MinSelfDelegation string `json:"min_self_delegation"`
}

// From stakingtypes.Validator
type Vals struct {
	Validators Validators `json:"validators"`
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
		ChainID     string `json:"chain_id"`
		Height      string `json:"height"`
		Time        string `json:"time"`
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
			BlockIDFlag      string `json:"block_id_flag"`
			ValidatorAddress string `json:"validator_address"`
			Timestamp        string `json:"timestamp"`
			Signature        string `json:"signature"`
		} `json:"signatures"`
	} `json:"last_commit"`
}

type POAConsensusPower struct {
	Power string `json:"consensus_power"`
}

type StakingParams struct {
	Params stakingtypes.Params `json:"params"`
}

type SingingInformation struct {
	Info []struct {
		Address     string `json:"address"`
		IndexOffset string `json:"index_offset,omitempty"`
		JailedUntil string `json:"jailed_until"`
	} `json:"info"`
	Pagination struct {
		Total string `json:"total"`
	} `json:"pagination"`
}

type POAPending struct {
	Pending []struct {
		OperatorAddress string `json:"operator_address"`
		ConsensusPubkey struct {
			Type  string `json:"type"`
			Value string `json:"value"`
		} `json:"consensus_pubkey"`
		Status          int    `json:"status"`
		Tokens          string `json:"tokens"`
		DelegatorShares string `json:"delegator_shares"`
		Description     struct {
			Moniker         string `json:"moniker"`
			Website         string `json:"website"`
			SecurityContact string `json:"security_contact"`
			Details         string `json:"details"`
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
	} `json:"pending"`
}

type CometBFTConsensus struct {
	BlockHeight string `json:"block_height"`
	Validators  []struct {
		Address string `json:"address"`
		PubKey  struct {
			Type  string `json:"type"`
			Value string `json:"value"`
		} `json:"pub_key"`
		VotingPower string `json:"voting_power"`
	} `json:"validators"`
	Pagination struct {
		Total string `json:"total"`
	} `json:"pagination"`
}
