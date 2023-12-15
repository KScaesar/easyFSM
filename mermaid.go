package easyFSM

import (
	"fmt"
	"strings"

	"golang.org/x/exp/constraints"
)

// MermaidGraphByTopDown generates a Mermaid graph representation of the FSM.
// It takes an FSM object and a TextTransformFunc as input.
// The TextTransformFunc is an optional function that can be used to transform the
// text representation of states and events before including them in the graph.
//
// The function returns a string containing the Mermaid graph definition.
//
// The Mermaid graph is generated in the Top-Down (TD) style, where each transition from a source state
// to a destination state is represented with an arrow labeled by the corresponding event.
//
// Example Usage:
//
//	fsm := NewFSM(StateA)
//	fsm.AddTransition(StateA, EventX, StateB)
//	fsm.AddTransition(StateB, EventY, StateC)
//	graph := MermaidGraphByTopDown(fsm, nil)
//	fmt.Println(graph)
//
// Output:
//
//	graph TD
//	  StateA --> |EventX| StateB
//	  StateB --> |EventY| StateC
func MermaidGraphByTopDown[E, S constraints.Ordered](fsm FSM[E, S], transform TextTransformFunc) string {
	var buf strings.Builder
	buf.WriteString("\ngraph TD\n")

	for _, k := range fsm.defineSequence {
		src := fmt.Sprintf("%v", k.src)
		event := fmt.Sprintf("%v", k.event)
		dest := fmt.Sprintf("%v", fsm.transitions[k])
		if transform != nil {
			src = transform(src)
			event = transform(event)
			dest = transform(dest)
		}
		buf.WriteString(fmt.Sprintf(`  %v --> |%v| %v`, src, event, dest))
		buf.WriteString("\n")
	}
	return buf.String()
}

type TextTransformFunc func(text string) string
