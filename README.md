# randomkit

generate random sequences like numpy ones

[![Build Status](https://travis-ci.org/pa-m/randomkit.svg?branch=master)](https://travis-ci.org/pa-m/randomkit)
[![Code Coverage](https://codecov.io/gh/pa-m/randomkit/branch/master/graph/badge.svg)](https://codecov.io/gh/pa-m/randomkit)
[![Go Report Card](https://goreportcard.com/badge/github.com/pa-m/randomkit)](https://goreportcard.com/report/github.com/pa-m/randomkit)
[![GoDoc](https://godoc.org/github.com/pa-m/randomkit?status.svg)](https://godoc.org/github.com/pa-m/randomkit)


### Example

```go
	var (
		state       *RKState
		a, expected []float64
		tol         = 1e-8
		ok          bool
	)
	// np.random.seed(7)
	// np.random.sample(5)
	state = NewRandomkitSource(7)
	a = make([]float64, 5)
	for i := range a {
		a[i] = state.Float64()
	}

	expected = []float64{0.07630829, 0.77991879, 0.43840923, 0.72346518, 0.97798951}
	ok = true
	for i := range a {
		ok = ok && math.Abs(expected[i]-a[i]) < tol
	}
	if !ok {
		fmt.Printf("expected %g got %g", expected, a)
	}

	expected = []float64{0.07630829, 0.77991879, 0.43840923, 0.72346518, 0.97798951}
	ok = true
	for i := range a {
		ok = ok && math.Abs(expected[i]-a[i]) < tol
	}
	if !ok {
		fmt.Printf("expected %g got %g", expected, a)
	}

	// test normal dist
	// np.random.seed(7)
	// np.random.standard_normal(5)
	state.Seed(7)
	for i := range a {
		a[i] = state.NormFloat64()
	}
	expected = []float64{1.6905257, -0.46593737, 0.03282016, 0.40751628, -0.78892303}
	ok = true
	for i := range a {
		ok = ok && math.Abs(expected[i]-a[i]) < tol
	}
	if !ok {

		fmt.Printf("expected %g got %g", expected, a)

	}

	// duplicate state have same future
	stateCopy := state.Clone()
	if state.Uint64() != stateCopy.Uint64() {
		fmt.Println("clone failure")
	}
```

### warning

RKState is a suitable Source for golang.org/x/exp/rand but using rand.Float64 don't provide numpy-like sequences because rand.Float64 operations differ form RKState.Float64

The only way to get numpy-like sequences is to use RKState.Float64 or RKState.NormFloat64 directly

