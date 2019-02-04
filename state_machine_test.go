package fsm_test

import (
	"testing"

	"github.com/d5/go-fsm"
	"github.com/d5/tengo/assert"
)

func TestStateMachine_Run(t *testing.T) {
	machine, err := fsm.New(testScript).
		State("s1", "", "").
		State("s2", "", "").
		Transition("s1", "s2", "", "fn1"). // value not changed
		Compile()
	assert.NoError(t, err)
	out, err := machine.Run("s1", 123)
	assert.NoError(t, err)
	assert.Equal(t, int64(123), out.Value())

	machine, err = fsm.New(testScript).
		State("s1", "", "").
		State("s2", "", "").
		Transition("s1", "s2", "", "fn2"). // change it to "foobar"
		Compile()
	assert.NoError(t, err)
	out, err = machine.Run("s1", 123)
	assert.NoError(t, err)
	assert.Equal(t, "foobar", out.Value())

	machine, err = fsm.New(testScript).
		State("s1", "", "").
		State("s2", "", "").
		Transition("s1", "s2", "", "err1"). // error returned
		Compile()
	assert.NoError(t, err)
	_, err = machine.Run("s1", 123)
	assert.Error(t, err)
}
