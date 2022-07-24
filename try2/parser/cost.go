package parser

import "sync"

var costCache = map[string]uint64{}
var costCacheMu sync.RWMutex

func costCacheSet(key string, value uint64) {
	costCacheMu.Lock()
	defer costCacheMu.Unlock()
	costCache[key] = value
}

func costOfResource(res string) (cost uint64) {
	switch res {
	case "http":
		return 1e6
	case "io.stdin", "io.stdout", "io.stderr":
		return 0.2e6
	case "fs.local":
		return 0.1e6
	default:
		return
	}
}

func costOfEvaling(env *Env, evaler Evaler) (cost uint64) {
	if evaler == nil {
		return
	}
	{
		var ok bool
		costCacheMu.RLock()
		cost, ok = costCache[evaler.String()]
		costCacheMu.RUnlock()
		if ok {
			return
		}
	}
	for _, res := range evaler.Info(env).Resources {
		cost += costOfResource(res.Name)
	}
	switch evaler := evaler.(type) {
	case *Globber, *Regexer, *SysEnv, *Number, *Rune, *Time, *Bool, *Map, *Native:
		return 0
	case *Block:
		for _, node := range evaler.Content.Content {
			cost += costOfEvaling(env, node.Select())
		}
		return
	case *String:
		for _, id := range stringTmplPattern.FindAllString(evaler.Content, -1) {
			cost += costOfEvaling(env, &ID{Content: id})
		}
		return
	case *Call:
		for _, node := range evaler.Content.Content {
			cost += costOfEvaling(env, node.Select())
		}
		return
	case *List:
		for _, node := range evaler.Content.Content {
			cost += costOfEvaling(env, node.Select())
		}
		return
	case *Nodes:
		for _, node := range evaler.Content {
			cost += costOfEvaling(env, node.Select())
		}
		return
	case *ID:
		evaler2, ok := env.Get(evaler.Content)
		if !ok {
			return 0
		}
		// unknown!
		return costOfEvaling(env, evaler2)
	default:
		return 100
	}
}
