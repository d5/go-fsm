package fsm

import (
	"errors"
	"fmt"

	"github.com/d5/tengo/objects"
	"github.com/d5/tengo/script"
)

// Builder constructs and compiles the state machine.
// Call New to create a new Builder.
type Builder struct {
	userScript  []byte
	entryFns    map[string]string
	exitFns     map[string]string
	transitions map[string][]transition
}

// New creates a new Builder with a user script.
// User script must export all functions that are used for
// state entry/exit function and transition condition functions.
// Each exported function must take 3 arguments: source state,
// destination state, and, the value.
//
//  // user script
//  export {
//    name: func(src, dst, v) { /* ... */ }
//  }
//
func New(userScript []byte) *Builder {
	return &Builder{
		userScript:  userScript,
		entryFns:    make(map[string]string),
		exitFns:     make(map[string]string),
		transitions: make(map[string][]transition),
	}
}

// State defines a state with its entry/exit function names.
// Entry and exit functions are optional, but, if specified, the function
// in the user script must take 3 arguments:
//
//  // user script
//  export {
//    name: func(src, dst, v) { /* ... */ }
//  }
//
// 'src' is the source state name, and, 'dst' is the destination state name.
// For entry functions, 'src' is the previous state, and, 'dst' is entering state name.
// For exit function, 'src' is the leaving state, and, 'dst' is the next state.
// A state machine maintains a value that's passed when running it (via StateMachine.Run function),
// and, here the argument 'v' is the current value. 'v' is immutable (cannot be mutated), but, entry
// and exit functions can return a new value to change it. If they don't return anything (or return
// 'undefined'), the value will not be changed.
func (b *Builder) State(name, entryFunc, exitFunc string) *Builder {
	b.entryFns[name] = entryFunc
	b.exitFns[name] = exitFunc

	return b
}

// Transition adds a transition from src to dst state. It takes an optional condition
// function name. You can use an empty string for condition function name if the transition is
// unconditional (which means the transition always evaluates to true). Condition function must \
// take 3 arguments:
//
//  // user script
//  export {
//    name: func(src, dst, v) { /* ... */ }
//  }
//
// 'src' is the current state, and, 'dst' is the next state that's being evaluated. 'v' is the current
// value the state machine is maintaining, and, it's immutable. Condition function should return a value,
// and, the truthiness of the return value determines whether transition should happen or not. (See
// https://github.com/d5/tengo/blob/master/docs/runtime-types.md#objectisfalsy.)
func (b *Builder) Transition(src, dst, condFunc string) *Builder {
	b.transitions[src] = append(b.transitions[src], transition{
		dst:  dst,
		cond: condFunc,
	})

	return b
}

// Compile compiles the user script and builds the state machine.
// Compile does not validate the states and transitions.
// Call Validate or ValidateCompile to validate them.
func (b *Builder) Compile() (*StateMachine, error) {
	return b.compile()
}

// Validate validates all states and transitions. It ensures that
// all states are properly defined and all functions are exported
// from the user script.
func (b *Builder) Validate() error {
	return b.validate()
}

// ValidateCompile validates all states and transitions, and, also
// builds the state machine. Call Compile if you don't need to validate
// states and transitions.
func (b *Builder) ValidateCompile() (*StateMachine, error) {
	if err := b.validate(); err != nil {
		return nil, err
	}

	return b.compile()
}

func (b *Builder) validate() error {
	// compile validation script
	s := script.New(retrieveScript)
	_ = s.Add("fn", "")
	s.SetUserModuleLoader(func(_ string) ([]byte, error) { return b.userScript, nil })
	c, err := s.Compile()
	if err != nil {
		return fmt.Errorf("failed to compile script: %s", err.Error())
	}

	// validate states
	for state, entryFunc := range b.entryFns {
		if state == "" {
			return errors.New("state name must not be empty")
		}

		if entryFunc != "" {
			if err := validateFunc(c, entryFunc); err != nil {
				return err
			}
		}

		if exitFunc := b.exitFns[state]; exitFunc != "" {
			if err := validateFunc(c, exitFunc); err != nil {
				return err
			}
		}
	}

	// validate transitions
	for src, transitions := range b.transitions {
		if _, ok := b.entryFns[src]; !ok {
			return fmt.Errorf("state '%s' not found", src)
		}
		for _, t := range transitions {
			if _, ok := b.entryFns[t.dst]; !ok {
				return fmt.Errorf("state '%s' not found", src)
			}

			// validate condition function
			if t.cond != "" {
				if err := validateFunc(c, t.cond); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func (b *Builder) compile() (*StateMachine, error) {
	// compile state machine invoke script
	s := script.New(invokeScript)
	_ = s.Add("src", "")
	_ = s.Add("dst", "")
	_ = s.Add("fn", "")
	_ = s.Add("v", nil)
	s.SetUserModuleLoader(func(_ string) ([]byte, error) { return b.userScript, nil })
	compiled, err := s.Compile()
	if err != nil {
		return nil, fmt.Errorf("failed to compile script: %s", err.Error())
	}

	transitions := make(map[string][]transition)
	for src, tx := range b.transitions {
		transitions[src] = append([]transition{}, tx...)
	}

	return &StateMachine{
		invokeScript: compiled,
		entryFns:     copyFuncMap(b.entryFns),
		exitFns:      copyFuncMap(b.exitFns),
		transitions:  transitions,
	}, nil
}

func validateFunc(c *script.Compiled, name string) error {
	_ = c.Set("fn", &objects.String{Value: name})
	err := c.Run()
	if err != nil {
		return fmt.Errorf("script execution error: %s", err.Error())
	}

	out := c.Get("out")
	if out.IsUndefined() {

	}
	switch out := out.Object().(type) {
	case *objects.Undefined:
		return fmt.Errorf("function '%s' not found", name)
	case *objects.CompiledFunction:
		if out.NumParameters != 3 {
			return fmt.Errorf("function '%s' wrong number of arguments: want 3 got %d", name, out.NumParameters)
		}
	case *objects.Closure:
		if out.Fn.NumParameters != 3 {
			return fmt.Errorf("function '%s' wrong number of arguments: want 3 got %d", name, out.Fn.NumParameters)
		}
	case objects.Callable:
	default:
		return fmt.Errorf("'%s' is not callable", name)
	}

	return nil
}

func copyFuncMap(s map[string]string) map[string]string {
	t := make(map[string]string, len(s))
	for k, v := range s {
		if v != "" { // skip empty function names
			t[k] = v
		}
	}

	return t
}
