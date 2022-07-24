package parser

import (
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	errs2 "gitlab.com/coalang/go-coa/try2/errs"
	"gitlab.com/coalang/go-coa/try2/util"
)

type Strand = strand

func CompileEvalers(keys []string, evalers []Evaler) ([]*strand, error) {
	deps := getEvalersDeps(keys, evalers)
	strands := reverseDepsStrands(cleanStrands(getStrands(evalers, deps)))
	var err error
	strands, err = checkStrands(evalers, strands)
	if err != nil {
		return nil, err
	}
	return strands, nil
}

func evalParallel2(env IEnv, evalers []Evaler) ([]Evaler, error) {
	log.Println("evalParallel2", env, env.Keys())
	strands, err := CompileEvalers(env.Keys(), evalers)
	if err != nil {
		return nil, err
	}
	if len(strands) == 1 {
		// short-circuit
		if env.Debug2() {
			env.Printf("sc %d", len(evalers))
		}
		s := strands[0]
		for _, i := range s.Todo {
			evalers[i], err = Eval(evalers[i], env)
			if err != nil {
				return nil, err
			}
		}
	}
	if env.Debug2() {
		env.Printf("parallel %d", len(evalers))
	}
	r := runStrands(env, evalers, strands)
	err = r.waitResults(env, strands, evalers)
	if env.Debug2() {
		env.Printf("parallel %d done", len(evalers))
	}
	return evalers, err
	//return evalers, waitResults(env, strands, evalers, ch)
}

func (r *runEnv) waitResults(env IEnv, ss []*strand, evalers []Evaler) error {
	recvd := 0
	var poss []int
	errs := errs2.Errors{}
	for r := range r.ch {
		recvd++
		//env.printf("done %d/%d %s %v %v", r.Index, len(evalers), r.Time, evalers[r.Index], r.Error)
		poss = append(poss, r.Index)
		if r.Error != nil {
			errs = append(errs, r.Error)
		}
		if recvd == len(evalers) {
			break
		}
	}
	if len(errs) != 0 {
		if len(errs) == 1 {
			return errs[0]
		}
		return errs
	}
	return nil
}

/*
func waitResults(env IEnv, ss []*strand, evalers []Evaler, ch chan result) error {
	recvd := 0
	var poss []int
	go func() {
		for range time.Tick(1 * time.Second) {
			if recvd == len(evalers) {
				break
			}
			env.printf(strings.Repeat("-", 100))
			env.printf("%d/%d %d %s %v", recvd, len(evalers), env.StackLen(), env.Pos, poss)
			for i, pos := range poss {
				env.printf("\t%d %d %s", i, pos, inspect(evalers[pos]))
			}
			env.printf("--------")
			for i := range evalers {
				env.printf("\t%d %d %s", i, i, inspect(evalers[i]))
			}
			env.printf("--------")
			for i, s := range ss {
				printStrand(env, evalers, i, s)
				runStrand(env, evalers, ch, s)
			}
		}
	}()
	errs := errs2.Errors{}
	for r := range ch {
		recvd++
		env.printf("done %d/%d %s %v %v", r.Index, len(evalers), r.Time, evalers[r.Index], r.Error)
		poss = append(poss, r.Index)
		if r.Error != nil {
			errs = append(errs, r.Error)
		}
		if recvd == len(evalers) {
			break
		}
	}
	if len(errs) != 0 {
		if len(errs) == 1 {
			return errs[0]
		}
		return errs
	}
	return nil
}
*/

type (
	evalerIndex = int
	strandIndex = int
	depIndex    = int
	evalersDeps = map[evalerIndex]evalerDep
	evalerDep   = map[depIndex]struct{}
	result      struct {
		Index evalerIndex
		Error error
		Time  time.Duration
	}
)

type strand struct {
	Deps        []strandIndex
	ReverseDeps []strandIndex
	Todo        []evalerIndex

	// TODO: optimize based on runtime performance
	minTime time.Duration
	maxTime time.Duration
}

func (s *strand) updateTime(t time.Duration) {
	if s.minTime == 0 || s.minTime > t {
		s.minTime = t
	}
	if s.maxTime < t {
		s.maxTime = t
	}
}

func newStrand(todo ...evalerIndex) *strand { return &strand{Todo: todo} }

func (s *strand) appendDeps(deps ...strandIndex) { s.Deps = append(s.Deps, deps...) }

func (s *strand) appendTodo(todo ...evalerIndex) { s.Todo = append(s.Todo, todo...) }

const (
	outerScopeEvalerIndex = -1
)

