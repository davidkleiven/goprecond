package precond

import (
	"math"
	"testing"

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
