package amd

import (
	"container/heap"

	"gonum.org/v1/gonum/mat"
)

// Struct to represent a node in the graph
type Node struct {
	degree int
	index  int
}

// A Min-Heap of Nodes based on their degree
type MinHeap []Node

func (h MinHeap) Len() int           { return len(h) }
func (h MinHeap) Less(i, j int) bool { return h[i].degree < h[j].degree }
func (h MinHeap) Swap(i, j int)      { h[i], h[j] = h[j], h[i] }

func (h *MinHeap) Push(x any) {
	*h = append(*h, x.(Node))
}

func (h *MinHeap) Pop() any {
	old := *h
	n := len(old)
	x := old[n-1]
	*h = old[0 : n-1]
	return x
}

// Adjacency list returns a list where of lists where each list represents nodes where there is a non-zero
// element in the matrix.
func AdjacencyList(mat mat.NonZeroDoer) [][]int {
	adj := make(map[int]map[int]bool)
	mat.DoNonZero(func(i, j int, v float64) {
		if i == j {
			return
		}
		if _, ok := adj[i]; !ok {
			adj[i] = make(map[int]bool)
		}
		adj[i][j] = true
	})

	largestRow := -1
	for k := range adj {
		if k > largestRow {
			largestRow = k
		}
	}

	result := make([][]int, largestRow+1)
	for i := range result {
		result[i] = make([]int, 0)
	}

	for row, neightbours := range adj {
		for col := range neightbours {
			result[row] = append(result[row], col)
		}
	}
	return result
}

type AmdCtx struct {
	// Degree of all nodes
	Degrees []int

	// If true, then the node has been eliminated
	Eliminated []bool

	// Row ordering. This gradually grows during the AMD process
	Ordering []int

	// Heap is a structure that keeps track of the nodes which is makes
	// the process of finding the node with the minimum degree efficient
	Heap *MinHeap
}

func NewAmdCtx(n int) AmdCtx {
	hp := &MinHeap{}
	heap.Init(hp)
	return AmdCtx{
		Degrees:    make([]int, n),
		Heap:       hp,
		Eliminated: make([]bool, n),
		Ordering:   make([]int, 0, n),
	}
}

type NodeDegree interface {
	// UpdateNodes returns a list of nodes for which the degree should be
	// updated
	UpdateNodes(eliminated Node, adjList [][]int, ctx *AmdCtx) []int

	// Calculates the degree of node
	Degree(node int, adjList [][]int, ctx *AmdCtx) int

	// Called when a node is eliminated
	OnNodeEliminated(eliminated Node, adjList [][]int, ctx *AmdCtx)
}

// SimpleNodeDegree calculates the degree of a node by considering its
// initial neighbours. The degree is reduced by one when one of its neighbours
// are removed
type SimpleNodeDegree struct{}

// UpdateNodes return the neighbours of the eliminated node
func (s *SimpleNodeDegree) UpdateNodes(eliminated Node, adjList [][]int, ctx *AmdCtx) []int {
	return adjList[eliminated.index]
}

// Degree sets the degree equal to the number of neighbours of the node
func (s *SimpleNodeDegree) Degree(node int, adjList [][]int, ctx *AmdCtx) int {
	deg := 0
	for _, n := range adjList[node] {
		if !ctx.Eliminated[n] {
			deg++
		}
	}
	return deg
}

// OnNodeEliminated does nothing in the simplified version
func (s *SimpleNodeDegree) OnNodeEliminated(eliminated Node, adjList [][]int, ctx *AmdCtx) {}

// Function to find the minimum degree ordering
// This currently an experimental function that is under development
func ApproximateMinimumDegree(n int, adjList [][]int, degCalc NodeDegree) []int {
	if degCalc == nil {
		degCalc = &SimpleNodeDegree{}
	}
	ctx := NewAmdCtx(n)
	for i := range adjList {
		ctx.Degrees[i] = degCalc.Degree(i, adjList, &ctx)
	}

	for i, d := range ctx.Degrees {
		heap.Push(ctx.Heap, Node{degree: d, index: i})
	}

	for ctx.Heap.Len() > 0 {
		node := heap.Pop(ctx.Heap).(Node)
		if ctx.Eliminated[node.index] {
			continue
		}
		ctx.Ordering = append(ctx.Ordering, node.index)
		ctx.Eliminated[node.index] = true
		degCalc.OnNodeEliminated(node, adjList, &ctx)

		for _, neighbor := range degCalc.UpdateNodes(node, adjList, &ctx) {
			if ctx.Eliminated[neighbor] {
				continue
			}
			ctx.Degrees[neighbor] = degCalc.Degree(neighbor, adjList, &ctx)
			heap.Push(ctx.Heap, Node{degree: ctx.Degrees[neighbor], index: neighbor})
		}
	}
	return ctx.Ordering
}
