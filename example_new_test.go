package fsm_test

import "github.com/d5/go-fsm"

func ExampleNew() {
	var script = []byte(`
		export {
			truthy: func(src, dst, v) { return !!v },
			falsy: func(src, dst, v) { return !v },
			enter: func(src, dst, v) { printf("ENTER %v: %v\n", dst, v) },
			leave: func(src, dst, v) { printf("LEAVE %v: %v\n", src, v) }
		}`)

	_ = fsm.New(script).
		State("S", "enter", "leave").
		State("T", "enter", "leave").
		State("F", "enter", "leave").
		Transition("S", "T", "truthy").
		Transition("S", "F", "falsy")
}
