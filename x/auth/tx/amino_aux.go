package tx

import (
	"encoding/json"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	types "github.com/cosmos/cosmos-sdk/types/tx"
	signingtypes "github.com/cosmos/cosmos-sdk/types/tx/signing"

	"github.com/cosmos/cosmos-sdk/x/auth/migrations/legacytx"
	"github.com/cosmos/cosmos-sdk/x/auth/signing"
)

var _ signing.SignModeHandler = signModeAminoAuxHandler{}

// signModeAminoAuxHandler defines the SIGN_MODE_AMINO_AUX SignModeHandler
type signModeAminoAuxHandler struct{}

// DefaultMode implements SignModeHandler.DefaultMode
func (signModeAminoAuxHandler) DefaultMode() signingtypes.SignMode {
	return signingtypes.SignMode_SIGN_MODE_AMINO_AUX
}

// Modes implements SignModeHandler.Modes
func (signModeAminoAuxHandler) Modes() []signingtypes.SignMode {
	return []signingtypes.SignMode{signingtypes.SignMode_SIGN_MODE_AMINO_AUX}
}

// GetSignBytes implements SignModeHandler.GetSignBytes
func (signModeAminoAuxHandler) GetSignBytes(
	mode signingtypes.SignMode, data signing.SignerData, tx sdk.Tx,
) ([]byte, error) {

	if mode != signingtypes.SignMode_SIGN_MODE_AMINO_AUX {
		return nil, fmt.Errorf("expected %s, got %s", signingtypes.SignMode_SIGN_MODE_AMINO_AUX, mode)
	}

	protoTx, ok := tx.(*wrapper)
	if !ok {
		return nil, fmt.Errorf("can only handle a protobuf Tx, got %T", tx)
	}

	msgs := protoTx.GetMsgs()
	msgsBytes := make([]json.RawMessage, 0, len(msgs))
	for _, msg := range msgs {
		legacyMsg, ok := msg.(legacytx.LegacyMsg)
		if !ok {
			panic(fmt.Errorf("expected %T when using amino JSON", (*legacytx.LegacyMsg)(nil)))
		}

		msgsBytes = append(msgsBytes, json.RawMessage(legacyMsg.GetSignBytes()))
	}

	signDocDirectAux := types.StdSignDocAux{
		AccountNumber: data.AccountNumber,
		Sequence:      data.Sequence,
		TimeoutHeight: protoTx.GetTimeoutHeight(),
		ChainId:       data.ChainID,
		Memo:          protoTx.tx.Body.Memo,
		Tip:           protoTx.tx.AuthInfo.Tip.Amount,
		Msgs:          msgsBytes,
	}

	return signDocDirectAux.Marshal()
}