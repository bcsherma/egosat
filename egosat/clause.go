package egosat

// Clause is the struct used for storing all CNF clauses
type Clause struct {
	learnt   bool    // Indicates whether clause was learnt or not
	activity float32 // Gives the activity of the clause
	lits     []Lit   //
}

// simplify simplifies returns true if the invoking clause is trivially
// satisfiable, and otherwise will eliminate any false literals from the clause
// before returning false.
func (clause *Clause) simplify(solver *Solver) bool {
	var j, val Lbool
	for _, l := range clause.lits {
		val = solver.litValue(l)
		if val == LTRUE {
			return true
		}
		if val == LNULL {
			clause.lits[j] = l
			j++
		}
	}
	clause.lits = clause.lits[0:j]
	return false
}

// propagate will enqueue unit information if the clause has become unit. It is
// assumed (and enforced at the beginning of the function) that lit is the first
// literal in the invoking clause. Another precondition is that lit is not being
// watched by the invoking clause. If any literals in the clause beyond the
// first two literals are not yet falsified, the clause will be set to watch one
// of those.
func (clause *Clause) propagate(solver *Solver, lit Lit) bool {
	if clause.lits[0] == lit.negation() {
		clause.lits[0], clause.lits[1] = clause.lits[1], clause.lits[0]
	}
	if solver.litValue(clause.lits[0]) == LTRUE {
		solver.addWatcher(lit, clause)
		return true
	}
	for i := 2; i < len(clause.lits); i++ {
		if solver.litValue(clause.lits[i]) != LFALSE {
			clause.lits[1], clause.lits[i] = clause.lits[i], clause.lits[1]
			solver.addWatcher(clause.lits[1].negation(), clause)
			return true
		}
	}
	solver.addWatcher(lit, clause)
	return solver.enqueue(clause.lits[0], clause)
}

// calcReason will compute the assignments that force the clause to be
// conflicting. As a precondition, lit will either be LNULL (0) or the first
// literal in the invoking clause. If p is LNULL, this method will return the
// negation of every literal in the clause, otherwise this method will return
// the negation of every literal except the first literal.
func (clause *Clause) calcReason(lit Lit) (reason []Lit) {
	var i int
	if lit == Lit(0) {
		i = 0
	} else {
		i = 1
	}
	for ; i < len(clause.lits); i++ {
		reason = append(reason, clause.lits[i].negation())
	}
	return
}

// removeWatched will remove the clause from the watcher lists of its first two
// literals
func (clause *Clause) removeWatched(solver *Solver) {
	for i := 0; i < 2; i++ {
		solver.removeWatcher(clause.lits[i], clause)
	}
}
