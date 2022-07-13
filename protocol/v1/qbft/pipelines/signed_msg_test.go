package pipelines

import (
	"fmt"
	"testing"

	specqbft "github.com/bloxapp/ssv-spec/qbft"
	"github.com/stretchr/testify/require"
)

func TestIfFirstTrueContinueToSecond(t *testing.T) {
	validPipline := WrapFunc("valid", func(signedMessage *specqbft.SignedMessage) error {
		return nil
	})

	invalidPipeline := WrapFunc("invalid", func(signedMessage *specqbft.SignedMessage) error {
		return fmt.Errorf("error")
	})

	t.Run("a valid, b valid", func(t *testing.T) {
		require.NoError(t, IfFirstTrueContinueToSecond(validPipline, validPipline).Run(nil))
	})

	t.Run("a valid, b invalid", func(t *testing.T) {
		require.EqualError(t, IfFirstTrueContinueToSecond(validPipline, invalidPipeline).Run(nil), "error")
	})

	t.Run("a invalid, b valid", func(t *testing.T) {
		require.NoError(t, IfFirstTrueContinueToSecond(invalidPipeline, validPipline).Run(nil))
	})

	t.Run("a invalid, b invalid", func(t *testing.T) {
		require.NoError(t, IfFirstTrueContinueToSecond(invalidPipeline, invalidPipeline).Run(nil))
	})
}

// IfFirstTrueContinueToSecond runs pipeline a, if returns no error continues to pipeline b. otherwise returns without an error
func IfFirstTrueContinueToSecond(a, b SignedMessagePipeline) SignedMessagePipeline {
	return WrapFunc("if first pipeline non error, continue to second", func(signedMessage *specqbft.SignedMessage) error {
		if a.Run(signedMessage) == nil {
			return b.Run(signedMessage)
		}
		return nil
	})
}
