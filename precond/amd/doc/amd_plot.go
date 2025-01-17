package main

import (
	"image/color"
	"math/rand"

	"github.com/davidkleiven/goprecond/precond/amd"
	"github.com/james-bowman/sparse"
	"gonum.org/v1/gonum/mat"
	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/vg"
	"gonum.org/v1/plot/vg/draw"
)

const size = 100

func randomSymMatrix() *sparse.DOK {
	rnd := rand.New(rand.NewSource(0))

	// Node numbers
	nodeNums := make([]int, size)
	for i := 0; i < size; i++ {
		nodeNums[i] = i
	}
	rnd.Shuffle(size, func(i, j int) { nodeNums[i], nodeNums[j] = nodeNums[j], nodeNums[i] })

	matrix := sparse.NewDOK(size, size)
	for i := 0; i < size; i++ {
		diag := nodeNums[i]

		if i < size-1 {
			right := nodeNums[i+1]
			matrix.Set(diag, right, -1.0)
		}

		if i > 0 {
			left := nodeNums[i-1]
			matrix.Set(diag, left, -1.0)
		}

		matrix.Set(diag, diag, 2.0)
	}
	return matrix
}

func plotNonZero(mat mat.NonZeroDoer) *plot.Plot {
	p := plot.New()

	pts := make(plotter.XYs, 0)
	mat.DoNonZero(func(i, j int, v float64) {
		pts = append(pts, plotter.XY{X: float64(j), Y: -float64(i)})
	})

	scatter, err := plotter.NewScatter(pts)
	scatter.GlyphStyle.Color = color.Gray{0}
	scatter.GlyphStyle.Shape = draw.BoxGlyph{}

	if err != nil {
		panic(err)
	}
	p.Add(scatter)
	return p

}

func pivotDOK(orig *sparse.DOK, order []int) *sparse.DOK {
	r, c := orig.Dims()
	pivoted := sparse.NewDOK(r, c)
	nodeMap := make(map[int]int)
	for i, v := range order {
		nodeMap[v] = i
	}
	orig.DoNonZero(func(i, j int, v float64) {
		pivoted.Set(nodeMap[i], nodeMap[j], v)
	})
	return pivoted
}

func main() {
	matrix := randomSymMatrix()
	adjList := amd.AdjacencyList(matrix)
	orig := plotNonZero(matrix)

	// Exact order using method that considers new edges formed
	qGraphOrder := amd.ApproximateMinimumDegree(size, adjList, &amd.QuotientGraphExactDegreeCalculator{})
	qGraphOrderPivoted := pivotDOK(matrix, qGraphOrder)
	qGraphOrderPivotPlot := plotNonZero(qGraphOrderPivoted)

	// Save plots
	orig.Save(4*vg.Inch, 4*vg.Inch, "orig.svg")
	qGraphOrderPivotPlot.Save(4*vg.Inch, 4*vg.Inch, "qGraphExact.svg")
}
