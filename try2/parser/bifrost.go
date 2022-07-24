package parser

import "fmt"

var Bifrost = &bifrost{}

type bifrost struct{}

func (b *bifrost) Profile(env IEnv) {
	fmt.Println(env.Dump())
}

func (b *bifrost) Peek(env IEnv, evalers []Evaler) {
	dump := env.Dump()
	dump.Vars = nil
	fmt.Println(dump)
	for i, evaler := range evalers {
		fmt.Printf("%d:\n\tuses: %s\n\tsets: %s\n\tresources: %s\n\t%s\n",
			i, evaler.IDUses(), evaler.IDSets(), evaler.Info(env), evaler.Inspect())
	}
}
