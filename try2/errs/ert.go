package errs

import (
	"fmt"
	"github.com/alecthomas/participle/v2/lexer"
	"gitlab.com/coalang/go-coa/try2/util"
	"strconv"
)


type Errors []error

func (errs Errors) Error() string {
	if errs == nil {
		return "<nil>"
	}
	if len(errs) == 0 {
		return "0 errors"
	}
	if len(errs) == 1 {
		if errs[0] == nil {
			return "1 <nil> error"
		}
		return errs[0].Error()
	}
	re := strconv.FormatInt(int64(len(errs)), 10) + " errors:\n"
	for _, err := range errs {
		if err == nil {
			re += "<nil>\n\n"
		}
		re += util.Indent(err.Error()) + "\n\n"
	}
	return re
}

func AppendERT(err error, frame ERTFrame) *ERT {
	if e, ok := err.(*ERT); ok {
		e.frames = append(e.frames, frame)
		return e
	} else {
		return &ERT{
			frames: []ERTFrame{
				frame,
			},
			Err: err,
		}
	}
}

type ERT struct {
	frames []ERTFrame
	Err    error
}

func (e *ERT) String() string {
	re := "error:\n"
	re += fmt.Sprintf("%T", e.Err)
	re += util.Indent(fmt.Sprint(e.Err)) + "\n"
	re += fmt.Sprintf("trace (%d frame(s)):\n", len(e.frames))
	for i, frame := range e.frames {
		re += fmt.Sprintf("%d%s\n", i, util.Indent(frame.String()))
	}
	return re
}

func (e *ERT) Error() string {
	return e.String()
}

type ERTFrame struct {
	Pos  lexer.Position
	Call string
}

func (e *ERTFrame) String() string {
	return fmt.Sprintf("%s\n%s", e.Pos, util.Indent(e.Call))
}
