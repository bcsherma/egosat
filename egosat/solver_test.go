package egosat

import (
	"testing"
)

func TestAddClause(t *testing.T) {
	solver := CreateSolver(10, 10)
	if ok, _ := solver.AddClause([]Lit{}, false); ok != false {
		t.Fail()
	}
	solver.AddClause([]Lit{1}, false)
	if len(solver.clauses) > 0 {
		t.Fail()
	}
	if solver.varValue(1) != LTRUE {
		t.Fail()
	}
	if len(solver.propQueue) != 1 {
		t.Fail()
	}
	if solver.propQueue[0] != Lit(1) {
		t.Fail()
	}
	solver.AddClause([]Lit{-1, 2, 3}, false)
	if len(solver.clauses) != 1 {
		t.Fail()
	}
	for _, lit := range []Lit{1, -2} {
		if len(solver.watcherLists[lit.index()]) != 1 {
			t.Fail()
		}
	}
}

func TestAddWatcher(t *testing.T) {
	solver := Solver{
		clauses: []*Clause{
			{lits: []Lit{1, 2, 3}},
		},
		watcherLists: make([][]*Clause, 6),
	}
	solver.addWatcher(Lit(-2), solver.clauses[0])
	if len(solver.watcherLists[Lit(-2).index()]) != 1 {
		t.Fail()
	}
	if solver.watcherLists[Lit(-2).index()][0] != solver.clauses[0] {
		t.Fail()
	}
}

func TestRemoveWatcher(t *testing.T) {
	solver := Solver{clauses: []*Clause{{lits: []Lit{1, 2, 3}}}}
	solver.watcherLists = [][]*Clause{
		{},
		{solver.clauses[0]},
		{},
		{solver.clauses[0]},
		{},
		{},
	}
	if len(solver.watcherLists[Lit(-2).index()]) != 1 {
		t.Fail()
	}
	solver.removeWatcher(Lit(-2), solver.clauses[0])
	if len(solver.watcherLists[Lit(-2).index()]) != 0 {
		t.Fail()
	}
}

func TestVarValue(t *testing.T) {
	solver := &Solver{assignments: []Lbool{LNULL, LTRUE, LFALSE, LNULL}}
	if solver.varValue(1) != LTRUE {
		t.Fail()
	}
	if solver.varValue(2) != LFALSE {
		t.Fail()
	}
	if solver.varValue(3) != LNULL {
		t.Fail()
	}
}

func TestLitValue(t *testing.T) {
	solver := &Solver{assignments: []Lbool{LNULL, LTRUE, LFALSE, LNULL}}
	if solver.litValue(-1) != LFALSE || solver.litValue(1) != LTRUE {
		t.Fail()
	}
	if solver.litValue(-2) != LTRUE || solver.litValue(2) != LFALSE {
		t.Fail()
	}
	if solver.litValue(3) != LNULL || solver.litValue(-3) != LNULL {
		t.Fail()
	}
}

func TestEnqueue(t *testing.T) {
	nVars := 5
	solver := &Solver{
		assignments: make([]Lbool, nVars+1),
		reasons:     make([]*Clause, nVars+1),
		level:       make([]int, nVars+1),
	}
	solver.enqueue(2, nil)
	if len(solver.propQueue) != 1 {
		t.Fail()
	}
	if solver.propQueue[0] != Lit(2) {
		t.Fail()
	}
	if solver.varValue(2) != LTRUE {
		t.Fail()
	}
	if len(solver.trail) != 1 {
		t.Fail()
	}
	if solver.trail[0] != Lit(2) {
		t.Fail()
	}
	if !solver.enqueue(2, nil) {
		t.Fail()
	}
	if solver.enqueue(-2, nil) {
		t.Fail()
	}
}

func TestDequeue(t *testing.T) {
	nVars := 5
	solver := &Solver{
		assignments: make([]Lbool, nVars+1),
		reasons:     make([]*Clause, nVars+1),
		level:       make([]int, nVars+1),
	}
	solver.enqueue(2, nil)
	if len(solver.propQueue) != 1 {
		t.Fail()
	}
	if solver.propQueue[0] != Lit(2) {
		t.Fail()
	}
	if solver.dequeue() != Lit(2) {
		t.Fail()
	}
	if len(solver.propQueue) != 0 {
		t.Fail()
	}
}

func TestPropagate(t *testing.T) {
	solver := CreateSolver(10, 10)
	solver.AddClause([]Lit{-1, -2}, false)
	solver.AddClause([]Lit{2, -3}, false)
	solver.assume(1)
	solver.propagate()
	if solver.varValue(2) != LFALSE {
		t.Fail()
	}
	if solver.varValue(3) != LFALSE {
		t.Fail()
	}
}

func TestAnalyze(t *testing.T) {
	solver := CreateSolver(10, 10)
	solver.AddClause([]Lit{-1, -2}, false)
	solver.AddClause([]Lit{2, -3}, false)
	solver.AddClause([]Lit{-1, 2, 3}, false)
	solver.assume(1)
	confl := solver.propagate()
	if confl == nil {
		t.Fail()
	}
	learnt, level := solver.analyze(confl)
	if level != 0 {
		t.Fail()
	}
	if len(learnt) != 1 {
		t.Fail()
	}
	if learnt[0] != -1 {
		t.Fail()
	}
}

func TestSearch(t *testing.T) {
	solver := CreateSolver(10, 10)
	solver.AddClause([]Lit{1, 2}, false)
	solver.AddClause([]Lit{-1, 2}, false)
	solver.AddClause([]Lit{-1, -2}, false)
	solver.AddClause([]Lit{1, -2}, false)
	params := SolverParams{
		MaxConflict:         100,
		MaxLearnts:          100,
		VarActivityDecay:    0.95,
		ClauseActivityDecay: 0.65,
	}
	if solver.Search(params) != LFALSE {
		t.Fail()
	}
}

func TestSortLearnts(t *testing.T) {
	solver := CreateSolver(10, 10)
	_, c1 := solver.AddClause([]Lit{1, 2, 3}, true)
	c1.activity = 3.0
	_, c2 := solver.AddClause([]Lit{1, 2, 3}, true)
	c2.activity = 1.0
	_, c3 := solver.AddClause([]Lit{1, 2, 3}, true)
	c3.activity = 2.0
	_, c4 := solver.AddClause([]Lit{1, 2, 3}, true)
	c4.activity += 4.0
	solver.sortLearnts(0, len(solver.learntClauses)-1)
	if solver.learntClauses[0] != c2 {
		t.Fail()
	}
	if solver.learntClauses[1] != c3 {
		t.Fail()
	}
	if solver.learntClauses[2] != c1 {
		t.Fail()
	}
	if solver.learntClauses[3] != c4 {
		t.Fail()
	}
}
