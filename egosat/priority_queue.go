package egosat

type queue struct {
	heap    []int       // Heap storage
	indices map[int]int // Maps elements to their indices in the heap
	solver  *Solver     // Reference to solver for access to literal activities
}

// These functions are used for computing the indices of the parent and children
// of heap nodes
func parent(idx int) int     { return (idx - 1) >> 1 }
func leftChild(idx int) int  { return 2*idx + 1 }
func rightChild(idx int) int { return 2*idx + 2 }

// createQueue  generates a new queue for the given solver and with the given
// capactity preallocated for the heap and the index map
func createQueue(solver *Solver, capacity int) *queue {
	return &queue{
		heap:    make([]int, 0, capacity),
		indices: make(map[int]int, capacity),
		solver:  solver,
	}
}

// contains returns true if the queue contains the given element else false
func (q *queue) contains(v int) bool {
	_, ok := q.indices[v]
	return ok
}

// insert adds a new integer/variable into the heap
func (q *queue) insert(n int) {
	q.heap = append(q.heap, n)
	q.indices[n] = len(q.heap) - 1
	q.moveUp(n)
}

// removeMax pops the maxmimum key from the heap
func (q *queue) removeMax() int {
	ret := q.heap[0]
	delete(q.indices, ret)
	q.heap[0] = q.heap[len(q.heap)-1]
	q.heap = q.heap[:len(q.heap)-1]
	if len(q.heap) > 0 {
		q.indices[q.heap[0]] = 0
		q.moveDown(q.heap[0])
	}
	return ret
}

// moveUp identifies an element of the heap and swaps the element with its
// parent until the heap property is respected locally
func (q *queue) moveUp(n int) {
	i := q.indices[n]
	a := q.priority(i)
	for i > 0 && q.priority(parent(i)) < a {
		q.indices[q.heap[parent(i)]] = i
		q.heap[i] = q.heap[parent(i)]
		i = parent(i)
	}
	q.indices[n] = i
	q.heap[i] = n
}

// moveDown identifies and element of the heap and swaps it with either of its
// children until the heap property is respected locally
func (q *queue) moveDown(n int) {
	i := q.indices[n]
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
		q.indices[q.heap[j]] = i
		q.heap[i] = q.heap[j]
		i = j
	}
	q.indices[n] = i
	q.heap[i] = n
}

func (q *queue) priority(i int) float64 {
	return q.solver.variableActivity[q.heap[i]]
}
