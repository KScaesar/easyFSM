package easyFSM

import (
	"fmt"
	"strings"

	"golang.org/x/exp/constraints"
)

func MermaidGraphByTopDown[E, S constraints.Ordered](fsm *FSM[E, S], transform TextTransformFunc) string {
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
