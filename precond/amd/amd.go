package amd

import (
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
}

func (ctx *AmdCtx) MinimumActiveDegree() Node {
	isFirst := true
	minimumDeg := 0
	minumNode := 0
	for i := range ctx.Degrees {
		if !ctx.Eliminated[i] {
			if ctx.Degrees[i] < minimumDeg || isFirst {
				isFirst = false
				minimumDeg = ctx.Degrees[i]
				minumNode = i
			}
		}
	}

	if isFirst {
		panic("No active elements")
	}
	return Node{degree: minimumDeg, index: minumNode}
}

func NewAmdCtx(n int) AmdCtx {
	return AmdCtx{
		Degrees:    make([]int, n),
		Eliminated: make([]bool, n),
		Ordering:   make([]int, 0, n),
	}
}

type NodeDegree interface {
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

	for i := 0; i < n; i++ {
		node := ctx.MinimumActiveDegree()
		ctx.Ordering = append(ctx.Ordering, node.index)
		ctx.Eliminated[node.index] = true
		degCalc.OnNodeEliminated(node, adjList, &ctx)

		for _, neighbor := range adjList[node.index] {
			if ctx.Eliminated[neighbor] {
				continue
			}
			ctx.Degrees[neighbor] = degCalc.Degree(neighbor, adjList, &ctx)
		}
	}
	return ctx.Ordering
}
