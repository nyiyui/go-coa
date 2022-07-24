package parser

import (
	"fmt"
	"log"
	"sync"
	"time"

	errs3 "gitlab.com/coalang/go-coa/try2/errs"
	"gitlab.com/coalang/go-coa/try2/util"
)

// evalParallel evaluates evalers in parallel.
// TODO: support access to resources in parallel
func evalParallel(env IEnv, evalers []Evaler) ([]Evaler, error) {
	env.Printf("parallel %d", len(evalers))
	if len(evalers) == 0 || len(evalers) == 1 {
		panic("evalParallel must not be called with 1 or 0 evalers")
	}
	errs := make(chan error, len(evalers))
	times := make([]time.Duration, len(evalers))
	dones := make(chan int, len(evalers))

	split := func(evalers []Evaler) ([][]int, error) {
		type evalerIndex = int
		type depIndex = int

		evalersDeps := map[evalerIndex]map[depIndex]struct{}{}
		{
			varIndexes := map[string]int{}
			for i, evaler := range evalers {
				evalersDeps[i] = map[depIndex]struct{}{}
				uses := util.NoArguments(util.NoBuiltins(evaler.IDUses()))
				sets := util.NoArguments(util.NoBuiltins(evaler.IDSets()))
				for _, use := range uses {
					index, ok := varIndexes[use]
					if !ok {
						panic("key use not found in varIndexes")
					}
					evalersDeps[i][index] = struct{}{}
				}
				for _, set := range sets {
					varIndexes[set] = i
				}
			}
		}

		var strands [][]int
		strandsM := map[int]int{}
		for i, evaler := range evalers {
			_ = evaler
			deps := evalersDeps[i]
			switch len(deps) {
			case 0:
				strands = append(strands, []int{i})
				strandsM[i] = len(strands) - 1
			case 1:
				var dep depIndex
				for key := range deps {
					dep = key
					break
				}
				j, ok := strandsM[dep]
				if !ok {
					panic("key i not found in strandsM")
				}
				strands[j] = append(strands[j], i)
			default:
				var newRel []int
				newRelIndex := len(strands) // not yet added
				for dep := range deps {
					j, ok := strandsM[dep]
					if !ok {
						panic("key i not found in strandsM")
					}
					newRel = append(newRel, strands[j]...)
					strandsM[dep] = newRelIndex // others might use strandsM[dep]
					strands[j] = nil            // empty strand (don't remove it, that would make things complicated)
				}
				strands = append(strands, newRel)
			}
		}
		// TODO: check this function's code
		return strands, nil
	}

	merge := func(evalers []Evaler, strands [][]int) []Evaler {
		var sum []Evaler
		for _, strand := range strands {
			for _, i := range strand {
				sum = append(sum, evalers[i])
			}
		}
		return sum
	}

	run := func(i int) {
		log.Println("run ", i, GetPos(evalers[i]))
		defer func() { dones <- i }()
		defer log.Println("done", i)
		var err error
		start := time.Now()
		evalers[i], err = Eval(evalers[i], env)
		if err != nil {
			errs <- err
		}
		end := time.Now()
		times[i] = end.Sub(start)
	}

	ensure := func(evaler Evaler, uses []string) {
		var ok sync.Mutex
		ok.Lock()
		env.AddHook(fmt.Sprintf("eval %s", GetPos(evaler)), func() (keep bool) {
			defer func() {
				if !keep {
					ok.Unlock()
				}
			}()

			keep = !env.HasKeys(uses)
			return
		})
		ok.Lock() // only runs after the hook calls Unlock
	}
	_ = ensure

	runStrand := func(indexes []int) {
		for _, i := range indexes {
			run(i)
		}
	}

	resources := func(env IEnv, indexes []int) []util.ResourceDef {
		var defs []util.ResourceDef
		for _, i := range indexes {
			defs = append(defs, evalers[i].Info(env).Resources...)
		}
		return defs
	}

	strands, err := split(evalers)
	if err != nil {
		return nil, err
	}
	{
		check := map[int]int{}
		for _, strand := range strands {
			for _, i := range strand {
				check[i]++
			}
		}
		result := ""
		for key, value := range check {
			if value != 1 {
				result += fmt.Sprintf("%v %v", key, value)
			}
		}
		if result != "" {
			panic("strands check failed" + result)
		}
	}
	for i, strand := range strands {
		if len(strand) != 0 {
			re := ""
			for _, j := range strand {
				re += evalers[j].Inspect() + "\n"
			}
			env.Printf("strand %d:\n%s", i, re)
			log.Println("res", resources(env, strand))
			_ = runStrand
			runStrand(strand)
		}
	}

	log.Println("waiting")
	i := 0
	for done := range dones {
		log.Println(i, done, len(evalers)-1)
		i++
		if i == len(evalers)-1 {
			break
		}
	}
	log.Println("done waiting")
	close(errs)
	select {
	case err, ok := <-errs:
		if !ok {
			break
		}
		if err == nil {
			panic("err from errs is nil")
		}
		errs2 := errs3.Errors{err}
		for err := range errs {
			if err == nil {
				continue
			}
			errs2 = append(errs2, err)
		}
		if len(errs2) == 1 {
			return nil, errs2[0]
		}
		return nil, errs2
	}
	var sum time.Duration
	for _, duration := range times {
		sum += duration
	}
	return merge(evalers, strands), nil
}

func evalersIsPure(env IEnv, evalers []Evaler) bool {
	for _, evaler := range evalers {
		if len(evaler.Info(env).Resources) > 0 {
			return false
		}
	}
	return true
}
