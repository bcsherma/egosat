package egosat

type Clause struct {
	learnt   bool
	activity float32
	lits     []Lit
}

// simplify will eliminate false literals from clauses
func (clause *Clause) simplify(solver *Solver) bool {
	var j, val int
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

// propagate will update the watched literals of a clause and detect if a
// clause has newly become unit
func (clause *Clause) propagate(solver *Solver, lit Lit) bool {
	if clause.lits[0] == lit {
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
	return solver.Enqueue(clause.lits[0], clause)
}

// calcReason will compute the assignments that force the clause to be
// conflicting
func (clause *Clause) calcReason(lit Lit) (reason []Lit) {
	var i int
	if lit == LNULL {
		i = 0
	} else {
		i = 1
	}
	for ; i < len(clause.lits); i++ {
		reason = append(reason, clause.lits[i])
	}
	return
}
