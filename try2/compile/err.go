package compile

import (
	"fmt"

	"github.com/alecthomas/participle/v2/lexer"
)

// TLNCError is returned when a top-level non-call is encountered in a block.
type TLNCError struct {
	Pos lexer.Position
}

func (t *TLNCError) Error() string {
	return fmt.Sprintf("%s: top-level non-call in block", t.Pos)
}

// Error wraps an error returned by the compiler. This is used to track a sort-of "return-trace" for diagnostics and position.
type Error struct {
	Pos lexer.Position
	Err error
}

// NewError creates a new Error.
func NewError(pos lexer.Position, err error) *Error { return &Error{pos, err} }

func (c *Error) Error() string {
	return fmt.Sprintf("%s: %s", c.Pos, c.Err)
}
