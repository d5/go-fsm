package fsm

import (
	"errors"

	"github.com/d5/tengo/objects"
	"github.com/d5/tengo/script"
)

// StateMachine represents a compiled state machine. Use Builder to
// construct and compile StateMachine.
type StateMachine struct {
	invokeScript *script.Compiled
	entryFns     map[string]string
	exitFns      map[string]string
	transitions  map[string][]*transition
}

// Run executes the state machine from an initial state 'src' and an input data
// value 'in'. See
// https://github.com/d5/tengo/blob/master/docs/interoperability.md#type-conversion-table
// for data value conversions. Run continues to evaluate and move between
// states, until there are no more transitions available. When it stops, Run
// returns the final output value 'out' or an error 'err' if a script returned
// an error while executing.
func (m *StateMachine) Run(
	src string,
	in interface{},
) (out *script.Variable, err error) {
	value, err := script.NewVariable("", in)
	if err != nil {
		return nil, err
	}

	for {
		t, err := m.eval(src, value)
		if err != nil {
			return nil, err
		}
		if t == nil {
			// no more transition
			break
		}
		value, err = m.doTransition(src, t.dst, t.action, value)
		if err != nil {
			return nil, err
		}
		src = t.dst
	}
	return value, nil
}

func (m *StateMachine) eval(
	src string,
	in *script.Variable,
) (*transition, error) {
	transitions, ok := m.transitions[src]
	if !ok {
		// no transition found
		return nil, nil
	}

	for _, t := range transitions {
		if t.condition == "" {
			return t, nil
		}
		out, err := m.invoke(src, t.dst, t.condition, in)
		if err != nil {
			return nil, err
		}
		if out.Bool() {
			return t, nil
		}
	}
	return nil, nil // no transition found
}

func (m *StateMachine) doTransition(
	src, dst, action string,
	in *script.Variable,
) (*script.Variable, error) {
	if exitFn := m.exitFns[src]; exitFn != "" {
		out, err := m.invoke(src, dst, exitFn, in)
		if err != nil {
			return nil, err
		}
		if !out.IsUndefined() {
			in = out
		}
	}

	if action != "" {
		out, err := m.invoke(src, dst, action, in)
		if err != nil {
			return nil, err
		}
		if !out.IsUndefined() {
			in = out
		}
	}

	if entryFn := m.entryFns[dst]; entryFn != "" {
		out, err := m.invoke(src, dst, entryFn, in)
		if err != nil {
			return nil, err
		}
		if !out.IsUndefined() {
			in = out
		}
	}
	return in, nil
}

func (m *StateMachine) invoke(
	src, dst, fn string,
	in *script.Variable,
) (out *script.Variable, err error) {
	_ = m.invokeScript.Set("src", &objects.String{Value: src})
	_ = m.invokeScript.Set("dst", &objects.String{Value: dst})
	_ = m.invokeScript.Set("fn", &objects.String{Value: fn})
	_ = m.invokeScript.Set("v", in.Object())
	err = m.invokeScript.Run()
	if err != nil {
		return
	}

	out = m.invokeScript.Get("out")
	if out, isErr := out.Object().(*objects.Error); isErr {
		return nil, errors.New(out.String())
	}
	return
}
