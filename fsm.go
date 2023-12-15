package easyFSM

import (
	"fmt"

	"golang.org/x/exp/constraints"
)

func NewFSM[E, S constraints.Ordered](startState S) FSM[E, S] {
	stateAll := make(map[S]bool)
	stateAll[startState] = true

	return FSM[E, S]{
		startState:     startState,
		current:        startState,
		stateAll:       stateAll,
		transitions:    make(map[matchKey[E, S]]S),
		defineSequence: make([]matchKey[E, S], 0, 5),
	}
}

type FSM[E, S constraints.Ordered] struct {
	startState     S
	current        S
	stateAll       map[S]bool
	transitions    map[matchKey[E, S]]S
	defineSequence []matchKey[E, S]
}

// DefineTransition adds a new transition to the Finite State Machine (FSM).
// It defines that when a specific event occurs in a particular source state,
// the FSM will transition to the destination state.
// If the same event and source state combination already exists in the FSM,
// it will panic with an error indicating that a duplicated transition is being added.
//
// src -->|event| dest
//
// Parameters:
// - event: The event triggering the transition.
// - src: The source state from which the transition is allowed.
// - dest: The destination state to which the FSM will transition when the event occurs in the source state.
//
// Note:
// 1. The FSM should be placed in the global scope.
// 2. When importing the package, transitions should be added using the DefineTransition function during the initialization step.
func (fsm FSM[E, S]) DefineTransition(event E, src, dest S) FSM[E, S] {
	key := matchKey[E, S]{
		event: event,
		src:   src,
	}
	_, exist := fsm.transitions[key]
	if exist {
		panic(fmt.Sprintf("duplicated transition: event=%v source state=%v", event, src))
	}

	fsm.transitions[key] = dest
	fsm.defineSequence = append(fsm.defineSequence, key)

	fsm.stateAll[src] = true
	fsm.stateAll[dest] = true
	return fsm
}

// OnAction triggers the transition in the Finite State Machine (FSM) when a specific event occurs.
// It first calls the internal doTransition function to determine the destination state after the event.
// If the transition is successful, it calls the provided action function with the destination state as a parameter.
// The action function is responsible for performing any necessary actions or operations associated with the state transition.
//
// Parameters:
// - event: The event that triggers the transition.
// - action: A function that takes the destination state as a parameter and returns an error if any.
func (fsm FSM[E, S]) OnAction(event E, action func(nextState S) error) error {
	dest, err := fsm.doTransition(event)
	if err != nil {
		return err
	}
	return action(dest)
}

func (fsm FSM[E, S]) doTransition(event E) (dest S, err error) {
	key := matchKey[E, S]{
		event: event,
		src:   fsm.current,
	}
	dest, ok := fsm.transitions[key]
	if ok {
		fsm.current = dest
		return dest, nil
	}

	for k := range fsm.transitions {
		if k.event == event {
			return dest, fmt.Errorf("for the trigger event '%v', the required source state is '%v' but the current state is '%v': %v",
				k.event,
				k.src,
				fsm.current,
				ErrStateNotMatch,
			)
		}
	}
	return dest, fmt.Errorf("event=%v: %w", event, ErrEventNotDefined)
}

func (fsm FSM[E, S]) CurrentState() S {
	return fsm.current
}

func (fsm FSM[E, S]) CopyFSM(currentState S) FSM[E, S] {
	fresh := fsm
	fresh.current = currentState
	return fresh
}

func (fsm FSM[E, S]) ShowStates() []S {
	var list []S
	list = append(list, fsm.startState)
	for state := range fsm.stateAll {
		if state == fsm.startState {
			continue
		}
		list = append(list, state)
	}
	return list
}

type matchKey[E, S constraints.Ordered] struct {
	event E
	src   S
}
