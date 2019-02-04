package fsm_test

import (
	"math"

	"github.com/d5/go-fsm"
)

var script = []byte(`
export {
	truthy: func(src, dst, v) { return !!v },
	falsy: func(src, dst, v) { return !v },
	enter: func(src, dst, v) { printf("ENTER %v: %v\n", dst, v) },
	leave: func(src, dst, v) { printf("LEAVE %v: %v\n", src, v) }
}
`)

func Example() {
	machine, err := fsm.New(script).
		State("S", "enter", "leave").
		State("T", "enter", "leave").
		State("F", "enter", "leave").
		Transition("S", "T", "truthy").
		Transition("S", "F", "falsy").
		Compile()
	if err != nil {
		panic(err)
	}

	if _, err := machine.Run("S", 1); err != nil {
		panic(err)
	}
	if _, err := machine.Run("S", math.NaN()); err != nil {
		panic(err)
	}
	if _, err := machine.Run("S", "foobar"); err != nil {
		panic(err)
	}
	if _, err := machine.Run("S", []interface{}{}); err != nil {
		panic(err)
	}

	// Output:
	// LEAVE S: 1
	// ENTER T: 1
	// LEAVE S: NaN
	// ENTER F: NaN
	// LEAVE S: foobar
	// ENTER T: foobar
	// LEAVE S: []
	// ENTER F: []
}
