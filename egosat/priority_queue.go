package egosat

type queue struct {
	heap    []Lit   // Heap storage
	indices []int   // Maps elements to their indices in the heap
	solver  *Solver // Reference to solver for access to literal activities
}

// These functions are used for computing the indices of the parent and children
// of heap nodes
func parent(idx int) int     { return (idx - 1) >> 1 }
func leftChild(idx int) int  { return 2*idx + 1 }
func rightChild(idx int) int { return 2*idx + 2 }

// createQueue  generates a new queue for the given solver and with the given
// capactity preallocated for the heap and the index map
func createQueue(solver *Solver, capacity int) (q *queue) {
	q = &queue{
		heap:    make([]Lit, 0, capacity),
		indices: make([]int, 2*capacity),
		solver:  solver,
	}
	for i := 0; i < len(q.indices); i++ {
		q.indices[i] = -1
	}
	return
}

// contains returns true if the queue contains the given element else false
func (q *queue) contains(l Lit) bool {
	if q.indices[l.index()] != -1 {
		return true
	}
	return false
}

// insert adds a new integer/variable into the heap
func (q *queue) insert(l Lit) {
	q.heap = append(q.heap, l)
	q.indices[l.index()] = len(q.heap) - 1
	q.moveUp(l)
}

// removeMax pops the maxmimum key from the heap
func (q *queue) removeMax() Lit {
	ret := q.heap[0]
	q.indices[ret.index()] = -1
	q.heap[0] = q.heap[len(q.heap)-1]
	q.heap = q.heap[:len(q.heap)-1]
	if len(q.heap) > 0 {
		q.indices[q.heap[0].index()] = 0
		q.moveDown(q.heap[0])
	}
	return ret
}

// moveUp identifies an element of the heap and swaps the element with its
// parent until the heap property is respected locally
func (q *queue) moveUp(l Lit) {
	i := q.indices[l.index()]
	a := q.priority(i)
	for i > 0 && q.priority(parent(i)) < a {
		q.indices[q.heap[parent(i)].index()] = i
		q.heap[i] = q.heap[parent(i)]
		i = parent(i)
	}
	q.indices[l.index()] = i
	q.heap[i] = l
}

// moveDown identifies and element of the heap and swaps it with either of its
// children until the heap property is respected locally
func (q *queue) moveDown(l Lit) {
	i := q.indices[l.index()]
	a := q.priority(i)
	var j int
	for leftChild(i) < len(q.heap) {
		if rightChild(i) >= len(q.heap) || q.priority(leftChild(i)) > q.priority(rightChild(i)) {
			j = leftChild(i)
		} else {
			j = rightChild(i)
		}
		if q.priority(j) < a {
			break
		}
		q.indices[q.heap[j].index()] = i
		q.heap[i] = q.heap[j]
		i = j
	}
	q.indices[l.index()] = i
	q.heap[i] = l
}

func (q *queue) priority(i int) float64 {
	return q.solver.literalActivity[q.heap[i].index()]
}
