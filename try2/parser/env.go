package parser

import (
	"fmt"
	"log"
	"strings"
	"sync"

	"github.com/alecthomas/participle/v2/lexer"
	"gitlab.com/coalang/go-coa/try2/common"
	"gitlab.com/coalang/go-coa/try2/util"
)

type Env struct {
	Pos            lexer.Position
	outer          *Env
	lone           bool
	vars           map[string]Evaler
	varsLock       sync.RWMutex
	resources      map[string]map[string]*sync.Mutex
	resourceLock   sync.Mutex
	hooks          []hook
	hookNames      []string
	hooksLock      sync.Mutex
	allowParallel  bool
	ResourcesGuard ResourcesGuard
	debug          bool
}

func (e *Env) AllowParallel2() bool { return e.allowParallel }
func (e *Env) Debug2() bool         { return e.debug }
func (e *Env) Pos2() lexer.Position { return e.Pos }

func (e *Env) Printf(format string, v ...interface{}) {
	_ = log.Output(3, strings.Repeat("\t", e.StackLen())+fmt.Sprintf(format, v...))
}

var _ common.Env = (*Env)(nil)

type hook func() (keep bool)

type Hook = hook

func NewEnv(pos lexer.Position, allowParallel bool) *Env {
	return &Env{
		Pos:           pos,
		vars:          newBase(),
		allowParallel: allowParallel,
		resources:     map[string]map[string]*sync.Mutex{},
		debug:         true,
	}
}

func (e *Env) StackLen() int {
	if e == nil {
		return 0
	}
	return 1 + e.outer.StackLen()
}

type EnvDump struct {
	Pos  lexer.Position
	Lone bool
	Vars map[string]string
}

func (e EnvDump) String() string {
	lone := ""
	if e.Lone {
		lone = ", lone"
	}
	vars := ""
	if e.Vars != nil {
		vars = ":\n"
		builtinCount := 0
		for key, value := range e.Vars {
			if util.IsBuiltin(key) {
				builtinCount++
				continue
			}
			vars += fmt.Sprintf("(@def %s %s)\n", key, value)
		}
		vars += fmt.Sprintf("# %d builtin(s) elided", builtinCount)
	}
	return fmt.Sprintf("Dump (%s%s)%s", e.Pos, lone, vars)
}

func (e *Env) Dump() *EnvDump {
	vars := map[string]string{}
	for key, value := range e.vars {
		vars[key] = value.Inspect()
	}
	return &EnvDump{
		Pos:  e.Pos,
		Lone: e.lone,
		Vars: vars,
	}
}

func (e *Env) checkResource(r util.Resource) bool {
	if e == nil {
		return true
	}
	if e.ResourcesGuard == nil {
		return true
	}
	return e.ResourcesGuard.Allowed(r) && e.outer.checkResource(r)
}

func (e *Env) CheckResources(rs []util.Resource) bool {
	if e == nil {
		return true
	}
	if e.ResourcesGuard == nil {
		return true
	}
	for _, r := range rs {
		if !e.ResourcesGuard.Allowed(r) {
			return false
		}
	}
	return e.outer.CheckResources(rs)
}

func (e *Env) BadResources(rs []util.Resource) []int {
	re := make([]int, 0)
	for i, r := range rs {
		if !e.checkResource(r) {
			re = append(re, i)
		}
	}
	return re
}

func (e *Env) ensureResourceLock(r util.ResourceDef, arg string) {
	e.resourceLock.Lock()
	defer e.resourceLock.Unlock()
	if _, ok := e.resources[r.Name]; !ok {
		e.resources[r.Name] = map[string]*sync.Mutex{}
	}
	if _, ok := e.resources[r.Name][arg]; !ok {
		e.resources[r.Name][arg] = new(sync.Mutex)
	}
}

func (e *Env) LockResources(pos lexer.Position, rs []util.ResourceDef, args []string) {
	for i, r := range rs {
		e.lockResource(pos, r, args[i])
	}
}

