package amd

// QuotientGraphExactDegreeCalculator calculates the exact degree of a quotient graph
// this is done by recursively following the nodes of neighouring node which makes
// this method very expensive for large graphs
type QuotientGraphExactDegreeCalculator struct{}

func (q *QuotientGraphExactDegreeCalculator) Degree(node int, adjList [][]int, ctx *AmdCtx) int {
	deg := 0
	visited := make([]bool, len(adjList))
	visited[node] = true
	tree := make([]int, len(adjList))
	depth := 0
	tree[depth] = node
	for depth >= 0 {
		current := nextUnvisitedNeighbour(visited, adjList[tree[depth]])

		if current == -1 {
			// All neighbours are visited
			depth -= 1
			continue
		}

		for current != -1 && ctx.Eliminated[current] {
			depth += 1
			tree[depth] = current
			visited[current] = true
			current = nextUnvisitedNeighbour(visited, adjList[tree[depth]])
		}

		if current == -1 {
			depth -= 1
			continue
		}
		deg += 1
		visited[current] = true
	}
	return deg
}

func (q *QuotientGraphExactDegreeCalculator) OnNodeEliminated(eliminated Node, adjList [][]int, ctx *AmdCtx) {
}

func nextUnvisitedNeighbour(visited []bool, neighbours []int) int {
	for _, n := range neighbours {
		if !visited[n] {
			return n
		}
	}
	return -1
}
