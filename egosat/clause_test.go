package egosat

import "testing"

// TestSimplify checks the simplify method of the Clause struct.
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
		clauses: []Clause{
			Clause{lits: []Lit{-1, 2, 3}},
			Clause{lits: []Lit{1, -2, 3}},
			Clause{lits: []Lit{1, 2, -3}},
		},
		assignments: []Lbool{LNULL, LTRUE, LNULL, LNULL},
	}
	solver.watcherLists = [][]*Clause{
		[]*Clause{&solver.clauses[0]},                     // watch lists for 1
		[]*Clause{&solver.clauses[1], &solver.clauses[2]}, // watch lists for -1
		[]*Clause{&solver.clauses[1]},
		[]*Clause{&solver.clauses[0], &solver.clauses[2]},
		[]*Clause{},
		[]*Clause{},
	}
	solver.clauses[0].propagate(&solver, Lit(1))
	if len(solver.watcherLists[Lit(-3).index()]) != 1 {
		t.Fail()
	}
	if solver.watcherLists[Lit(-3).index()][0] != &solver.clauses[0] {
		t.Fail()
	}
}
