package parser

import (
	"fmt"
)

func toEvaler(v interface{}) (Evaler, error) {
	switch v := v.(type) {
	case float64:
		return NewNumber(v), nil
	case int:
		return NewNumber(float64(v)), nil
	case map[string]interface{}:
		m := &Map{Content: map[string]Evaler{}}
		var err error
		for key, val := range v {
			m.Content[key], err = toEvaler(val)
			if err != nil {
				return nil, err
			}
		}
		return m, nil
	case string:
		return NewString(v), nil
	case []interface{}:
		l := &List{
			Content: Nodes{Content: make([]Node, len(v))},
		}
		var err error
		var evaler Evaler
		for i, val := range v {
			evaler, err = toEvaler(val)
			if err != nil {
				return nil, err
			}
			l.Content.Content[i] = Node{Evaler: evaler}
		}
		return l, nil
	case bool:
		return &Bool{Content: v}, nil
	default:
		return nil, fmt.Errorf("unsupported type %T", v)
	}
}