func (e *Env) lockResource(_ lexer.Position, r util.ResourceDef, arg string) {
	e.ensureResourceLock(r, arg)
	e.getMutex(r.Name, arg).Lock()
}

func (e *Env) UnlockResources(pos lexer.Position, rs []util.ResourceDef, args []string) {
	for i, r := range rs {
		e.unlockResource(pos, r, args[i])
	}
}

func (e *Env) unlockResource(_ lexer.Position, r util.ResourceDef, arg string) {
	e.getMutex(r.Name, arg).Unlock()
}

func (e *Env) getMutex(name, arg string) *sync.Mutex {
	e.resourceLock.Lock()
	defer e.resourceLock.Unlock()
	return e.resources[name][arg]
}

func (e *Env) InheritLone(pos lexer.Position) IEnv {
	e.varsLock.RLock()
	defer e.varsLock.RUnlock()
	return &Env{
		Pos:           pos,
		vars:          map[string]Evaler{},
		lone:          true,
		allowParallel: e.allowParallel,
		resources:     map[string]map[string]*sync.Mutex{},
		outer:         e,
		debug:         e.debug,
	}
}

func (e *Env) Inherit(pos lexer.Position) IEnv {
	e.varsLock.RLock()
	defer e.varsLock.RUnlock()
	return &Env{
		Pos:           pos,
		vars:          map[string]Evaler{},
		allowParallel: e.allowParallel,
		resources:     map[string]map[string]*sync.Mutex{},
		outer:         e,
		debug:         e.debug,
	}
}

func (e *Env) Get(key string) (Evaler, bool) {
	if e == nil {
		return nil, false
	}
	e.varsLock.RLock()
	defer e.varsLock.RUnlock()
	return e._get(key)
}

func (e *Env) _get(key string) (Evaler, bool) {
	if e == nil {
		return nil, false
	}
	evaler, ok := e.vars[key]
	if ok {
		return evaler, true
	}
	if e.lone && !util.IsBuiltin(key) {
		return nil, false
	}
	return e.outer.Get(key)
}

func (e *Env) Has(key string) bool {
	_, ok := e.Get(key)
	return ok
}

func (e *Env) _has(key string) bool {
	_, ok := e._get(key)
	return ok
}

func (e *Env) HasKeys(keys []string) bool {
	e.varsLock.RLock()
	defer e.varsLock.RUnlock()
	for _, key := range keys {
		if !e._has(key) {
			return false
		}
	}
	return true
}

func (e *Env) Keys() []string {
	if e == nil {
		return nil
	}
	return append(e.MyKeys(), e.outer.Keys()...)
}

func (e *Env) MyKeys() []string {
	e.varsLock.Lock()
	defer e.varsLock.Unlock()
	re := make([]string, 0, len(e.vars))
	for key := range e.vars {
		re = append(re, key)
	}
	return re
}

func (e *Env) Def(key string, evaler Evaler) {
	e.varsLock.Lock()
	defer e.varsLock.Unlock()
	e.vars[key] = evaler
	e.callHooksConcurrent()
}

func (e *Env) Mod(key string, evaler Evaler) {
	if e.outer.Has(key) {
		e.outer.Mod(key, evaler)
	}
	e.Def(key, evaler)
}

func (e *Env) callHooksConcurrent() {
	go e.callHooks()
}

func (e *Env) callHooks() {
	e.hooksLock.Lock()
	defer e.hooksLock.Unlock()
	next := make([]hook, 0, len(e.hooks))
	nextNames := make([]string, 0, len(e.hooks))
	for i, hook := range e.hooks {
		if hook() {
			next = append(next, hook)
			nextNames = append(nextNames, e.hookNames[i])
		}
	}
	e.hooks = next
	e.hookNames = nextNames
}

func (e *Env) AddHook(name string, f hook) {
	e.hooksLock.Lock()
	defer e.hooksLock.Unlock()
	e.hooks = append(e.hooks, f)
	e.hookNames = append(e.hookNames, name)
	e.callHooksConcurrent()
}
