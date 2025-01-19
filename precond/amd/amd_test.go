package amd

import (
	"fmt"
	"math"
	"slices"
	"testing"
	"testing/quick"

	"math/rand"

	"github.com/davidkleiven/goprecond/precond"
	"github.com/davidkleiven/goprecond/precond/precondtest"
	"github.com/davidkleiven/goprecond/precond/property"
	"gonum.org/v1/gonum/mat"
	"pgregory.net/rapid"
)

func TestAdjacencyList(t *testing.T) {
	for _, test := range []struct {
		mat  *mat.Dense
		want [][]int
		desc string
	}{
		{
			mat:  mat.NewDense(2, 2, []float64{1.0, 0.0, 0.0, 1.0}),
			want: [][]int{},
			desc: "Diagonal matrix should result in empty adjecancy list",
		},
		{
			mat:  mat.NewDense(2, 2, []float64{0.0, 1.0, 0.0, 1.0}),
			want: [][]int{{1}},
			desc: "One connecttion between 0, 1",
		},
		{
			mat:  mat.NewDense(2, 2, []float64{1.0, 0.0, 1.0, 1.0}),
			want: [][]int{{}, {0}},
			desc: "One connecttion between 0, 1, but only from the last row",
		},
		{
			mat:  mat.NewDense(2, 2, []float64{0.0, 1.0, 1.0, 0.0}),
			want: [][]int{{1}, {0}},
			desc: "One connecttion between 0, 1, from both rows",
		},
	} {
		result := AdjacencyList(&precondtest.DenseNonZeroDoer{Dense: test.mat})

		if len(result) != len(test.want) {
			t.Errorf("Test %s: Wanted\n%v\ngot\n%v\n", test.desc, test.want, result)
			break
		}

		for i := range result {
			if !slices.Equal(result[i], test.want[i]) {
				t.Errorf("Test %s: Wanted\n%v\ngot\n%v\n", test.desc, test.want, result)
			}
		}
	}
}

type pair struct {
	smallest, largest int
}

func TestAdjacencyListProperties(t *testing.T) {
	config := quick.Config{
		Rand: rand.New(rand.NewSource(0)),
	}

	pairsAreUnique := func(gen property.DenseSquareMatrixGenerator) bool {
		result := AdjacencyList(&precondtest.DenseNonZeroDoer{Dense: gen.Matrix})
		pairs := make(map[pair]bool)

		// Pairs occur symmetrically
		numConnections := 0
		for i, neighbours := range result {
			for _, neighbour := range neighbours {
				if i == neighbour {
					return false
				}

				pair := pair{smallest: i, largest: neighbour}

				if pair.smallest > pair.largest {
					pair.smallest, pair.largest = pair.largest, pair.smallest
				}
				pairs[pair] = true
				numConnections += 1
			}
		}

		return len(pairs) == numConnections/2
	}

	if err := quick.Check(pairsAreUnique, &config); err != nil {
		t.Error(err)
	}
}

func TestAmdOrdering(t *testing.T) {
	matrix := mat.NewDense(4, 4, []float64{
		1.0, 0.0, 1.0, 1.0, // Degree: 2
		0.0, 2.0, 0.0, 1.0, // Degree: 1
		1.0, 0.0, 1.0, 0.0, // Degree: 1
		1.0, 1.0, 0.0, 1.0, // Degree: 2
	})

	adjList := AdjacencyList(&precondtest.DenseNonZeroDoer{Dense: matrix})
	order := ApproximateMinimumDegree(4, adjList, nil)
	want := []int{1, 2, 0, 3}

	if !slices.Equal(want, order) {
		t.Errorf("Wanted\n%v\ngot\n%v\n", want, order)
	}
}

func nnz(mat mat.Matrix) int {
	n := 0
	tol := 1e-8
	r, c := mat.Dims()
	for i := 0; i < r; i++ {
		for j := 0; j < c; j++ {
			if math.Abs(mat.At(i, j)) > tol {
				n++
			}
		}
	}
	return n
}

type SymmetricNonZeroDoer struct {
	matrix mat.Symmetric
}

func (s *SymmetricNonZeroDoer) DoNonZero(fn func(i, j int, v float64)) {
	n := s.matrix.SymmetricDim()
	for i := 0; i < n; i++ {
		for j := 0; j < n; j++ {
			if v := s.matrix.At(i, j); v != 0.0 {
				fn(i, j, v)
			}
		}
	}
}

func fillInIsReduced(t *rapid.T, matrix mat.Symmetric, calcName string) {
	r := matrix.SymmetricDim()
	var chol mat.Cholesky
	if ok := chol.Factorize(matrix); !ok {
		panic("Matrix is not positive definite")
	}

	initialL := mat.NewTriDense(r, mat.Lower, nil)
	chol.LTo(initialL)
	initNonZero := nnz(initialL)

	adj := AdjacencyList(&SymmetricNonZeroDoer{matrix})
	var calc NodeDegree

	if calcName == exact {
		calc = &QuotientGraphExactDegreeCalculator{}
	} else if calcName == weigthedEnode {
		calc = NewWeightedEnode(r)
	} else {
		panic(fmt.Sprintf("Unknown calc %s", calcName))
	}
	order := ApproximateMinimumDegree(r, adj, calc)

	pivot := precond.Pivot{Pivots: order}

	// Pivoted matrix
	var colPivoted mat.Dense
	colPivoted.Mul(matrix, pivot.T())

	var fullyPivoted mat.Dense
	fullyPivoted.Mul(&pivot, &colPivoted)

	fullyPivotedSym := mat.NewSymDense(r, fullyPivoted.RawMatrix().Data)

	if ok := chol.Factorize(fullyPivotedSym); !ok {
		panic("Matrix is not positive definite")
	}

	var finalL mat.TriDense
	chol.LTo(&finalL)
	finalNonZero := nnz(&finalL)
	if finalNonZero > initNonZero {
		t.Fatalf("Fill-in should be reduced when applying AMD. nnz before: %d nnz after %d\n", initNonZero, finalNonZero)
	}
}

const (
	exact         = "exact"
	weigthedEnode = "weightedEnode"
)

func TestReducedFillIn(t *testing.T) {
	calcName := []string{exact, weigthedEnode}
	rapid.Check(t, func(t *rapid.T) {
		calcName := calcName[rapid.IntRange(0, 1).Draw(t, "calc-name-decisor")]
		matrix := property.SparseSymmetricMatrix(t, 10, 50)
		fillInIsReduced(t, matrix, calcName)
	})
}
