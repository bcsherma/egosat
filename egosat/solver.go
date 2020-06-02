package egosat

import "fmt"

type Solver struct {
	clauses             []Clause
	learntClauses       []Clause
	clauseActivityInc   float32
	clauseActivityDecay float32
	varActivityInc      float32
	varActivityDecay    float32
	variableActivity    []float32
	variableOrder       []int
	watcherLists        [][]*Clause
	propQueue           []Lit
	assignments         []Lbool
	trail               []Lit
	trailDelim          []int
	reasons             []*Clause
	level               []int
}

// decisionLevel returns the current decision level of the solver.
func (solver *Solver) decisionLevel() int {
	return len(solver.trailDelim)
}

// numVariables returns the number of variables in the formula
func (solver *Solver) numVariables() int {
	return len(solver.assignments) - 1
}

// numClauses returns the number of clauses in the formula
func (solver *Solver) numClauses() int {
	return len(solver.clauses)
}

// numLearnts returns the number of learnt clauses in the formula
func (solver *Solver) numLearnts() int {
	return len(solver.learntClauses)
}

// numAssigns returns the number of assignments that have been made
func (solver *Solver) numAssigns() (n int) {
	for _, val := range solver.assignments[1:] {
		if val != LNULL {
			n++
		}
	}
	return
}

// varValue returns the current assignment to the given variable
func (solver *Solver) varValue(variable int) Lbool {
	return solver.assignments[variable]
}

// addClause adds a new clause to the solver
func (solver *Solver) addClause(lits []Lit, learnt bool) bool {
	if !learnt {
		seen := make(map[Lit]bool)
		for _, l := range lits {
			if solver.litValue(l) == LTRUE {
				return true
			}
			if _, ok := seen[l.negation()]; ok {
				return true
			}
			seen[l] = true
		}
	}
	if len(lits) == 0 {
		return false
	}
	if len(lits) == 1 {
		return solver.Enqueue(lits[0], nil)
	}
	clause := Clause{
		lits:     lits,
		learnt:   learnt,
		activity: 0.0,
	}
	var c *Clause
	if learnt {
		solver.learntClauses = append(solver.learntClauses, clause)
		c = &solver.learntClauses[len(solver.learntClauses)-1]
	} else {
		solver.clauses = append(solver.clauses, clause)
		c = &solver.clauses[len(solver.clauses)-1]
	}
	solver.addWatcher(lits[0].negation(), c)
	solver.addWatcher(lits[1].negation(), c)
	return true
}

// addWatcher adds a clause to the watch list of a literal
func (solver *Solver) addWatcher(lit Lit, clause *Clause) {
	i := lit.index()
	solver.watcherLists[i] = append(solver.watcherLists[i], clause)
}

// removeWatcher removes a clause from the watch list of a literal
func (solver *Solver) removeWatcher(lit Lit, clause *Clause) {
	i := lit.index()
	for j := 0; j < len(solver.watcherLists[i]); j++ {
		if solver.watcherLists[i][j] == clause {
			solver.watcherLists[i] = append(
				solver.watcherLists[i][:j],
				solver.watcherLists[i][j+1:]...,
			)
		}
	}
}

func (solver *Solver) clearWatchers(lit Lit) (clauses []*Clause) {
	clauses = solver.watcherLists[lit.index()]
	solver.watcherLists[lit.index()] = []*Clause{}
	return
}

// litValues returns the current assignment to the given Lit
func (solver *Solver) litValue(lit Lit) int {
	varAsg := solver.assignments[lit.variable()]
	if varAsg == LNULL {
		return LNULL
	}
	if lit.polarity() == varAsg {
		return LTRUE
	}
	return LFALSE
}

// Enqueue adds a literal to the propagation queue
func (solver *Solver) Enqueue(lit Lit, from *Clause) bool {
	if solver.litValue(lit) != 0 {
		if solver.litValue(lit) == LTRUE {
			return true
		}
		return false
	}
	solver.assignments[lit.variable()] = lit.polarity()
	solver.level[lit.variable()] = solver.decisionLevel()
	solver.reasons[lit.variable()] = from
	solver.trail = append(solver.trail, lit)
	solver.propQueue = append(solver.propQueue, lit)
	return true
}

