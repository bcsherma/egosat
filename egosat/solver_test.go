package egosat

import (
	"testing"
)

func TestAddClause(t *testing.T) {
	nVars := 10
	solver := &Solver{
		clauses:       make([]Clause, 0, 10),
		learntClauses: make([]Clause, 0, 100),
		watcherLists:  make([][]*Clause, 2*nVars),
		assignments:   make([]Lbool, nVars+1),
		trail:         make([]Lit, 0, nVars),
		reasons:       make([]*Clause, nVars+1),
		level:         make([]int, nVars+1),
	}
	if ok, _ := solver.addClause([]Lit{}, false); ok != false {
		t.Fail()
	}
	solver.addClause([]Lit{1}, false)
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
	solver.addClause([]Lit{-1, 2, 3}, false)
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
		clauses: []Clause{
			{lits: []Lit{1, 2, 3}},
		},
		watcherLists: make([][]*Clause, 6),
	}
	solver.addWatcher(Lit(-2), &solver.clauses[0])
	if len(solver.watcherLists[Lit(-2).index()]) != 1 {
		t.Fail()
	}
	if solver.watcherLists[Lit(-2).index()][0] != &solver.clauses[0] {
		t.Fail()
	}
}

func TestRemoveWatcher(t *testing.T) {
	solver := Solver{clauses: []Clause{{lits: []Lit{1, 2, 3}}}}
	solver.watcherLists = [][]*Clause{
		{},
		{&solver.clauses[0]},
		{},
		{&solver.clauses[0]},
		{},
		{},
	}
	if len(solver.watcherLists[Lit(-2).index()]) != 1 {
		t.Fail()
	}
	solver.removeWatcher(Lit(-2), &solver.clauses[0])
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
	solver.Enqueue(2, nil)
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
	if !solver.Enqueue(2, nil) {
		t.Fail()
	}
	if solver.Enqueue(-2, nil) {
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
	solver.Enqueue(2, nil)
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

func TestUndoOne(t *testing.T) {
	solver := &Solver{
		assignments: []Lbool{LNULL, LTRUE, LFALSE, LNULL, LTRUE},
		level:       []int{0, 1, 2, -1, 1},
		trail:       []Lit{1, 4, -2},
		clauses:     []Clause{{lits: []Lit{-1, -4}}},
	}
	solver.reasons = []*Clause{nil, nil, nil, nil, &solver.clauses[0]}
	solver.undoOne()
	if solver.assignments[2] != LNULL || solver.level[2] != -1 {
		t.Fail()
	}
	if len(solver.trail) != 2 {
		t.Fail()
	}
	solver.undoOne()
	if solver.assignments[4] != LNULL || solver.level[4] != -1 || solver.reasons[4] != nil {
		t.Fail()
	}
}

func TestPropagate(t *testing.T) {
	nVars := 10
	solver := &Solver{
		clauses:       make([]Clause, 0, 10),
		learntClauses: make([]Clause, 0, 10),
		watcherLists:  make([][]*Clause, 2*nVars),
		assignments:   make([]Lbool, nVars+1),
		trail:         make([]Lit, 0, nVars),
		reasons:       make([]*Clause, nVars+1),
		level:         make([]int, nVars+1),
	}
	solver.addClause([]Lit{-1, -2}, false)
	solver.addClause([]Lit{2, -3}, false)
	solver.assume(1)
	solver.Propagate()
	if solver.varValue(2) != LFALSE {
		t.Fail()
	}
	if solver.varValue(3) != LFALSE {
		t.Fail()
	}
}

func TestAnalyze(t *testing.T) {
	nVars := 10
	solver := &Solver{
		clauses:       make([]Clause, 0, 10),
		learntClauses: make([]Clause, 0, 10),
		watcherLists:  make([][]*Clause, 2*nVars),
		assignments:   make([]Lbool, nVars+1),
		trail:         make([]Lit, 0, nVars),
		reasons:       make([]*Clause, nVars+1),
		level:         make([]int, nVars+1),
	}
	solver.addClause([]Lit{-1, -2}, false)
	solver.addClause([]Lit{2, -3}, false)
	solver.addClause([]Lit{-1, 2, 3}, false)
	solver.assume(1)
	confl := solver.Propagate()
	if confl == nil {
		t.Fail()
	}
	learnt, level := solver.Analyze(confl)
	if level != 0 {
		t.Fail()
	}
	t.Log(learnt, level)
}
