<img src="logo.svg" width="400" alt="Logo">

Package implementing common spare preconditioners

* Incomplete LU
* Incomplete Cholesky

## Installation

```bash
go get -u github.com/davidkleiven/goprecond
```

## Example

Example showing the effect of using the incomplete LU preconditioner

```go
package precond_test

import (
	"fmt"

	"golang.org/x/exp/rand"
	"golang.org/x/exp/slices"

	"github.com/davidkleiven/goprecond/precond"
	"github.com/james-bowman/sparse"
	"gonum.org/v1/exp/linsolve"
	"gonum.org/v1/gonum/mat"
)

type linearSystem struct {
	matrix precond.CSRMulVecToer
	rhs    *mat.VecDense
}

func randomMatrix(N int, fractionNonZero float64) linearSystem {
	matrix := sparse.NewDOK(N, N)

	rnd := rand.New(rand.NewSource(1))
	rhs := mat.NewVecDense(N, nil)

	for i := 0; i < N; i++ {
		matrix.Set(i, i, 1.0+rnd.NormFloat64())
		rhs.SetVec(i, rand.NormFloat64())
	}

	numNonZeroToDraw := slices.Min([]int{int(float64(N*N)*fractionNonZero) - N, 0})

	for i := 0; i < numNonZeroToDraw; i++ {
		row := rnd.Int() % N
		col := rnd.Int() % N

		if row == col {
			col = (col + 1) % N
		}

		matrix.Set(row, col, rnd.NormFloat64())
	}

	return linearSystem{
		matrix: precond.NewCSRMulVecToer(matrix.ToCSR()),
		rhs:    rhs,
	}
}

func ExampleILUZero() {
	N := 1000
	system := randomMatrix(N, 0.1)
	iluPrecon := precond.ILUZero(system.matrix.Matrix)

	for _, settings := range []*linsolve.Settings{nil, {PreconSolve: iluPrecon.SolveVecTo}} {

		result, err := linsolve.Iterative(&system.matrix, system.rhs, &linsolve.GMRES{}, settings)

		if err != nil {
			fmt.Println(err)
			return
		}

		fmt.Printf("# MulVec: %v, # PreconSolve: %v\n", result.Stats.MulVec, result.Stats.PreconSolve)
	}

	// Output:
	// # MulVec: 686, # PreconSolve: 687
	// # MulVec: 1, # PreconSolve: 2
}
```


## Approximate Minimum Degree Ordering

Approximate Minimum Degree ordering can be used to minimize the fill-ins in a full Cholesky decomposition.
This library is primarily conserned with the incomplete factorizations which does not have fill-ins anyways.
However, the incomplete decomposition is closer to the exact decomposition when the fill-ins are minimzed.
The central part of the Approximate Minimum Degree method is how the degrees of the nodes are calcaulted.
This library has an interface that allows users to pass different algorithms for evaluating the degree of the nodes.

<p align="center">
	<img src=./precond/amd/doc/orig.svg width=200px/>
	<img src=./precond/amd/doc/qGraphExact.svg width=200px/>
	<img src=./precond/amd/doc/wEnode.svg width=200px/>
</p>

The leftmost figure shows the initial matrix which is a tridiagonal matrix where the indices has been randomly shuffled.
The center image shows the result after re-ordering according to the `QuotientGraphExactDegreeCalculator` and the right image shows the result of `WeightedEnode` degree calculator. Both the the latter calculators are capable of almost recovering the tridiagonal structure of the matrix.
It should be noted the `QuotientGraphExactDegreeCalculator` is a quite expensive, but accurate, degree calculator and is most likely not practical for large matrices.