func getEvalersDeps(keys []string, evalers []Evaler) evalersDeps {
	sDeps := evalersDeps{}
	varIndexes := map[string]int{}
	log.Println("keys", keys)
	for _, set := range util.NoBuiltins(keys) {
		log.Println("set", set)
		varIndexes[set] = outerScopeEvalerIndex
	}
	for i, evaler := range evalers {
		sDeps[i] = evalerDep{}
		uses := util.NoArguments(util.NoBuiltins(evaler.IDUses()))
		sets := util.NoArguments(util.NoBuiltins(evaler.IDSets()))
		for _, use := range uses {
			index, ok := varIndexes[use]
			if !ok {
				log.Println(use, GetPos(evaler))
				panic("key use not found in varIndexes")
			}
			switch index {
			case outerScopeEvalerIndex:
				// it's already there ⇒ ignore
			default:
				sDeps[i][index] = struct{}{}
			}
		}
		for _, set := range sets {
			varIndexes[set] = i
		}
	}
	return sDeps
}

func cleanStrands(ss []*strand) (cleaned []*strand) {
	for _, s := range ss {
		//if len(s) != 0 {
		if s != nil {
			cleaned = append(cleaned, s)
		}
	}
	return cleaned
}

func reverseDepsStrands(ss []*strand) []*strand {
	depss := map[strandIndex][]strandIndex{}
	// NOTE: deps always point before
	for i := len(ss) - 1; i >= 0; i-- {
		s := ss[i]
		for _, dep := range s.Deps {
			depss[dep] = append(depss[dep], i)
		}
		s.ReverseDeps = depss[i]
	}
	return ss
}

func checkStrands(evalers []Evaler, ss []*strand) ([]*strand, error) {
	evalersRan := make([]bool, len(evalers))
	for _, s := range ss {
		for _, i := range s.Todo {
			evalersRan[i] = true
		}
	}
	var result []string
	for i, ok := range evalersRan {
		if !ok {
			result = append(result, fmt.Sprint(i))
		}
	}
	if len(result) != 0 {
		panic(fmt.Errorf("unran indexes: %s", strings.Join(result, ", ")))
	}
	return ss, nil
}

func getStrands(evalers []Evaler, depss evalersDeps) []*strand {
	var strands []*strand
	strandsM := map[evalerIndex]strandIndex{} // stores which strand is each evaler is in
	for i := range evalers {
		deps := depss[i]
		switch len(deps) {
		case 0:
			// start
			strands = append(strands, newStrand(i))
			strandsM[i] = len(strands) - 1
		case 1:
			// next
			var dep int
			for dep2 := range deps {
				dep = dep2
				break
			}
			strandI := strandsM[dep]
			strandsM[i] = strandI
			strands[strandI].appendTodo(i)
		default:
			// join
			// TODO: if a dep strand has the same deps (except itself, of course) then merge
			strand := newStrand(i)
			for dep := range deps {
				j, ok := strandsM[dep]
				if !ok {
					panic("key i not found in strandsM")
				}
				/*
						depStrand := strands[j]
						if len(depStrand.Deps) == len(deps)-1 {
							theirDeps := map[strandIndex]struct{}{}
							for dep2 := range depStrand.Deps {
								theirDeps[dep2] = struct{}{}
							}
							theirDeps[j] = struct{}{}
							for myDep := range deps {
								_, ok := theirDeps[myDep]
								if !ok {
									goto notSame
								}
							}
							// same
							strandsM[i] = j
							strands[j].appendTodo(i)
							continue
						}
					notSame:
				*/
				strand.appendDeps(j)
			}
			strands = append(strands, strand)
			strandsM[i] = len(strands) - 1
			/*
				newRel := newStrand(i)
				newRelIndex := len(strands) // not yet added
				for dep := range deps {
					j, ok := strandsM[dep]
					if !ok {
						panic("key i not found in strandsM")
					}
					//newRel = append(newRel, j)
					//newRel = append(newRel, strands[j]...)
					newRel.appendTodo(j)
					newRel.appendTodo(strands[j].Todo...)
					strandsM[dep] = newRelIndex // others might use strandsM[dep]
					strands[j] = nil            // empty strand (don't remove it, that would make things complicated)
				}
				strands = append(strands, newRel)
				strandsM[i] = newRelIndex
			*/
		}
	}
	return strands
}

func printStrand(env IEnv, evalers []Evaler, i int, s *strand) {
	b := new(strings.Builder)
	fmt.Fprintf(b, "strand %d:\n", i)
	fmt.Fprintf(b, "\ttodo:\n")
	for _, j := range s.Todo {
		fmt.Fprintf(b, "\t\t%d %s\n", j, inspect(evalers[j]))
	}
	fmt.Fprintf(b, "\treverseDeps:\n")
	for _, j := range s.ReverseDeps {
		fmt.Fprintf(b, "\t\t%d\n", j)
	}
	fmt.Fprintf(b, "\tdeps:\n")
	for _, j := range s.Deps {
		fmt.Fprintf(b, "\t\t%d\n", j)
	}
	env.Printf("%s\n", b.String())
}

func inspect(evaler Evaler) string {
	if evaler == nil {
		return "<nil>"
	}
	return fmt.Sprintf("%s @ %s", evaler.Inspect(), GetPos(evaler))
}

