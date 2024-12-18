<img src="logo.svg" width="400" alt="Logo">

Package implementing common spare preconditioners

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
