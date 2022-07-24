package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/alecthomas/participle/v2/lexer"
	"gitlab.com/coalang/go-coa/try2/compile"
	"gitlab.com/coalang/go-coa/try2/parser"
	"gitlab.com/coalang/go-coa/try2/vm"
)

func main_() error {
	var err error
	var filepath string
	var allowParallel bool
	flag.StringVar(&filepath, "path", "", "path of file to run")
	flag.BoolVar(&allowParallel, "parallel", true, "allow parallel evaluation")
	flag.Parse()

	if n := flag.NArg(); n != 0 {
		parser.OsArgs = []string{filepath}
		for i := 0; i < n; i++ {
			parser.OsArgs = append(parser.OsArgs, flag.Arg(i))
		}
	}

	env := parser.NewEnv(lexer.Position{
		Filename: "root",
	}, allowParallel)
	root, err := env.LoadPathOnly(filepath)
	if err != nil {
		val, err := parser.ReturnVals(err)
		if err != nil {
			return err
		}
		if f, ok := val.(parser.BecomesFloat64); ok {
			code := int(f.BecomeFloat64())
			os.Exit(code)
		}
	}

	ce := compile.NewEnv(lexer.Position{Filename: "root"})
	s := ce.NewScope()
	insts, err := s.CompileNodes(*root)
	if err != nil {
		return err
	}
	s2 := compile.Instructions(insts).String()
	log.Printf("instructions:\n%s", s2)

	{
		p := vm.NewProgram(insts)
		v := vm.NewVM()
		err = v.Execute(p)
		if err != nil {
			return err
		}
	}
	return nil
}

func main() {
	err := main_()
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "error lul\n%s\n", err)
		os.Exit(1)
	}
}
