package main

import (
	_ "embed"
	"fmt"
	"github.com/alecthomas/participle/v2"
	"github.com/alecthomas/participle/v2/lexer"
	"gitlab.com/coalang/go-coa/try2/parser"
	"log"
	"time"
)

//go:embed main.coa
var src string

func _main() error {
	root := parser.Nodes{}
	filename := "try2/main/main.coa"
	err := parser.Parser.ParseString(filename, src, &root)
	if err != nil {
		if err, ok := err.(participle.UnexpectedTokenError); ok {
			return fmt.Errorf("%s: expected %s, got %s (%v)", err.Position(), err.Expected, err.Unexpected, err.Unexpected.Type)
		}
		return err
	}
	env := parser.NewEnv(lexer.Position{Filename: filename}, true)

	start := time.Now()
	_, err = root.Eval(env, 0)
	end := time.Now()
	if err != nil {
		return fmt.Errorf("error return trace (most recent call last):\n%w", err)
	}
	log.Printf("done in %s", end.Sub(start))
	return nil
}

func main() {
	err := _main()
	if err != nil {
		panic(err)
	}
}
