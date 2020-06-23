package egosat

import (
	"fmt"
	"math/rand"
)

// The SolverParams struct stores the solver parameters pertaining to search.
type SolverParams struct {
	MaxConflict         int     // Number of conflicts before restart is required
	MaxLearnts          int     // Maximum number of learnt clauses to store at one time
	VarActivityDecay    float32 // Decay factor for variable activities
	ClauseActivityDecay float32 // Decay factor for clause activities
}

// The Solver struct contains the formula as well as the state of the solver
// over the course of solving the formulae.
type Solver struct {
	clauses             []*Clause
	learntClauses       []*Clause
	clauseActivityInc   float32
	clauseActivityDecay float32
	varActivityInc      float32
	varActivityDecay    float32
	variableActivity    []float32
	variableOrder       *queue
	watcherLists        [][]*Clause
	propQueue           []Lit
	assignments         []Lbool
	trail               []Lit
	trailDelim          []int
	reasons             []*Clause
	level               []int
}

// CreateSolver creates a new Solver for a formulae with the given number of
// variables and clauses. The number of variables given must be exact, but the
// number of clauses is only used for pre-allocation of dynamically sized data
// structures.
func CreateSolver(nClauses, nVars int) *Solver {
	solver := &Solver{
		clauses:          make([]*Clause, 0, nClauses),
		learntClauses:    make([]*Clause, 0, 100),
		watcherLists:     make([][]*Clause, 2*nVars),
		assignments:      make([]Lbool, nVars+1),
		trail:            make([]Lit, 0, nVars),
		reasons:          make([]*Clause, nVars+1),
		level:            make([]int, nVars+1),
		variableActivity: make([]float32, nVars+1),
	}
	solver.variableOrder = createQueue(solver, nVars)
	for i := 1; i <= nVars; i++ {
		solver.variableActivity[i] = 1e6
		solver.variableOrder.insert(i)
	}
	return solver
}

// AddClause adds a clause to the Solver. The provided literals make up the
// clause and the learnt flag indicates whether the clause is learnt, i.e.
// deduced from the original formula, or part of the original formula. In
// general, the case learnt=true should only be used by the internals of the
// solver.
func (solver *Solver) AddClause(lits []Lit, learnt bool) (bool, *Clause) {
	if !learnt {
		seen := make(map[Lit]bool)
		for _, l := range lits {
			if solver.litValue(l) == LTRUE {
				return true, nil
			}
			if _, ok := seen[l.negation()]; ok {
				return true, nil
			}
			seen[l] = true
		}
	}
	if len(lits) == 0 {
		return false, nil
	}
	if len(lits) == 1 {
		return solver.enqueue(lits[0], nil), nil
	}
	clause := &Clause{
		lits:     lits,
		learnt:   learnt,
		activity: 0.0,
	}
	if learnt {
		solver.learntClauses = append(solver.learntClauses, clause)
	} else {
		solver.clauses = append(solver.clauses, clause)
	}
	solver.addWatcher(lits[0].negation(), clause)
	solver.addWatcher(lits[1].negation(), clause)
	return true, clause
}

// Search will probe variable assignments until it either:
//      i) Finds a satisfying assignment
//      ii) Finds a conflict at the root level, meaning the formula is UNSAT
//      iii) Reaches the conflict limit
// If the conflict limit is reached, no conclusion can be drawn about whether
// the formula is satisfiable or not. In the case of (iii), Search can be
// reinvoked until (i) or (ii) occur.
func (solver *Solver) Search(params SolverParams) Lbool {
	var conflict *Clause
	var numConflicts int
	for {
		conflict = solver.propagate()
		if conflict != nil {
			numConflicts++
			solver.bumpClause(conflict)
			for _, l := range conflict.lits {
				solver.bumpVar(l.variable())
			}
			if solver.decisionLevel() == 0 {
				return LFALSE
			}
			learnt, level := solver.analyze(conflict)
			solver.cancelUntil(level)
			solver.record(learnt)
			solver.varActivityInc *= params.VarActivityDecay
			solver.clauseActivityInc *= params.ClauseActivityDecay
		} else {
			if len(solver.learntClauses) > params.MaxLearnts {
				solver.trimLearnts()
			}
			if solver.numAssigns() == solver.numVariables() {
				if solver.checkAsg() {
					return LTRUE
				}
				panic("invalid satisfying assignment detected through search")
			}
			if numConflicts > params.MaxConflict {
				solver.cancelUntil(0)
				return LNULL
			}
			l := solver.pickVar()
			solver.assume(l)
		}
	}
}

