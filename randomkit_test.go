package randomkit

import (
	"fmt"
	"math"
	mathRand "math/rand"
	"testing"

	expRand "golang.org/x/exp/rand"
)

var (
	_ expRand.Source    = &RKState{}
	_ mathRand.Source   = (&RKState{}).AsMathRandSource()
	_ mathRand.Source64 = (&RKState{}).AsMathRandSource()
)

func ExampleRKState() {
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
	// Output:
}

func TestSeed(t *testing.T) {
	src := NewRandomkitSource(7)
	t.Run("Float64", func(t *testing.T) {
		if math.Abs(src.Float64()-0.07630829) > 1e-8 {
			t.Fail()
		}
	})
	msrc := src.AsMathRandSource()
	t.Run("math/rand/Source/Int63", func(t *testing.T) {
		msrc.Seed(7)
		ex, ac := int64(1407639518939636932), msrc.Int63()
		if ac != ex {
			t.Errorf("expected %d got %d", ex, ac)
		}
	})
	t.Run("math/rand/Source/Uint64", func(t *testing.T) {
		msrc.Seed(7)
		ex, ac := uint64(1407639518939636932), msrc.Uint64()
		if ac != ex {
			t.Errorf("expected %d got %d", ex, ac)
		}
	})
	t.Run("math/rand/Source/Clone", func(t *testing.T) {
		msrc.Seed(7)
		msrc2 := msrc.Clone()
		ex := int64(1407639518939636932)
		if msrc2.Int63() != ex || msrc.Clone().Int63() != ex {
			t.Errorf("expected %d", ex)
		}
	})
	t.Run("Uint64s", func(t *testing.T) {
		ex := "[47 68 25 67 83]"
		src.Seed(7)
		ac := make([]uint64, 5, 5)
		src.Uint64s(0, 100, ac)
		if fmt.Sprintf("%d", ac) != ex {
			t.Errorf("excepted %s got %s", ex, fmt.Sprintf("%d", ac))
		}
	})
}
