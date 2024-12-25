package amd

type undirectedEdge struct {
	from int
	to   int
}

func newUndirectedEdge(a, b int) undirectedEdge {
	if a > b {
		return undirectedEdge{from: a, to: b}
	}
	return undirectedEdge{from: b, to: a}
}

// Implements the NodeDegree interface. This degree calculator
// takes into account the number of new edges formed when a node
// is eliminated
type FormedEdgeDegree struct {
	edges map[undirectedEdge]bool
}

func (f *FormedEdgeDegree) newEdges(node int, adjList [][]int, ctx *AmdCtx) []undirectedEdge {
	// If this node is remoed there will be a new edge between the neighbours
	// of this node and neighbours of the neighbours
	newEdges := []undirectedEdge{}
	for i, neighbour := range adjList[node] {

		if ctx.Eliminated[neighbour] {
			continue
		}

		for _, otherNeighbour := range adjList[node][i+1:] {
			edge := newUndirectedEdge(neighbour, otherNeighbour)
			if _, ok := f.edges[edge]; !ok && !ctx.Eliminated[otherNeighbour] {
				newEdges = append(newEdges, edge)
			}
		}
	}
	return newEdges
}

// Calculates the degree of node
func (f *FormedEdgeDegree) Degree(node int, adjList [][]int, ctx *AmdCtx) int {
	// Nearest neighbours
	nn := 0
	for _, n := range adjList[node] {
		if !ctx.Eliminated[n] {
			nn++
		}
	}
	newEdges := f.newEdges(node, adjList, ctx)
	return nn + len(newEdges)
}

// UpdateNodes return the list of nodes that should be considered for update
func (f *FormedEdgeDegree) UpdateNodes(eliminated Node, adjList [][]int, ctx *AmdCtx) []int {
	return adjList[eliminated.index]
}

// OnNodeEliminated updates the adjacency lists and the current new edges when a node
// is removed
func (f *FormedEdgeDegree) OnNodeEliminated(eliminated Node, adjList [][]int, ctx *AmdCtx) {
	newEdges := f.newEdges(eliminated.index, adjList, ctx)

	// Update known edges the first time this method is called when a new node is eliminated
	for _, edge := range newEdges {
		f.edges[edge] = true
		adjList[edge.from] = append(adjList[edge.from], edge.to)
		adjList[edge.to] = append(adjList[edge.to], edge.from)
	}
}

func NewFormedEdgeDegree(adjList [][]int) *FormedEdgeDegree {
	edges := make(map[undirectedEdge]bool)

	for i, neighbours := range adjList {
		for _, neighbour := range neighbours {
			edges[newUndirectedEdge(i, neighbour)] = true
		}
	}
	return &FormedEdgeDegree{
		edges: edges,
	}
}
