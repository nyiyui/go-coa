package parser

import (
	"encoding/json"
	"fmt"

	"github.com/alecthomas/participle/v2/lexer"
	"gitlab.com/coalang/go-coa/try2/util"
)

type MapLike interface {
	Get(key string) (Evaler, bool, error)
	Set(key string, value Evaler) error
	Keys() (keys []string)
}

type Map struct {
	Pos     lexer.Position
	Content map[string]Evaler
	keys    []string
	iterI   int
}

func (m *Map) Len() int {
	if m.keys == nil || len(m.keys) != len(m.Content) {
		m.keys = make([]string, 0, len(m.Content))
		for key := range m.Content {
			m.keys = append(m.keys, key)
		}
	}
	return len(m.keys)
}

func (m *Map) Index(i int) (key, value Evaler) {
	if m.keys == nil || len(m.keys) != len(m.Content) {
		m.keys = make([]string, 0, len(m.Content))
		for key := range m.Content {
			m.keys = append(m.keys, key)
		}
	}
	k := m.keys[i]
	return NewString(k), m.Content[k]
}

var _ Evaler = new(Map)
var _ MapLike = new(Map)
var _ Iter = new(Map)

func (m *Map) MarshalJSON() ([]byte, error) {
	return json.Marshal(m.Content)
}

func (m *Map) UnmarshalJSON(data []byte) error {
	return json.Unmarshal(data, &m.Content)
}

func newMapFromNodes(nodes Nodes) (*Map, error) {
	m := &Map{
		Pos:     nodes.Pos,
		Content: map[string]Evaler{},
	}
	evalers := nodes.Select()
	if l := len(evalers); l%2 != 0 {
		return nil, fmt.Errorf("map lit len must be even, not %d", l)
	}
	var key, value Evaler
	for i := 0; i < len(evalers); i += 2 {
		key, value = evalers[i], evalers[i+1]
		if key2, ok := key.(BecomesString); !ok {
			return nil, fmt.Errorf("%s: key must be string, not %T", GetPos(key), key)
		} else {
			m.Content[key2.BecomeString()] = value
		}
	}
	return m, nil
}
func (m *Map) Info(env IEnv) util.Info {
	re := make([]util.ResourceDef, 0)
	for _, value := range m.Content {
		re = append(re, value.Info(env).Resources...)
	}
	return util.Info{Resources: re}
}
func (m *Map) Eval(_ IEnv) (result Evaler, err error) { return m, nil }
func (m *Map) String() string {
	re := "[m\n"
	for key, val := range m.Content {
		re += util.Indent((&String{Content: key}).String()+" "+val.String()) + "\n"
	}
	return re[:len(re)-1] + "\n]"
}
func (m *Map) Inspect() string {
	re := "[m\n"
	for key, val := range m.Content {
		re += util.Indent((&String{Content: key}).Inspect()+" "+val.Inspect()) + "\n"
	}
	return re[:len(re)-1] + "\n]"
}
func (m *Map) IDUses() []string {
	re := make([]string, 0)
	for _, value := range m.Content {
		re = append(re, value.IDUses()...)
	}
	return re
}
func (m *Map) IDSets() []string {
	re := make([]string, 0)
	for _, value := range m.Content {
		re = append(re, value.IDSets()...)
	}
	return re
}
func (m *Map) Get(s string) (Evaler, bool, error) { evaler, ok := m.Content[s]; return evaler, ok, nil }
func (m *Map) Set(s string, evaler Evaler) error  { m.Content[s] = evaler; return nil }
func (m *Map) Keys() (keys []string) {
	keys = make([]string, 0, len(m.Content))
	for key := range m.Content {
		keys = append(keys, key)
	}
	return keys
}
