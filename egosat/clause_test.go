package egosat

import "testing"

// TestSimplify tests the behavior of the Clause.simplify method.
func TestSimplify(t *testing.T) {
	// Test that clauses with true literals will result in true being returned
	solver := Solver{assignments: []Lbool{LNULL, LFALSE}}
	clause := Clause{learnt: false, activity: 0.0, lits: []Lit{-1}}
	if !clause.simplify(&solver) {
		t.Fail()
	}
	// Test that false literals will be removed from clauses
	solver = Solver{assignments: []Lbool{LNULL, LTRUE, LNULL}}
	clause = Clause{learnt: false, activity: 0.0, lits: []Lit{-1, 2}}
	clause.simplify(&solver)
	if clause.lits[0] != Lit(2) || len(clause.lits) != 1 {
		t.Fail()
	}
}

// TestClausePropagate will test that unit information is correctly propagated
// from clauses and that watcher lists are updated appropriately in light of new
// assignments.
func TestClausePropagate(t *testing.T) {
	solver := Solver{
		clauses: []*Clause{
			{lits: []Lit{-1, 2, 3}},
			{lits: []Lit{1, -2, 3}},
			{lits: []Lit{1, 2, -3}},
		},
		assignments: []Lbool{LNULL, LTRUE, LNULL, LNULL},
	}
	solver.watcherLists = [][]*Clause{
		{solver.clauses[0]},                    // watch lists for 1
		{solver.clauses[1], solver.clauses[2]}, // watch lists for -1
		{solver.clauses[1]},
		{solver.clauses[0], solver.clauses[2]},
		{},
		{},
	}
	// The first clause should be added to the watcher list of -3
	solver.clauses[0].propagate(&solver, Lit(1))
	if len(solver.watcherLists[Lit(-3).index()]) != 1 {
		t.Fail()
	}
	if solver.watcherLists[Lit(-3).index()][0] != solver.clauses[0] {
		t.Fail()
	}
}

// TestCalcReason tests the calc_reason method of the clause struct. This method
// should return the literal assignments that force make the clause unsatisifed.
// If the lit argument of calcreason is not null then it is assumed to be the
// first literal of the clause and is not returned.
func TestCalcReason(t *testing.T) {
	c := &Clause{lits: []Lit{1, 2, 3}}
	ret := c.calcReason(Lit(1))
	if len(ret) != 2 {
		t.Fail()
	}
	if ret[0] != Lit(-2) || ret[1] != Lit(-3) {
		t.Fail()
	}
	ret = c.calcReason(Lit(0))
	if len(ret) != 3 {
		t.Fail()
	}
	if ret[0] != Lit(-1) || ret[1] != Lit(-2) || ret[2] != Lit(-3) {
		t.Fail()
	}
}
