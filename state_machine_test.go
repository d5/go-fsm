package fsm_test

import (
	"testing"

	"github.com/d5/go-fsm"
	"github.com/d5/tengo/v2/require"
)

func TestStateMachine_Run(t *testing.T) {
	machine, err := fsm.New(testScript).
		State("s1", "", "").
		State("s2", "", "").
		Transition("s1", "s2", "", "fn1"). // value not changed
		Compile()
	require.NoError(t, err)
	out, err := machine.Run("s1", 123)
	require.NoError(t, err)
	require.Equal(t, int64(123), out.Value())

	machine, err = fsm.New(testScript).
		State("s1", "", "").
		State("s2", "", "").
		Transition("s1", "s2", "", "fn2"). // change it to "foobar"
		Compile()
	require.NoError(t, err)
	out, err = machine.Run("s1", 123)
	require.NoError(t, err)
	require.Equal(t, "foobar", out.Value())

	machine, err = fsm.New(testScript).
		State("s1", "", "").
		State("s2", "", "").
		Transition("s1", "s2", "", "err1"). // error returned
		Compile()
	require.NoError(t, err)
	_, err = machine.Run("s1", 123)
	require.Error(t, err)
}
