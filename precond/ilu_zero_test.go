package precond

import (
	"fmt"
	"math"
	"testing"

	"github.com/davidkleiven/goprecond/precond/precondtest"
	"golang.org/x/exp/rand"
	"gonum.org/v1/gonum/mat"
)

func TestKnownDecompositions(t *testing.T) {

	for testNum, test := range []struct {
		matrix *mat.Dense
		wantL  *mat.Dense
		wantU  *mat.Dense
	}{
		{
			matrix: mat.NewDense(2, 2, []float64{1.0, 2.0, 3.0, 4.0}),
			wantL:  mat.NewDense(2, 2, []float64{1.0, 0.0, 3.0, 1.0}),
			wantU:  mat.NewDense(2, 2, []float64{1.0, 2.0, 0.0, -2.0}),
		},
		{
			matrix: mat.NewDense(3, 3, []float64{1.0, 2.0, 3.0, 4.0, 5.0, 6.0, 7.0, 8.0, 9.0}),
			wantL:  mat.NewDense(3, 3, []float64{1.0, 0.0, 0.0, 4.0, 1.0, 0.0, 7.0, 2.0, 1.0}),
			wantU:  mat.NewDense(3, 3, []float64{1.0, 2.0, 3.0, 0.0, -3.0, -6.0, 0.0, 0.0, 0.0}),
		},
	} {
		zeroAware := &precondtest.DenseNonZeroDoer{Dense: test.matrix}
		lu := ILUZero(zeroAware)

		tol := 1e-6

		dim, _ := test.matrix.Dims()
		for i := 0; i < dim; i++ {
			for j := 0; j < dim; j++ {
				lGot := lu.lower.At(i, j)
				lWant := test.wantL.At(i, j)
				uGot := lu.upper.At(i, j)
				uWant := test.wantU.At(i, j)
				if math.Abs(lGot-lWant) > tol {
					t.Errorf("Test #%d: L (%d, %d): Wanted %f got %f", testNum, i, j, lWant, lGot)
					return
				}

				if math.Abs(uGot-uWant) > tol {
					t.Errorf("Test #%d: U (%d, %d): Wanted %f got %f", testNum, i, j, uWant, uGot)
					return
				}
			}
		}
	}

}

func Test_Solve2x2(t *testing.T) {
	matrix := precondtest.DenseNonZeroDoer{Dense: mat.NewDense(2, 2, []float64{1.0, 2.0, 3.0, 4.0})}
	lu := ILUZero(&matrix)
	rhs := mat.NewVecDense(2, []float64{1.0, 1.0})
	want := mat.NewVecDense(2, []float64{-1.0, 1.0})
	result := mat.NewVecDense(2, nil)
	lu.SolveVecTo(result, false, rhs)

	tol := 1e-6
	for i := 0; i < 2; i++ {
		if diff := math.Abs(want.AtVec(i) - result.AtVec(i)); diff > tol {
			t.Errorf("Wanted: %v\nGot: %v\n", want, result)
			return
		}
	}

}

type testCase struct {
	matrix *mat.Dense
	rhs    *mat.VecDense
}

func randomTestCase(dim int) testCase {
	rnd := rand.New(rand.NewSource(1))
	matrix := mat.NewDense(dim, dim, nil)
	rhs := mat.NewVecDense(dim, nil)
	for i := 0; i < dim; i++ {
		rhs.SetVec(i, rnd.NormFloat64())
		for j := 0; j < dim; j++ {
			matrix.Set(i, j, rnd.NormFloat64())
		}
	}
	return testCase{
		matrix: matrix,
		rhs:    rhs,
	}
}

func TestSelfConsistency(t *testing.T) {
	t.Parallel()
	for _, trans := range []bool{true, false} {
		for _, dim := range []int{5, 10, 15, 20, 25} {
			t.Run(fmt.Sprintf("trans %v dim %d", trans, dim), func(t *testing.T) {
				tc := randomTestCase(dim)
				zeroAware := &precondtest.DenseNonZeroDoer{Dense: mat.NewDense(dim, dim, nil)}
				zeroAware.CloneFrom(tc.matrix)

				splu := ILUZero(zeroAware)
				result := mat.NewVecDense(dim, nil)
				if err := splu.SolveVecTo(result, trans, tc.rhs); err != nil {
					t.Errorf("%v\n", err)
					return
				}

				want := mat.NewVecDense(dim, nil)
				if trans {
					want.MulVec(tc.matrix.T(), result)
				} else {
					want.MulVec(tc.matrix, result)
				}

				tol := 1e-6
				for i := 0; i < dim; i++ {
					if diff := math.Abs(want.AtVec(i) - tc.rhs.AtVec(i)); diff > tol {
						t.Errorf("Dim: %d, trans: %v, Wanted:\n%v\ngot\n%v\n", dim, trans, want, result)
						return
					}
				}
			})

		}
	}
}

func TestRandomMatricesAgainstGonum(t *testing.T) {
	t.Parallel()

	for _, dim := range []int{5, 10, 15, 20} {
		t.Run(fmt.Sprintf("%d", dim), func(t *testing.T) {
			tc := randomTestCase(dim)

			zeroAware := &precondtest.DenseNonZeroDoer{Dense: mat.NewDense(dim, dim, nil)}
			zeroAware.CloneFrom(tc.matrix)

			var lu mat.LU
			lu.Factorize(tc.matrix)

			resultGonum := mat.NewVecDense(dim, nil)
			result := mat.NewVecDense(dim, nil)

			if err := lu.SolveVecTo(resultGonum, false, tc.rhs); err != nil {
				t.Errorf("%v\n", err)
				return
			}

			splu := ILUZero(zeroAware)
			if err := splu.SolveVecTo(result, false, tc.rhs); err != nil {
				t.Errorf("%v\n", err)
				return
			}

			tol := 1e-6
			for i := 0; i < dim; i++ {
				if diff := math.Abs(resultGonum.AtVec(i) - result.AtVec(i)); diff > tol {
					t.Errorf("Too large difference (%f)\nGonum: %v\nUs: %v", diff, resultGonum, result)
					return
				}
			}
		})

	}
}
