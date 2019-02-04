# go-fsm

[![GoDoc](https://godoc.org/github.com/d5/go-fsm?status.svg)](https://godoc.org/github.com/d5/go-fsm)
[![Build Status](https://travis-ci.org/d5/go-fsm.svg?branch=master)](https://travis-ci.org/d5/go-fsm)

**A scriptable [FSM](https://en.wikipedia.org/wiki/Finite-state_machine) library for Go**

- [Tengo](https://github.com/d5/tengo) language: fast and secure
- Scriptable functions: transition conditions, transition actions, state entry and exit actions
- Immutable values

## Concepts

### State Machines

A state machine is defined by 

- A set of named **states**, and,
- An ordered list of **transitions** between states


### States

A state is defined by:

- **Name**: a unique string identifier in the state machine
- **Entry Action**: a function that's executed when entering the state
- **Exit Action**: a function that's executed when exiting the state

### Transitions

A transition is defined by:

- **Src**: the current state
- **Dst**: the next state
- **Condition**: a function that evaluates the condition 
- **Action**: a function that's executed when the condition is fulfilled

### Condition Functions

If condition function name is not specified (an empty space), the transition is considered as unconditional (always evalutes to true).

Condition functions in the script should take 3 arguments:

```golang
func(src, dst, v) {
    /* some logic */
    return some_value 
}
```

- `src`: the current state
- `dst`: the next state
- `v`: the data value (immutable)

The state machine use the returned value to determine the condition of the transition. E.g. condition is fulfilled if the value is [truthy](https://github.com/d5/tengo/blob/master/docs/runtime-types.md#objectisfalsy). In Tengo, the function that does not return anything is treated as if it returns `undefined` which is falsy. 

### Action Functions

Action functions in the script should take 3 arguments:

```golang
func(src, dst, v) {
    /* some logic */
    return some_value 
}
```

- `src`: the current state
- `dst`: the next state
- `v`: the data value (immutable)

The data value passed to action functions is immutable, but, the function may return a new value to change the data value for the future condition/action functions.

- If the function returns `undefined` _(or does not return anything)_, the data value remains unmodified.
- If the function returns `error` objects (e.g. `return error("some error")`), the state machine stops and returns an error from `StateMachine.Run` function call.
- If the function returns a value of any other type, the data value of the state machine is changed to the returned value.  

### Input and Output

When running the state machine, user can pass an input data value that will be used by condition and action functions. The state machine will return the final _output_ data value when there are no more transitions available.

### Execution Flow

1. When the state machine starts, it's given an initial state and the input data.
2. The state machine evaluates a list of transitions that are defined with the current state as its `src` state. The state machine evaluates the transitions in the same order they were added (defined). 
    1. If `condition` script is specified, the state machine runs the script to determines whether the condition is fulfilled or not.
    2. If `condition` script is not specified, 
3. If one of the transition's condition is fulfilled, the state machine runs the action scripts:
    1. It runs `exit action` of the current state if it's defined.
    2. It runs `action` of the transition if it's defined.
    3. It runs `entry action` of the next state if it's defined.
4. If no transitions were fulfilled, the state machine stops and returns the final value.
5. Repeat from the step 2.

## Example

Here's an example code for an FSM that tests if the input string is valid decimal numbers (e.g. `123.456`) or not: 

```golang
package main

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
	// prints out transition info
	print_tx: func(src, dst, v) {
		printf("%s -> %s: %q\n", src, dst, v)
	},
	// cut the first character
	enter: func(src, dst, v) {
		return v[1:]
	},
	enter_end: func(src, dst, v) {
		return "valid number"
	}, 
	enter_error: func(src, dst, v) {
		return "invalid number: " + v
	}
}`)

func main() {
    // build and compile state machine
    machine, err := fsm.New(decimalsScript).
        State("S", "enter", "").       // start
        State("N", "enter", "").       // whole numbers
        State("P", "enter", "").       // decimal point
        State("F", "enter", "").       // fractional part
        State("E", "enter_end", "").   // end
        State("X", "enter_error", ""). // error
        Transition("S", "E", "is_eol", "print_tx").
        Transition("S", "N", "is_digit", "print_tx").
        Transition("S", "X", "", "print_tx").
        Transition("N", "E", "is_eol", "print_tx").
        Transition("N", "N", "is_digit", "print_tx").
        Transition("N", "P", "is_dot", "print_tx").
        Transition("N", "X", "", "print_tx").
        Transition("P", "F", "is_digit", "print_tx").
        Transition("P", "X", "", "print_tx").
        Transition("F", "E", "is_eol", "print_tx").
        Transition("F", "F", "is_digit", "print_tx").
        Transition("F", "X", "", "print_tx").
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
}
```

