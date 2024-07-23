package client

import (
	govclient "github.com/cosmos/cosmos-sdk/x/gov/client"

	"swisstronik/x/compliance/client/cli"
)

var (
	VerifyIssuerProposalHandler = govclient.NewProposalHandler(cli.CmdVerifyIssuerProposal)
)
