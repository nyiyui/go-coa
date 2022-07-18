package parser

func toNodes(evalers []Evaler) []Node {
	re := make([]Node, len(evalers))
	for i, evaler := range evalers {
		re[i] = toNode(evaler)
	}
	return re
}

func toNode(evaler Evaler) Node {
	switch evaler := evaler.(type) {
	case *Block:
		return Node{Block: evaler}
	case *Bool:
		return Node{Evaler: evaler}
	case *Call:
		return Node{Call: evaler}
	case *List:
		return Node{List: evaler}
	case *Map:
		return Node{Evaler: evaler}
	case *Native:
		return Node{Evaler: evaler}
	case *Nodes:
		return Node{Evaler: evaler}
	case *Number:
		return Node{Number: evaler}
	case *ID:
		return Node{ID: evaler}
	case *String:
		return Node{String_: evaler}
	case *Rune:
		return Node{Rune: evaler}
	default:
		return Node{Evaler: evaler}
	}
}
