package precond

import (
	"math"

	"testing"

	"github.com/davidkleiven/goprecond/precond/precondtest"
	"gonum.org/v1/gonum/mat"
)

func randomSymmetricTestCase(N int) testCase {
	c := randomTestCase(N)
	c.matrix.Add(c.matrix, c.matrix.T())

	for i := 0; i < N; i++ {
		c.matrix.Set(i, i, math.Abs(c.matrix.At(i, i)))
	}
	return c
}

func TestICholSolution(t *testing.T) {
	t.Parallel()

	for testNum, n := range []int{5, 10, 15, 20} {
		c := randomSymmetricTestCase(n)
		ichol := IChol(&precondtest.DenseNonZeroDoer{Dense: c.matrix})
		icholResultvec := mat.NewVecDense(n, nil)
		ichoTResultVec := mat.NewVecDense(n, nil)
		dotProduct := mat.NewVecDense(n, nil)
		ichol.SolveVecTo(ichoTResultVec, false, c.rhs)
		ichol.SolveVecTo(ichoTResultVec, false, c.rhs)
		dotProduct.MulVec(c.matrix, icholResultvec)

		tol := 1e-6
		for i := 0; i < n; i++ {
			if math.Abs(ichoTResultVec.AtVec(i)-icholResultvec.AtVec(i)) > tol {
				t.Errorf("Test #%d: transposed result\n%v\nuntransposed result\n%v\n", testNum, icholResultvec, ichoTResultVec)
				return
			}

			if math.Abs(icholResultvec.AtVec(i)-dotProduct.AtVec(i)) > tol {
				t.Errorf("Test #%d: wanted\n%v\ngot\n%v\n", testNum, c.rhs, dotProduct)
				return
			}
		}

	}
}
