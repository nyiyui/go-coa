package test

import (
	"bytes"
	"testing"

	"github.com/alecthomas/participle/v2/lexer"
	"gitlab.com/coalang/go-coa/try2/compile"
	"gitlab.com/coalang/go-coa/try2/parser"
	"gitlab.com/coalang/go-coa/try2/vm"
)

type TestCase struct {
	root *parser.Nodes
}

func NewTestCase(name, source string) (*TestCase, error) {
	var err error
	root := parser.Nodes{}
	source2 := bytes.NewBufferString(source)
	err = parser.Parser.Parse(name, source2, &root)
	if err != nil {
		return nil, err
	}
	err = parser.Check(&root)
	if err != nil {
		return nil, err
	}
	return &TestCase{&root}, nil
}

type log interface {
	Fatal(...interface{})
}

func testCase(l log, name, source string) *TestCase {
	tc, err := NewTestCase(name, source)
	if err != nil {
		l.Fatal(err)
	}
	return tc
}

type Engine int

const (
	EngineInvalid Engine = iota
	EngineInterp
	EngineVM
)

type TestCaseConfig struct {
	Engine   Engine
	Parallel bool
}

func (tc *TestCase) Run(b *testing.B, cfg TestCaseConfig) {
	switch cfg.Engine {
	case EngineInterp:
		env := parser.NewEnv(lexer.Position{
			Filename: "root",
		}, false)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, err := tc.root.Eval(env)
			if err != nil {
				b.Fatal(err)
			}
		}
	case EngineVM:
		var insts []compile.Instruction
		var err error
		{
			ce := compile.NewEnv(lexer.Position{Filename: "root"})
			s := ce.NewScope()
			insts, err = s.CompileNodes(*tc.root)
			if err != nil {
				b.Fatal(err)
			}
		}

		{
			p := vm.NewProgram(insts)
			v := vm.NewVM()
			// log.Println(compile.Instructions(insts))
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				err = v.Execute(p)
				if err != nil {
					b.Fatal(err)
				}
			}
		}

	default:
		b.Fatal("unsupported engine")
	}
}

func (tc *TestCase) Test(t *testing.T, cfg TestCaseConfig) {
	switch cfg.Engine {
	case EngineInterp:
		env := parser.NewEnv(lexer.Position{
			Filename: "root",
		}, false)
		_, err := tc.root.Eval(env)
		if err != nil {
			t.Fatal(err)
		}
	case EngineVM:
		var insts []compile.Instruction
		var err error
		{
			ce := compile.NewEnv(lexer.Position{Filename: "root"})
			s := ce.NewScope()
			insts, err = s.CompileNodes(*tc.root)
			if err != nil {
				t.Fatal(err)
			}
		}

		{
			p := vm.NewProgram(insts)
			v := vm.NewVM()
			// log.Println(compile.Instructions(insts))
			err = v.Execute(p)
			if err != nil {
				t.Fatal(err)
			}
		}

	default:
		t.Fatal("unsupported engine")
	}
}