// PrintModel should only be invoked when the solver has found a satisfying
// assignment. When invoked it will print the satisfying assignment to stdout in
// the DIMACS output format.
func (solver *Solver) PrintModel() {
	fmt.Print("v ")
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
	fmt.Print("0\n")
}

// decisionLevel returns the current decision level of the solver.
func (solver *Solver) decisionLevel() int { return len(solver.trailDelim) }

// numVariables returns the number of variables in the formula.
func (solver *Solver) numVariables() int { return len(solver.assignments) - 1 }

// numClauses returns the number of clauses in the formula.
func (solver *Solver) numClauses() int { return len(solver.clauses) }

// numLearnts returns the number of learnt clauses in the formula.
func (solver *Solver) numLearnts() int {
	return len(solver.learntClauses)
}

// numAssigns returns the number of assignments that have been made.
func (solver *Solver) numAssigns() (n int) {
	for _, val := range solver.assignments[1:] {
		if val != LNULL {
			n++
		}
	}
	return
}

// varValue returns the current assignment to the given variable.
func (solver *Solver) varValue(variable int) Lbool {
	return solver.assignments[variable]
}

// litValues returns the current assignment to the given Lit.
func (solver *Solver) litValue(lit Lit) Lbool {
	varAsg := solver.assignments[lit.variable()]
	if varAsg == LNULL {
		return LNULL
	}
	if lit.polarity() == varAsg {
		return LTRUE
	}
	return LFALSE
}

// addWatcher adds a clause to the watch list of a literal.
func (solver *Solver) addWatcher(lit Lit, clause *Clause) {
	i := lit.index()
	solver.watcherLists[i] = append(solver.watcherLists[i], clause)
}

// removeWatcher removes a clause from the watch list of a literal.
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

// clearWatchers removes and returns all clauses from the watcher list of the
// given literal.
func (solver *Solver) clearWatchers(lit Lit) (clauses []*Clause) {
	clauses = solver.watcherLists[lit.index()]
	solver.watcherLists[lit.index()] = []*Clause{}
	return
}

// enqueue adds a literal to the propagation queue.
func (solver *Solver) enqueue(lit Lit, from *Clause) bool {
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

// Removes and returns the front of the propagation queue.
func (solver *Solver) dequeue() (lit Lit) {
	lit = solver.propQueue[0]
	solver.propQueue = solver.propQueue[1:]
	return
}

// undoOne undoes a single assignment.
func (solver *Solver) undoOne() {
	l := solver.trail[len(solver.trail)-1]
	v := l.variable()
	solver.assignments[v] = LNULL
	solver.reasons[v] = nil
	solver.level[v] = -1
	solver.trail = solver.trail[:len(solver.trail)-1]
	solver.variableOrder.insert(v)
}

// assume will force the given literal to be true by assigning its variable.
func (solver *Solver) assume(lit Lit) bool {
	solver.trailDelim = append(solver.trailDelim, len(solver.trail))
	return solver.enqueue(lit, nil)
}

// cancel will undo the most recent assumption and all assignments that followed
// from unit propagation.
func (solver *Solver) cancel() {
	numDel := len(solver.trail) - solver.trailDelim[len(solver.trailDelim)-1]
	for ; numDel > 0; numDel-- {
		solver.undoOne()
	}
	solver.trailDelim = solver.trailDelim[:len(solver.trailDelim)-1]
}

// cancel decisions until at the given decision level.
func (solver *Solver) cancelUntil(level int) {
	for solver.decisionLevel() > level {
		solver.cancel()
	}
}

// record adds a learnt clause.
func (solver *Solver) record(lits []Lit) {
	_, c := solver.AddClause(lits, true)
	solver.enqueue(lits[0], c)
}

// varActivityCmp compares the activity of two variables.
func (solver *Solver) varActivityCmp(var1 int, var2 int) bool {
	return solver.variableActivity[var1] < solver.variableActivity[var2]
}

// propagate invokes clause propagation for all watchers of each literal in the
// queue until the queue is empty
func (solver *Solver) propagate() *Clause {
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

// analyze generates a learnt clause from the given conflict clause and the
// state of the solver. It returns the learnt clause and the decision level at
// which the learnt clauses becomes unit.
func (solver *Solver) analyze(confl *Clause) (learnt []Lit, level int) {
	learnt = []Lit{0}
	var seen = make([]bool, solver.numVariables()+1)
	var counter = 0
	var p Lit = Lit(0)
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
			confl = solver.reasons[p.variable()]
			solver.undoOne()
			if seen[p.variable()] {
				break
			}
		}
		counter--
		if counter < 1 {
			break
		}
	}
	learnt[0] = p.negation()
	return
}

