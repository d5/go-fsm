package fsm_test

import (
	"math"

	"github.com/d5/go-fsm"
)

var script = []byte(`
export {
	truthy: func(src, dst, v) { return !!v },
	falsy: func(src, dst, v) { return !v },
	action: func(src, dst, v) { printf("%s -> %s: %v\n", src, dst, v) },
	enter: func(src, dst, v) { printf("%v ->: %v\n", dst, v) },
	leave: func(src, dst, v) { printf("-> %v: %v\n", src, v) }
}
`)

func Example() {
	machine, err := fsm.New(script).
		State("S", "enter", "leave").
		State("T", "enter", "leave").
		State("F", "enter", "leave").
		Transition("S", "T", "truthy", "action").
		Transition("S", "F", "falsy", "action").
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
	// -> S: 1
	// S -> T: 1
	// T ->: 1
	// -> S: NaN
	// S -> F: NaN
	// F ->: NaN
	// -> S: foobar
	// S -> T: foobar
	// T ->: foobar
	// -> S: []
	// S -> F: []
	// F ->: []
}
