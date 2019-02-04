package fsm_test

import (
	"testing"

	"github.com/d5/go-fsm"
	"github.com/d5/tengo/assert"
)

var testScript = []byte(`
export {
	fn1: func(src, dst, v) {},
	fn2: func(src, dst, v) { return "foobar" },
	fn3: func(src, dst) {},
	err1: func(src, dst, v) { return error("an error occurred") },
	foo: [1, 2, 3]
}`)

func TestBuilder_Validate(t *testing.T) {
	// empty state name
	err := fsm.New(testScript).State("", "", "").Validate()
	assert.Equal(t, "state name must not be empty", err.Error())

	// entry function not found
	err = fsm.New(testScript).State("s1", "fn4", "").Validate()
	assert.Equal(t, "function 'fn4' not found", err.Error())

	// exit function not found
	err = fsm.New(testScript).State("s1", "", "fn4").Validate()
	assert.Equal(t, "function 'fn4' not found", err.Error())

	// transition src not found
	err = fsm.New(testScript).State("s1", "", "").Transition("s0", "s1", "", "").Validate()
	assert.Equal(t, "state 's0' not found", err.Error())

	// transition dst not found
	err = fsm.New(testScript).State("s1", "", "").Transition("s1", "s2", "", "").Validate()
	assert.Equal(t, "state 's1' not found", err.Error())

	// transition condition function not found
	err = fsm.New(testScript).State("s1", "", "").Transition("s1", "s1", "fn4", "").Validate()
	assert.Equal(t, "function 'fn4' not found", err.Error())

	// transition action function not found
	err = fsm.New(testScript).State("s1", "", "").Transition("s1", "s1", "", "fn4").Validate()
	assert.Equal(t, "function 'fn4' not found", err.Error())

	// not a function
	err = fsm.New(testScript).State("s1", "foo", "").Validate()
	assert.Equal(t, "'foo' is not callable", err.Error())

	// wrong number of arguments
	err = fsm.New(testScript).State("s1", "fn3", "").Validate()
	assert.Equal(t, "function 'fn3' wrong number of arguments: want 3 got 2", err.Error())

	// no error
	err = fsm.New(testScript).State("s1", "fn1", "fn2").Transition("s1", "s1", "fn1", "fn2").Validate()
	assert.NoError(t, err)
}