type runEnv struct {
	strandsDepCount     []int
	strandsDepCountLock []sync.Mutex
	strandsStartCh      []chan struct{}
	ss                  []*strand
	evalers             []Evaler
	env                 IEnv
	ch                  chan result
}

func newRunEnv(env IEnv, evalers []Evaler, ss []*strand) *runEnv {
	n := len(evalers)
	re := &runEnv{
		strandsDepCount:     make([]int, n),
		strandsDepCountLock: make([]sync.Mutex, n),
		strandsStartCh:      make([]chan struct{}, n),
		ss:                  ss,
		evalers:             evalers,
		env:                 env,
		ch:                  make(chan result, n),
	}
	for i, s := range ss {
		re.strandsDepCount[i] = len(s.Deps)
		re.strandsDepCountLock[i] = sync.Mutex{}
		re.strandsStartCh[i] = make(chan struct{})
	}
	return re
}

func (re *runEnv) strands() []*strand { return re.ss }

func startStrands(ss []*strand) (starts []strandIndex) {
	starts = make([]strandIndex, 0)
	for i, s := range ss {
		if len(s.Deps) == 0 {
			starts = append(starts, i)
		}
	}
	return starts
}

func runStrands(env IEnv, evalers []Evaler, ss []*strand) *runEnv {
	r := newRunEnv(env, evalers, ss)
	starts := startStrands(ss)
	// TODO: run only start strands
	if env.Debug2() {
		for i, s := range ss {
			printStrand(env, evalers, i, s)
		}
	}
	for _, i := range starts {
		go r.runStartStrand(i)
	}
	//for i, s := range ss {
	//	printStrand(env, evalers, i, s)
	//	go r.runStrand(env, evalers, ch, i)
	//}
	return r
}

func (r *runEnv) signalDepDone(env IEnv, strandI strandIndex) {
	r.strandsDepCountLock[strandI].Lock()
	defer r.strandsDepCountLock[strandI].Unlock()
	r.strandsDepCount[strandI]--
	if r.strandsDepCount[strandI] < 0 {
		panic("strandI must have positive depCount")
	}
	if r.strandsDepCount[strandI] == 0 {
		r.strandsStartCh[strandI] <- struct{}{}
	}
}

func (r *runEnv) signalDepDone2(env IEnv, strandI strandIndex) bool {
	r.strandsDepCountLock[strandI].Lock()
	defer r.strandsDepCountLock[strandI].Unlock()
	r.strandsDepCount[strandI]--
	if r.strandsDepCount[strandI] < 0 {
		panic("strandI must have positive depCount")
	}
	return r.strandsDepCount[strandI] == 0
}

func (r *runEnv) runStartStrand(strandI strandIndex) {
	prefix := fmt.Sprintf("[strand %d] ", strandI)
	s := r.ss[strandI]
	start := time.Now()
	for _, i := range s.Todo {
		r.runEvaler(i)
	}
	took := time.Since(start)
	r.ss[strandI].updateTime(took)

	lastRDI := len(s.ReverseDeps) - 1
	for i, reverseDep := range s.ReverseDeps {
		lastRD := i == lastRDI
		rdLastDepDone := r.signalDepDone2(r.env, reverseDep)
		if rdLastDepDone {
			if lastRD {
				// takeover so we don't have to spawn superfluous goroutines
				if r.env.Debug2() {
					r.env.Printf("%stakeover %d", prefix, reverseDep)
				}
				r.runStartStrand(reverseDep)
			} else {
				// spawn
				if r.env.Debug2() {
					r.env.Printf("%sspawn %d", prefix, reverseDep)
				}
				go r.runStartStrand(reverseDep)
			}
		}
	}
}

func (r *runEnv) runStrand(env IEnv, evalers []Evaler, ch chan<- result, strandI strandIndex) {
	env.Printf("runStrand %d", strandI)
	<-r.strandsStartCh[strandI]
	env.Printf("runStrand start %d", strandI)
	s := r.ss[strandI]
	for _, i := range s.Todo {
		r.runEvaler(i)
	}
	env.Printf("runStrand sendSignals %d", strandI)
	for _, reverseDep := range s.ReverseDeps {
		r.signalDepDone(env, reverseDep)
	}
}

func (r *runEnv) runEvaler(i evalerIndex) {
	var err error
	start := time.Now()
	r.evalers[i], err = Eval(r.evalers[i], r.env)
	end := time.Now()
	r.ch <- result{
		Index: i,
		Error: err,
		Time:  end.Sub(start),
	}
}

func runEvaler(env IEnv, evalers []Evaler, ch chan<- result, i evalerIndex) {
	var err error
	start := time.Now()
	evalers[i], err = Eval(evalers[i], env)
	end := time.Now()
	ch <- result{
		Index: i,
		Error: err,
		Time:  end.Sub(start),
	}
}
