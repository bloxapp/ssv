package controller

import (
	"testing"

	"github.com/bloxapp/ssv/protocol/v2/qbft/instance"

	specqbft "github.com/bloxapp/ssv-spec/qbft"
	"github.com/stretchr/testify/require"
)

func TestInstances_FindInstance(t *testing.T) {
	i := NewInstanceContainer(5,
		&instance.Instance{State: &specqbft.State{Height: 1}},
		&instance.Instance{State: &specqbft.State{Height: 2}},
		&instance.Instance{State: &specqbft.State{Height: 3}},
	)

	t.Run("find 1", func(t *testing.T) {
		require.NotNil(t, i.FindInstance(1))
	})
	t.Run("find 2", func(t *testing.T) {
		require.NotNil(t, i.FindInstance(2))
	})
	t.Run("find 5", func(t *testing.T) {
		require.Nil(t, i.FindInstance(5))
	})
}

func TestInstances_AddNewInstance(t *testing.T) {
	t.Run("add to full", func(t *testing.T) {
		i := NewInstanceContainer(5,
			&instance.Instance{State: &specqbft.State{Height: 1}},
			&instance.Instance{State: &specqbft.State{Height: 2}},
			&instance.Instance{State: &specqbft.State{Height: 3}},
			&instance.Instance{State: &specqbft.State{Height: 4}},
			&instance.Instance{State: &specqbft.State{Height: 5}},
		)
		i.AddNewInstance(&instance.Instance{State: &specqbft.State{Height: 6}})

		require.EqualValues(t, 6, i.Instances()[0].State.Height)
		require.EqualValues(t, 1, i.Instances()[1].State.Height)
		require.EqualValues(t, 2, i.Instances()[2].State.Height)
		require.EqualValues(t, 3, i.Instances()[3].State.Height)
		require.EqualValues(t, 4, i.Instances()[4].State.Height)
	})

	t.Run("add to empty", func(t *testing.T) {
		i := NewInstanceContainer(5)
		i.AddNewInstance(&instance.Instance{State: &specqbft.State{Height: 1}})

		require.EqualValues(t, 1, i.Instances()[0].State.Height)
		require.Len(t, i.Instances(), 1)
	})

	t.Run("add to semi full", func(t *testing.T) {
		i := NewInstanceContainer(5,
			&instance.Instance{State: &specqbft.State{Height: 1}},
			&instance.Instance{State: &specqbft.State{Height: 2}},
			&instance.Instance{State: &specqbft.State{Height: 3}},
		)
		i.AddNewInstance(&instance.Instance{State: &specqbft.State{Height: 4}})

		require.EqualValues(t, 4, i.Instances()[0].State.Height)
		require.EqualValues(t, 1, i.Instances()[1].State.Height)
		require.EqualValues(t, 2, i.Instances()[2].State.Height)
		require.EqualValues(t, 3, i.Instances()[3].State.Height)
		require.Len(t, i.Instances(), 4)
	})
}
