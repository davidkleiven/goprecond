package amd

// Weighted Enode is a quotient graph method that does not traverse the graph
// when calculating the degree of node. Instead each node gets a weight. A
// regular node gets a weight 1. When a node is eliminated (e.g. e-node) the
// weight is set to the sum of the weights of its neighbours
type WeightedEnode struct {
	weights []int
}

func NewWeightedEnode(n int) *WeightedEnode {
	weights := make([]int, n)
	for i := range weights {
		weights[i] = 1
	}
	return &WeightedEnode{
		weights: weights,
	}
}

// Degree calculates the degree of node. The degree is equal to the sum of weights
// ofs its neighbours
func (w *WeightedEnode) Degree(node int, adjList [][]int, ctx *AmdCtx) int {
	deg := 0
	for _, n := range adjList[node] {
		deg += w.weights[n]
	}
	return deg
}

// OnNodeEliminated updates the weight when a node is eliminated. The weight of the eliminated node
// is set equal to its degree prior to removal minus the weight of itself
func (w *WeightedEnode) OnNodeEliminated(eliminated int, adjList [][]int, ctx *AmdCtx) {
	w.weights[eliminated] = w.Degree(eliminated, adjList, ctx) - w.weights[eliminated]
}
