package fsm

import (
	"github.com/d5/tengo/objects"
	"github.com/d5/tengo/script"
)

// StateMachine is a compiled state machine. Use Builder to
// construct and compile StateMachine.
type StateMachine struct {
	invokeScript *script.Compiled
	entryFns     map[string]string
	exitFns      map[string]string
	transitions  map[string][]transition
}

// Run executes the state machine from an initial state 'src' and an optional data value 'in'.
// See https://github.com/d5/tengo/blob/master/docs/interoperability.md#type-conversion-table for
// data value conversion details. Run continues to evaluate and move between states, until there
// are no more transitions available. When it stops, Run returns the final value 'out' or an error
// if there was one.
func (m *StateMachine) Run(src string, in interface{}) (out interface{}, err error) {
	value, err := objects.FromInterface(in)
	if err != nil {
		return nil, err
	}

	for {
		dst, err := m.eval(src, value)
		if err != nil {
			return nil, err
		}
		if dst == "" {
			// no more transition
			break
		}

		value, err = m.doTransition(src, dst, value)
		if err != nil {
			return nil, err
		}

		src = dst
	}

	return value, nil
}

func (m *StateMachine) eval(src string, in objects.Object) (string, error) {
	transitions, ok := m.transitions[src]
	if !ok {
		// no transition found
		return "", nil
	}

	for _, t := range transitions {
		if t.cond == "" {
			return t.dst, nil
		}

		out, err := m.invoke(src, t.dst, t.cond, in)
		if err != nil {
			return "", err
		}

		if out.Bool() {
			return t.dst, nil
		}
	}

	// no transition found
	return "", nil
}

func (m *StateMachine) doTransition(src, dst string, in objects.Object) (objects.Object, error) {
	if exitFn := m.exitFns[src]; exitFn != "" {
		out, err := m.invoke(src, dst, exitFn, in)
		if err != nil {
			return nil, err
		}

		if !out.IsUndefined() {
			in = out.Object()
		}
	}

	if entryFn := m.entryFns[dst]; entryFn != "" {
		out, err := m.invoke(src, dst, entryFn, in)
		if err != nil {
			return nil, err
		}

		if !out.IsUndefined() {
			in = out.Object()
		}
	}

	return in, nil
}

func (m *StateMachine) invoke(src, dst, fn string, in objects.Object) (out *script.Variable, err error) {
	_ = m.invokeScript.Set("src", &objects.String{Value: src})
	_ = m.invokeScript.Set("dst", &objects.String{Value: dst})
	_ = m.invokeScript.Set("fn", &objects.String{Value: fn})
	_ = m.invokeScript.Set("v", in)

	err = m.invokeScript.Run()
	if err != nil {
		return
	}

	out = m.invokeScript.Get("out")

	return
}
