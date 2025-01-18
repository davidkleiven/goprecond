package amd

import (
	"gonum.org/v1/gonum/mat"
)

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

func (ctx *AmdCtx) MinimumActiveDegree() int {
	isFirst := true
	minimumDeg := 0
	minimumNode := 0
	for i := range ctx.Degrees {
		if !ctx.Eliminated[i] {
			if ctx.Degrees[i] < minimumDeg || isFirst {
				isFirst = false
				minimumDeg = ctx.Degrees[i]
				minimumNode = i
			}
		}
	}

	if isFirst {
		panic("No active elements")
	}
	return minimumNode
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
	OnNodeEliminated(eliminated int, adjList [][]int, ctx *AmdCtx)
}

// Function to find the minimum degree ordering
// This currently an experimental function that is under development
func ApproximateMinimumDegree(n int, adjList [][]int, degCalc NodeDegree) []int {
	if degCalc == nil {
		degCalc = NewWeightedEnode(n)
	}
	ctx := NewAmdCtx(n)
	for i := range adjList {
		ctx.Degrees[i] = degCalc.Degree(i, adjList, &ctx)
	}

	for i := 0; i < n; i++ {
		node := ctx.MinimumActiveDegree()
		ctx.Ordering = append(ctx.Ordering, node)
		ctx.Eliminated[node] = true
		degCalc.OnNodeEliminated(node, adjList, &ctx)

		for _, neighbor := range adjList[node] {
			if ctx.Eliminated[neighbor] {
				continue
			}
			ctx.Degrees[neighbor] = degCalc.Degree(neighbor, adjList, &ctx)
		}
	}
	return ctx.Ordering
}
