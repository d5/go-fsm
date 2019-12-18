package fsm

import (
	"errors"
	"fmt"

	"github.com/d5/tengo/objects"
	"github.com/d5/tengo/script"
	"github.com/d5/tengo/stdlib"
)

// Builder represents a state machine builder that constructs and compiles
// the state machine. Call New to create a new Builder.
type Builder struct {
	userScript  []byte
	entryFns    map[string]string
	exitFns     map[string]string
	transitions map[string][]*transition
}

// New creates a new Builder with a user script.
//
// User script must export functions for all condition and actions of the state
// machine.
func New(userScript []byte) *Builder {
	return &Builder{
		userScript:  userScript,
		entryFns:    make(map[string]string),
		exitFns:     make(map[string]string),
		transitions: make(map[string][]*transition),
	}
}

// State defines a state with its entry/exit action function names.
//
// Entry and exit action functions are optional, but, if specified, the function
// in the user script must take 3 arguments:
//
//  export {
//    action_name: func(src, dst, v) {
//      return some_value // optional
//    }
//  }
//
// For entry functions, 'src' is the previous state, and, 'dst' is entering
// state. For exit functions, 'src' is the leaving state, and, 'dst' is the
// next state. 'v' is the current data value of the state machine. 'v' itself
// is immutable, but, entry and exit action functions may return a new value
// to change it. If they don't return anything (or return 'undefined'), the
// value will not be changed. If it returns a Tengo error object, the state
// machine will stop and returns the error.
//
//  export {
//    action_name: func(src, dst, v) {
//      return error("an error occurred")
//    }
//  }
//
func (b *Builder) State(name, entryFunc, exitFunc string) *Builder {
	b.entryFns[name] = entryFunc
	b.exitFns[name] = exitFunc
	return b
}

// Transition defines (adds) a transition from 'src' to 'dst' states. It also
// takes the condition and action function names, which are optional. An empty
// condition function name makes the transition unconditional (which means the
// transition always evaluates to true). Condition function and action function
// must take 3 arguments:
//
//  export {
//    action_name: func(src, dst, v) {
//      return some_value // truthy or falsy
//    }
//  }
//
// 'src' is the current state, and, 'dst' is next state of the transition. 'v'
// is the current data value of the state machine, and, 'v' is immutable. For
// condition functions, the truthiness
// (https://github.com/d5/tengo/blob/master/docs/runtime-types.md#objectisfalsy)
// of the returned value determines whether the condition is fulfilled or not.
// For action functions, they may return a new value to change it. If they
// don't return anything (or return 'undefined'), the value will not be changed.
// If it returns a Tengo error object, the state machine will stop and returns
// the error.
//
//  export {
//    action_name: func(src, dst, v) {
//      return error("an error occurred")
//    }
//  }
//
func (b *Builder) Transition(src, dst, condition, action string) *Builder {
	b.transitions[src] = append(b.transitions[src], &transition{
		dst:       dst,
		condition: condition,
		action:    action,
	})
	return b
}

// Compile compiles the script and builds the state machine. This function does
// not validate the states and transitions. Call Validate or ValidateCompile if
// you want to validate them.
func (b *Builder) Compile() (*StateMachine, error) {
	return b.compile()
}

// Validate validates all states and transitions. It ensures that all states
// are properly defined and all condition and action functions are exported
// from the user script.
func (b *Builder) Validate() error {
	return b.validate()
}

// ValidateCompile is combination of Validate and Compile functions. Call
// Compile if you don't need to validate the states and transitions.
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
	importModules := stdlib.GetModuleMap(stdlib.AllModuleNames()...)
	importModules.Remove("os")
	importModules.Add("user", &objects.SourceModule{Src: b.userScript})
	s.SetImports(importModules)
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
			if t.condition != "" {
				if err := validateFunc(c, t.condition); err != nil {
					return err
				}
			}
			// validate action function
			if t.action != "" {
				if err := validateFunc(c, t.action); err != nil {
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
	importModules := stdlib.GetModuleMap(stdlib.AllModuleNames()...)
	importModules.Remove("os")
	importModules.Add("user", &objects.SourceModule{Src: b.userScript})
	s.SetImports(importModules)
	compiled, err := s.Compile()
	if err != nil {
		return nil, fmt.Errorf("failed to compile script: %s", err.Error())
	}
	transitions := make(map[string][]*transition)
	for src, tx := range b.transitions {
		transitions[src] = append([]*transition{}, tx...)
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
			return fmt.Errorf(
				"function '%s' wrong number of arguments: want 3 got %d",
				name, out.NumParameters)
		}
	case *objects.Closure:
		if out.Fn.NumParameters != 3 {
			return fmt.Errorf(
				"function '%s' wrong number of arguments: want 3 got %d",
				name, out.Fn.NumParameters)
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
