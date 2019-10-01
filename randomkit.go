package randomkit

import (
	// "encoding/binary"
	"errors"
	"math"
	"sort"

	"golang.org/x/exp/rand"
	// "os"
	// "time"
)

// port some functions of https://github.com/numpy/numpy/blob/master/numpy/random/mtrand/randomkit.c

const rkStateLen = 624

// RKState is a random source based on numpy's randomkit
// it implements "golang.org/x/exp/rand".Source
// and can be used as "math/rand".Source via AsMathRandSource
// but RKState Float64 and NormFloat64 methods must be used directly to reproduce original randomkit numpy sequences.
type RKState struct {
	key           [rkStateLen]uint64
	pos           int
	hasGauss      bool
	gauss         float64
	hasBinomial   bool
	psave         float64
	nsave         int32
	r, q, fm      float64
	m             int32
	p1            float64
	xm, xl, xr, c float64
	laml, lamr    float64
	p2, p3, p4    float64
}

var errNoDev = errors.New("random device unavailable")

// Seed initializes state with Knuth's PRNG
func (state *RKState) Seed(seed uint64) {
	seed &= 0xffffffff

	/* Knuth's PRNG as used in the Mersenne Twister reference implementation */
	for pos := uint64(0); pos < rkStateLen; pos++ {
		state.key[pos] = seed
		seed = (1812433253*(seed^(seed>>30)) + pos + 1) & 0xffffffff
	}
	state.pos = rkStateLen
	state.gauss = 0
	state.hasGauss = false
	state.hasBinomial = false
}

/* Thomas Wang 32 bits integer hash function */
/*
func rkHash(key uint64) uint64 {
	key += ^(key << 15)
	key ^= (key >> 10)
	key += (key << 3)
	key ^= (key >> 6)
	key += ^(key << 11)
	key ^= (key >> 16)
	return key
}

func rkDevfill(buffer interface{}, size int, strong bool) error {
	var rfile *os.File
	var err error
	if strong {
		rfile, err = os.Open("/dev/random")
	} else {
		rfile, err = os.Open("/dev/urandom")
	}
	if err != nil {
		return errNoDev
	}
	err = binary.Read(rfile, binary.LittleEndian, buffer)
	if err == nil {
		return nil
	}
	return errNoDev

}
*/
/*
func rkRandomseed(state *RKState) error {
	if rkDevfill(state.key, rkStateLen*4, false) == nil {
		// ensures non-zero key
		state.key[0] |= 0x80000000
		state.pos = rkStateLen
		state.gauss = 0
		state.hasGauss = false
		state.hasBinomial = false

		for i := range state.key {
			state.key[i] &= 0xffffffff
		}
		return nil

	}
	state.Seed(uint64(time.Now().UnixNano()))
	return errNoDev
}
*/
/* Magic Mersenne Twister constants */
const (
	N         = 624
	M         = 397
	MatrixA   = 0x9908b0df
	UpperMask = 0x80000000
	LowerMask = 0x7fffffff
)

// Uint32 generator
//   Slightly optimised reference implementation of the Mersenne Twister
//   Note that regardless of the precision of long, only 32 bit random
//   integers are produced
func (state *RKState) Uint32() uint32 {
	var y uint32
	if state.pos == rkStateLen {
		var i int
		for i = 0; i < N-M; i++ {
			y = uint32(state.key[i]&UpperMask) | uint32(state.key[i+1]&LowerMask)
			state.key[i] = uint64(uint32(state.key[i+M]) ^ (y >> 1) ^ (-(y & 1) & MatrixA))
		}
		for ; i < N-1; i++ {
			y = uint32(state.key[i]&UpperMask) | uint32(state.key[i+1]&LowerMask)
			state.key[i] = uint64(uint32(state.key[i+(M-N)]) ^ (y >> 1) ^ (-(y & 1) & MatrixA))
		}
		y = (uint32(state.key[N-1]) & UpperMask) | (uint32(state.key[0]) & LowerMask)
		state.key[N-1] = state.key[M-1] ^ uint64(y>>1) ^ uint64(-(y&1)&MatrixA)

		state.pos = 0
	}
	y = uint32(state.key[state.pos])
	state.pos++

	/* Tempering */
	y ^= (y >> 11)
	y ^= (y << 7) & 0x9d2c5680
	y ^= (y << 15) & 0xefc60000
	y ^= (y >> 18)

	return y
}

// Uint64 Returns an unsigned 64 bit random integer.
func (state *RKState) Uint64() uint64 {
	upper := uint64(state.Uint32()) << 32
	lower := uint64(state.Uint32())
	return upper | lower
}