// pickVar selects the highest activity unbound variable for assumption.
func (solver *Solver) pickVar() Lit {
	for {
		v := solver.variableOrder.removeMin()
		if solver.assignments[v] == LNULL {
			if rand.Float64() < 0.5 {
				return Lit(-1 * v)
			}
			return Lit(v)
		}
	}
}

// bumpVar increases the activity level of the given variable and rescales all
// activities if necessary.
func (solver *Solver) bumpVar(v int) {
	solver.variableActivity[v] += solver.varActivityInc
	if solver.variableOrder.contains(v) {
		solver.variableOrder.moveUp(v)
	}
	if solver.variableActivity[v] > 1e6 {
		for i := 1; i < len(solver.variableActivity); i++ {
			solver.variableActivity[i] *= 1e-6
		}
	}
}

// bumpClause increases the activity level of the given clause and rescales all
// clause activities if necessary.
func (solver *Solver) bumpClause(c *Clause) {
	c.activity += solver.clauseActivityInc
	if c.activity > 1e6 {
		for i := 0; i < len(solver.learntClauses); i++ {
			solver.learntClauses[i].activity *= 1e-6
		}
	}
}

// checkAsg checks that the current assignment satisfies all clauses.
func (solver *Solver) checkAsg() bool {
	for _, c := range solver.clauses {
		violated := true
		for _, l := range c.lits {
			if solver.litValue(l) == LTRUE {
				violated = false
				break
			}
		}
		if violated {
			return false
		}
	}
	return true
}

// trimLearnts removes the least active half of the learnt clauses from the
// formula.
func (solver *Solver) trimLearnts() {
	solver.sortLearnts(0, len(solver.learntClauses)-1)
	for i := 0; i < (len(solver.learntClauses) / 2); i++ {
		solver.learntClauses[i].removeWatched(solver)
	}
	solver.learntClauses = solver.learntClauses[len(solver.learntClauses)/2:]
}

// sortLearnts will sort the learnt clauses in place according to their activity
// levels.
func (solver *Solver) sortLearnts(low, high int) {
	if low < high {
		p := solver.partition(low, high)
		solver.sortLearnts(low, p-1)
		solver.sortLearnts(p+1, high)
	}
}

// parition is a helper of sort learnts which returns i an index of the learnt
// clause list such that all values in learnts[low:i] are < learnts[i] and all
// values in [learnts[i+1:high]] are greater than learnts[i].
func (solver *Solver) partition(low, high int) int {
	pivot := solver.learntClauses[high]
	i := low
	for j := low; j <= high; j++ {
		if solver.learntClauses[j].activity < pivot.activity {
			solver.learntClauses[i], solver.learntClauses[j] = // swap values
				solver.learntClauses[j], solver.learntClauses[i]
			i++
		}
	}
	solver.learntClauses[i], solver.learntClauses[high] = // swap values
		solver.learntClauses[high], solver.learntClauses[i]
	return i
}
