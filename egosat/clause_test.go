package egosat

import "testing"

// TestSimplify checks the simplify method of the Clause struct
func TestSimplify(t *testing.T) {
	solver := Solver{
		assignments: make([]Lbool, 3),
	}
	clause := Clause{learnt: false, activity: 0.0, lits: []Lit{-1, -1, 2}}
	clause.simplify(&solver)
}