// Uint64s Fills an array with random uint64 between off and off + rng inclusive. The numbers wrap if rng is sufficiently large.
func (state *RKState) Uint64s(off, rng uint64, out []uint64) {
	var val uint64
	mask := rng
	var i int
	cnt := len(out)

	if rng == 0 {
		for i = 0; i < cnt; i++ {
			out[i] = off
		}
		return
	}

	/* Smallest bit mask >= max */
	mask |= mask >> 1
	mask |= mask >> 2
	mask |= mask >> 4
	mask |= mask >> 8
	mask |= mask >> 16
	mask |= mask >> 32

	for i = 0; i < cnt; i++ {
		if rng <= 0xffffffff {
			for {
				val = uint64(state.Uint32() & uint32(mask))
				if val <= rng {
					break
				}
			}
		} else {
			for {
				val = state.Uint64() & mask
				if val <= rng {
					break
				}
			}
		}
		out[i] = off + val
	}
}

// Float64 uniform generator
func (state *RKState) Float64() float64 {
	/* shifts : 67108864 = 0x4000000, 9007199254740992 = 0x20000000000000 */
	a, b := float64(state.Uint32()>>5), float64(state.Uint32()>>6)
	return (a*67108864.0 + b) / 9007199254740992.0
}

// NormFloat64 normal generator
func (state *RKState) NormFloat64() float64 {
	if state.hasGauss {
		tmp := state.gauss
		state.gauss = 0
		state.hasGauss = false
		return tmp
	}
	var f, x1, x2, r2 float64

	for {
		x1 = 2.0*state.Float64() - 1.0
		x2 = 2.0*state.Float64() - 1.0
		r2 = x1*x1 + x2*x2
		if r2 >= 1.0 || r2 == 0.0 {
			continue
		}
		break
	}

	/* Polar method, a more efficient version of the Box-Muller approach. */
	f = math.Sqrt(-2.0 * math.Log(r2) / r2)
	/* Keep for next call */
	state.gauss = f * x1
	state.hasGauss = true
	return f * x2

}

// Int63 generator
func (state *RKState) Int63() int64 {
	return int64(state.Uint64() &^ (1 << 63))
}

// NewRandomkitSource create a new randomkit source
func NewRandomkitSource(seed uint64) (state *RKState) {
	state = &RKState{}
	state.Seed(seed)
	return
}

// Clone clones the randomkit state
func (state *RKState) Clone() rand.Source {
	newstate := *state
	return &newstate
}

// Clone clones the randomkit state
func (state *RKState) SourceClone() rand.Source {
	newstate := *state
	return &newstate
}

const maxUint64 = (1 << 64) - 1

// Uint64n returns, as a uint64, a pseudo-random number in [0,n).
// It is guaranteed more uniform than taking a Source value mod n
// for any n that is not a power of 2.
func (state *RKState) Uint64n(n uint64) uint64 {
	if n&(n-1) == 0 { // n is power of two, can mask
		if n == 0 {
			panic("invalid argument to Uint64n")
		}
		return state.Uint64() & (n - 1)
	}
	// If n does not divide v, to avoid bias we must not use
	// a v that is within maxUint64%n of the top of the range.
	v := state.Uint64()
	if v > maxUint64-n { // Fast check.
		ceiling := maxUint64 - maxUint64%n
		for v >= ceiling {
			v = state.Uint64()
		}
	}

	return v % n
}

// Intn returns, as an int, a non-negative pseudo-random number in [0,n).
// It panics if n <= 0.
func (state *RKState) Intn(n int) int {
	if n <= 0 {
		panic("invalid argument to Intn")
	}
	// TODO: Avoid some 64-bit ops to make it more efficient on 32-bit machines.
	return int(state.Uint64n(uint64(n)))
}

func (state *RKState) Shuffle(n int, swap func(i, j int)) {
	for i := n - 1; i >= 1; i-- {
		swap(i, int(random_interval(state, uint64(i))))
	}
	return
}

func (state *RKState) Perm(n int) []int {
	idx := make([]int, n)
	for i := range idx {
		idx[i] = i
	}
	slice := sort.IntSlice(idx)
	state.Shuffle(slice.Len(), slice.Swap)
	return idx
}

func random_interval(state *RKState, max uint64) uint64 {
	var mask, value uint64
	if max == 0 {
		return 0
	}

	mask = max

	/* Smallest bit mask >= max */
	mask |= mask >> 1
	mask |= mask >> 2
	mask |= mask >> 4
	mask |= mask >> 8
	mask |= mask >> 16
	mask |= mask >> 32

	/* Search a random value in [0..mask] <= max */
	var ok bool
	if max <= 0xffffffff {

		for !ok {
			value = uint64(state.Uint32() & uint32(mask))
			ok = value <= max
		}
	} else {
		for !ok {
			value = (state.Uint64() & mask)
			ok = value <= max
		}
	}
	return value
}
