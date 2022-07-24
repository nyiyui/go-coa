package vm

import (
	"fmt"
	"log"
	"strings"

	"gitlab.com/coalang/go-coa/try2/compile"
	"gitlab.com/coalang/go-coa/try2/parser"
)

type Instructions struct {
	pos    string
	offset int
	insts  []compile.Instruction
	sn     *scopeSnapshot
}

func (i *Instructions) Evaler() parser.Evaler { return nil }

func (i *Instructions) Run(v *VM) error {
	return nil
}

func (i *Instructions) VMCall(v *VM) (Value, error) {
	p := &Program{offset: i.offset, insts: i.insts}
	log.Println("Running instructions", p)
	for i, inst := range i.insts {
		log.Printf("%d: %s", i, &inst)
	}
	v.s().sn = i.sn
	err := v.exec(p)
	if err != nil {
		return nil, err
	}
	v.logCurrent()
	returned := v.popFrame()
	log.Println("VMCall returned", returned)
	return returned, nil
}

func (i *Instructions) String() string {
	b := new(strings.Builder)
	fmt.Fprintf(b, "Instructions<%d insts @%s>", len(i.insts), i.pos)
	fmt.Fprintf(b, "\n%s", i.sn)
	return b.String()
}
