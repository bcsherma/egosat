package egosat

type Lit int

type Lbool int

const (
	LNULL  = iota
	LTRUE  = iota
	LFALSE = iota
)

// polarity returns the polarity of this literal
func (lit Lit) polarity() Lbool {
	if lit > 0 {
		return LTRUE
	}
	return LFALSE
}

// variable returns the variable constituent of this literal
func (lit Lit) variable() int {
	if lit > 0 {
		return int(lit)
	}
	return -1 * int(lit)
}

// negation returns the negation of this literal
func (lit Lit) negation() Lit {
	return -1 * lit
}

// index returns the index associated with this literal
func (lit Lit) index() (idx int) {
	idx = 2 * (lit.variable() - 1)
	if lit.polarity() == LFALSE {
		idx++
	}
	return
}
