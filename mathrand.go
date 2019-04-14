package randomkit

import (
	"math/rand"
)

// RKMathRandSource is a wrapper to adapt RKState to math/rand
type RKMathRandSource struct{ RKState *RKState }

// AsMathRandSource return RKState as a source suitable for rand.New
func (state *RKState) AsMathRandSource() *RKMathRandSource {
	return &RKMathRandSource{state}
}

// Seed proto for math/rand source
func (state *RKMathRandSource) Seed(seed int64) {
	state.RKState.Seed(uint64(seed))
}

// Int63 for math/rand source
func (state *RKMathRandSource) Int63() int64 {
	return state.RKState.Int63()
}

// Uint64 for math/rand source
func (state *RKMathRandSource) Uint64() uint64 {
	return state.RKState.Uint64()
}

// Clone allow duplicating source state
func (state *RKMathRandSource) Clone() rand.Source {
	newrkstate := *((state.RKState).Clone().(*RKState))
	return &RKMathRandSource{RKState: &newrkstate}
}
