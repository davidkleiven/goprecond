package precond

import (
	"math"
	"math/rand"
	"slices"
	"testing"
	"testing/quick"

	"github.com/davidkleiven/goprecond/precond/precondtest"
	"github.com/davidkleiven/goprecond/precond/property"
	"gonum.org/v1/gonum/mat"
)

func linspaceMatrix(n, m int) *mat.Dense {
	data := make([]float64, n*m)
	for i := range data {
		data[i] = float64(i)
	}
	return mat.NewDense(n, m, data)
}

func equal(A mat.Matrix, B mat.Matrix, tol float64) bool {
	rA, cA := A.Dims()
	rB, cB := B.Dims()
	if rA != rB || cA != cB {
		return false
	}

	for i := 0; i < rA; i++ {
		for j := 0; j < rB; j++ {
			if math.Abs(A.At(i, j)-B.At(i, j)) > tol {
				return false
			}
		}
	}
	return true
}

func TestIdentityPivot(t *testing.T) {
	matrix := linspaceMatrix(3, 3)
	pivot := NoPivot(3)

	result := mat.NewDense(3, 3, nil)
	result.Mul(&pivot, matrix)

	if !equal(matrix, result, 1e-6) {
		t.Errorf("Wanted\n%v\ngot\n%v\n", matrix, result)
	}
}

func TestPivoting(t *testing.T) {
	for i, test := range []struct {
		pivot   []int
		matrix  *mat.Dense
		want    *mat.Dense
		colSwap bool
	}{
		{
			pivot:   []int{1, 0},
			matrix:  linspaceMatrix(2, 2),
			want:    mat.NewDense(2, 2, []float64{2.0, 3.0, 0.0, 1.0}),
			colSwap: false,
		},
		{
			pivot:   []int{1, 0, 2},
			matrix:  linspaceMatrix(3, 3),
			want:    mat.NewDense(3, 3, []float64{3.0, 4.0, 5.0, 0.0, 1.0, 2.0, 6.0, 7.0, 8.0}),
			colSwap: false,
		},
		{
			pivot:   []int{0, 2, 1},
			matrix:  linspaceMatrix(3, 3),
			want:    mat.NewDense(3, 3, []float64{0.0, 1.0, 2.0, 6.0, 7.0, 8.0, 3.0, 4.0, 5.0}),
			colSwap: false,
		},
		{
			pivot:   []int{1, 0},
			matrix:  linspaceMatrix(2, 2),
			want:    mat.NewDense(2, 2, []float64{1.0, 0.0, 3.0, 2.0}),
			colSwap: true,
		},
		{
			pivot:   []int{1, 0, 2},
			matrix:  linspaceMatrix(3, 3),
			want:    mat.NewDense(3, 3, []float64{1.0, 0.0, 2.0, 4.0, 3.0, 5.0, 7.0, 6.0, 8.0}),
			colSwap: true,
		},
		{
			pivot:   []int{2, 1, 0},
			matrix:  linspaceMatrix(3, 3),
			want:    mat.NewDense(3, 3, []float64{2.0, 1.0, 0.0, 5.0, 4.0, 3.0, 8.0, 7.0, 6.0}),
			colSwap: true,
		},
	} {
		n, m := test.matrix.Dims()
		pivot := Pivot{test.pivot}

		result := mat.NewDense(n, m, nil)

		if test.colSwap {
			result.Mul(test.matrix, pivot.T())
		} else {
			result.Mul(&pivot, test.matrix)
		}

		if !equal(result, test.want, 1e-6) {
			t.Errorf("test #%d: wanted\n%v\ngot\n%v\n", i, test.want, result)
		}
	}
}

func TestPartialPivot(t *testing.T) {
	for _, test := range []struct {
		matrix precondtest.DenseNonZeroDoer
		want   []int
		desc   string
	}{
		{
			matrix: precondtest.DenseNonZeroDoer{Dense: mat.NewDense(2, 2, []float64{1.0, 2.0, 3.0, 4.0})},
			want:   []int{1, 0},
			desc:   "Simple 2x2 matrix",
		},
		{
			matrix: precondtest.DenseNonZeroDoer{Dense: mat.NewDense(2, 2, []float64{3.0, 4.0, 1.0, 2.0})},
			want:   []int{0, 1},
			desc:   "Simple 2x2 matrix already ordered",
		},
		{
			matrix: precondtest.DenseNonZeroDoer{Dense: mat.NewDense(2, 2, []float64{1.0, 2.0, -3.0, -4.0})},
			want:   []int{1, 0},
			desc:   "Simple 2x2 matrix with negative numbers",
		},
	} {
		nrows, _ := test.matrix.Dims()
		pivot := PartialPivotMatrix(&test.matrix, nrows)

		if slices.Compare(pivot.Pivots, test.want) != 0 {
			t.Errorf("Wanted\n%v\ngot%v\n", test.want, pivot.Pivots)
		}
	}
}

func TestPivotingDiagonalIsDescending(t *testing.T) {
	config := quick.Config{
		Rand: rand.New(rand.NewSource(0)),
	}

	diagonalIsDesc := func(gen property.DenseSquareMatrixGenerator) bool {
		nrows, ncols := gen.Matrix.Dims()
		denseDoer := precondtest.DenseNonZeroDoer{gen.Matrix}
		pivot := PartialPivotMatrix(&denseDoer, nrows)

		// Apply pivot and confirm that diagonal is descending in absolute value
		result := mat.NewDense(nrows, ncols, nil)
		result.Mul(&pivot, gen.Matrix)
		for i := 0; i < nrows-1; i++ {
			diag := math.Abs(result.At(i, i))
			for j := i + 1; j < nrows; j++ {
				if v := math.Abs(result.At(j, i)); v > diag {
					return false
				}
			}
		}
		return true
	}

	if err := quick.Check(diagonalIsDesc, &config); err != nil {
		t.Error(err)
	}
}
