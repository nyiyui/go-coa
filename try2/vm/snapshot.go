package vm

import (
	"fmt"
	"strings"
)

type varSnapshot struct {
	name string
	v2   Value
}

type scopeSnapshot struct {
	matrix [][]varSnapshot
}

func (s *scopeSnapshot) set(level, index int, vs varSnapshot) {
	if level >= len(s.matrix) {
		// extend s.matrix to fit level
		missing := level - len(s.matrix) + 1
		s.matrix = append(s.matrix, make([][]varSnapshot, missing)...)
	}
	if index >= len(s.matrix[level]) {
		// extend s.matrix[level] to fit level
		missing := index - len(s.matrix[level]) + 1
		s.matrix[level] = append(s.matrix[level], make([]varSnapshot, missing)...)
	}
	s.matrix[level][index] = vs
}

func (s *scopeSnapshot) get(level, index int) varSnapshot {
	return s.matrix[level][index]
}

func (s *scopeSnapshot) String() string {
	b := new(strings.Builder)
	for level, s := range s.matrix {
		for index, vs := range s {
			fmt.Fprintf(b, "%d:%d: %s = %s\n", level, index, vs.name, vs.v2)
		}
	}
	return b.String()
}
