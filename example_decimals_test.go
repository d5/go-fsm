package fsm_test

import (
	"fmt"

	"github.com/d5/go-fsm"
)

var decimalsScript = []byte(`
export {
	// test if the first character is a digit
	is_digit: func(src, dst, v) {
		return v[0] >= '0' && v[0] <= '9'
	},
	// test if the first character is a period
	is_dot: func(src, dst, v) {
		return v[0] == '.'  
	},
	// test if there are no more characters left
	is_eol: func(src, dst, v) {
		return len(v) == 0  
	},
	// prints out transition info and cut the first character
	enter: func(src, dst, v) {
		printf("%s -> %s: %v\n", src, dst, v)
		return v[1:]
	},
	enter_end: func(src, dst, v) {
		return "valid number"
	}, 
	enter_error: func(src, dst, v) {
		return "invalid number: " + v
	}
}`)

func Example_decimals() {
	// build and compile state machine
	machine, err := fsm.New(decimalsScript).
		State("S", "enter", "").       // start
		State("N", "enter", "").       // whole numbers
		State("P", "enter", "").       // decimal point
		State("F", "enter", "").       // fractional part
		State("E", "enter_end", "").   // end
		State("X", "enter_error", ""). // error
		Transition("S", "E", "is_eol").
		Transition("S", "N", "is_digit").
		Transition("S", "X", "").
		Transition("N", "E", "is_eol").
		Transition("N", "N", "is_digit").
		Transition("N", "P", "is_dot").
		Transition("N", "X", "").
		Transition("P", "F", "is_digit").
		Transition("P", "X", "").
		Transition("F", "E", "is_eol").
		Transition("F", "F", "is_digit").
		Transition("F", "X", "").
		Compile()
	if err != nil {
		panic(err)
	}

	// test case 1: "123.456"
	res, err := machine.Run("S", "123.456")
	if err != nil {
		panic(err)
	}
	fmt.Println(res)

	// test case 2: "12.34.65"
	res, err = machine.Run("S", "12.34.56")
	if err != nil {
		panic(err)
	}
	fmt.Println(res)

	// Output:
	// S -> N: 123.456
	// N -> N: 23.456
	// N -> N: 3.456
	// N -> P: .456
	// P -> F: 456
	// F -> F: 56
	// F -> F: 6
	// "valid number"
	// S -> N: 12.34.56
	// N -> N: 2.34.56
	// N -> P: .34.56
	// P -> F: 34.56
	// F -> F: 4.56
	// "invalid number: .56"
}
