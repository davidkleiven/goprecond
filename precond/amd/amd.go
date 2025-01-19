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

		if _, ok := adj[j]; !ok {
			adj[j] = make(map[int]bool)
		}
		adj[j][i] = true
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

func isAdjacencyList(adjList [][]int) bool {
	maxNodeNum := 0
	for _, neighbours := range adjList {
		if len(neighbours) == 0 {
			return false
		}
		for _, neighbour := range neighbours {
			if neighbour > maxNodeNum {
				maxNodeNum = neighbour
			}
		}
	}
	return maxNodeNum == len(adjList)-1
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
// It takes an adjecancy list that describes the neighbours of each node
// e.g. [][]int{{1}, {0, 2}, {1}} describes the a graph where node 0
// is neighbour to node 1. Node one is neighbour to 0 and 2 and node
// 2 is neighbour to 1: 0 ---- 1 ---- 2
//
// degCalc implements the details of how the degree of each node is calculated
// if nil, the WeightedEnode is used
//
// The method panics if the adjList is invalid. The adjacancy list is invalid
// if either some node have zero neighbours or some nodes does not have a
// neighbour list
func ApproximateMinimumDegree(adjList [][]int, degCalc NodeDegree) []int {
	if !isAdjacencyList(adjList) {
		panic("The provided adjecancy list appears to not be an adjacency list. All nodes must have a neighbour list")
	}
	n := len(adjList)
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