// Removes and returns the front of the queue
func (solver *Solver) dequeue() (lit Lit) {
	lit = solver.propQueue[0]
	solver.propQueue = solver.propQueue[1:]
	return
}

// undoOne undoes a single assignment
func (solver *Solver) undoOne() {
	l := solver.trail[len(solver.trail)-1]
	v := l.variable()
	solver.assignments[v] = LNULL
	solver.reasons[v] = nil
	solver.level[v] = -1
	solver.trail = solver.trail[:len(solver.trail)-1]
}

// Assumes one literal value
func (solver *Solver) assume(lit Lit) bool {
	solver.trailDelim = append(solver.trailDelim, len(solver.trail))
	return solver.Enqueue(lit, nil)
}

// Cancels an assumption and the resulting assignments
func (solver *Solver) cancel() {
	numDel := len(solver.trail) - solver.trailDelim[len(solver.trailDelim)-1]
	for ; numDel > 0; numDel-- {
		solver.undoOne()
	}
	solver.trailDelim = solver.trailDelim[:len(solver.trailDelim)-1]
}

// cancel decisions until at the given level
func (solver *Solver) cancelUntil(level int) {
	for solver.decisionLevel() > level {
		solver.cancel()
	}
}

// record adds a learnt clause
func (solver *Solver) record(lits []Lit) {
	solver.addClause(lits, true)
	solver.Enqueue(lits[0], &solver.learntClauses[len(solver.learntClauses)-1])
}

// Propagate invokes clause propagation for all watchers of each literal in the
// queue until the queue is empty
func (solver *Solver) Propagate() *Clause {
	for len(solver.propQueue) > 0 {
		l := solver.dequeue()
		tmp := solver.clearWatchers(l)
		for i := 0; i < len(tmp); i++ {
			if !tmp[i].propagate(solver, l) {
				for j := i + 1; j < len(tmp); j++ {
					solver.addWatcher(l, tmp[j])
				}
				solver.propQueue = []Lit{}
				return tmp[i]
			}
		}
	}
	return nil
}

// Analyze will assess the cause of a conflict
func (solver *Solver) Analyze(confl *Clause) (learnt []Lit, level int) {
	learnt = []Lit{0}
	var seen = make([]bool, solver.numVariables()+1)
	var counter = 0
	var p Lit = LNULL
	var reason []Lit
	for {
		reason = confl.calcReason(p)
		for j := 0; j < len(reason); j++ {
			var q = reason[j]
			if !seen[q.variable()] {
				seen[q.variable()] = true
				if solver.level[q.variable()] == solver.decisionLevel() {
					counter++
				} else if solver.level[q.variable()] > 0 {
					learnt = append(learnt, q.negation())
					if solver.level[q.variable()] > level {
						level = solver.level[q.variable()]
					}
				}
			}
		}
		for {
			p = solver.trail[len(solver.trail)-1]
			if seen[p.variable()] {
				break
			}
			confl = solver.reasons[p.variable()]
			solver.undoOne()
		}
		counter--
		if counter < 1 {
			break
		}
	}
	learnt[0] = p
	return
}

// pickVar selects a variable for assumption
func (solver *Solver) pickVar() Lit {
	for i := 1; i < len(solver.assignments); i++ {
		if solver.assignments[i] == LNULL {
			return Lit(i)
		}
	}
	panic("Unable to select a variable for assignment")
}

func (solver *Solver) Search() Lbool {
	var conflict *Clause
	var numConflicts int
	for {
		conflict = solver.Propagate()
		if conflict != nil {
			numConflicts++
			if solver.decisionLevel() == 0 {
				return LFALSE
			}
			learnt, level := solver.Analyze(conflict)
			solver.cancelUntil(level)
			solver.record(learnt)
		} else {
			if solver.numAssigns() == solver.numVariables() {
				return LTRUE
			}
			l := solver.pickVar()
			solver.assume(l)
		}
	}
}

// PrintAsg prints out the model
func (solver *Solver) PrintModel() {
	for i := 1; i < len(solver.assignments); i++ {
		switch solver.assignments[i] {
		case LFALSE:
			fmt.Printf("%d ", -1*i)
		case LTRUE:
			fmt.Printf("%d ", i)
		case LNULL:
			panic(fmt.Errorf("variable %d is unassigned", i))
		}
	}
}
