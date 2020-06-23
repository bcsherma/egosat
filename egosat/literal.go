package egosat

// Lit type is used to represent literals, i.e. a variable or it's negation.
type Lit int

// Lbool is used to represent boolean values with the possibility of null,
// or undetermined.
type Lbool int

const (
	// LNULL indicates a non-true, non-false boolean value.
	LNULL = Lbool(iota)
	// LTRUE indicates a true boolean value.
	LTRUE = Lbool(iota)
	// LFALSE indicates the boolean value false.
	LFALSE = Lbool(iota)
)

// polarity returns the polarity of this literal.
func (lit Lit) polarity() Lbool {
	if lit > 0 {
		return LTRUE
	}
	return LFALSE
}

// variable returns the variable component of this literal. For example,
// Lit(-1).variable() returns 1, as does Lit(1).variable().
func (lit Lit) variable() int {
	if lit > 0 {
		return int(lit)
	}
	return -1 * int(lit)
}

// negation returns the negation of this literal.
func (lit Lit) negation() Lit {
	return -1 * lit
}

// index returns the index associated with this literal in the solver data
// structures.
func (lit Lit) index() (idx int) {
	idx = 2 * (lit.variable() - 1)
	if lit.polarity() == LFALSE {
		idx++
	}
	return
}
