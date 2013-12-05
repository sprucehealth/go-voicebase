// Copyright 2012 Samuel Stauffer. All rights reserved.
// Use of this source code is governed by a 3-clause BSD
// license that can be found in the LICENSE file.

package metrics

import (
	"fmt"
	"math"
	"strconv"
	"sync/atomic"
	"unsafe"
)

type variance struct {
	m float64
	s float64
}

type Distribution struct {
	count    uint64
	sum      int64
	min      int64
	max      int64
	variance unsafe.Pointer // pointer to variance struct
}

func NewDistribution() *Distribution {
	return &Distribution{
		min:      math.MaxInt64,
		max:      math.MinInt64,
		variance: unsafe.Pointer(&variance{-1, 0}),
	}
}

func (d *Distribution) String() string {
	return fmt.Sprintf("{\"count\":%d,\"sum\":%d,\"min\":%d,\"max\":%d,\"stddev\":%s}",
		d.Count(), d.Sum(), d.Min(), d.Max(), strconv.FormatFloat(d.StdDev(), 'g', -1, 64))
}

func (d *Distribution) Clear() {
	atomic.StoreUint64(&d.count, 0)
	atomic.StoreInt64(&d.sum, 0)
	atomic.StoreInt64(&d.min, math.MaxInt64)
	atomic.StoreInt64(&d.max, math.MinInt64)
	atomic.StorePointer(&d.variance, unsafe.Pointer(&variance{-1, 0}))
}

func (d *Distribution) Update(value int64) {
	atomic.AddUint64(&d.count, 1)
	atomic.AddInt64(&d.sum, value)
	for {
		min := atomic.LoadInt64(&d.min)
		if value > min || atomic.CompareAndSwapInt64(&d.min, min, value) {
			break
		}
	}
	for {
		max := atomic.LoadInt64(&d.max)
		if value < max || atomic.CompareAndSwapInt64(&d.max, max, value) {
			break
		}
	}
	floatValue := float64(value)
	newV := &variance{}
	for {
		uv := atomic.LoadPointer(&d.variance)
		v := (*variance)(uv)
		oldM := v.m
		if oldM == -1 {
			newV.m = floatValue
			newV.s = 0
		} else {
			newV.m = oldM + ((floatValue - oldM) / float64(atomic.LoadUint64(&d.count)))
			newV.s = v.s + ((floatValue - oldM) * (floatValue - newV.m))
		}
		if atomic.CompareAndSwapPointer(&d.variance, uv, unsafe.Pointer(newV)) {
			break
		}
	}
}

func (d *Distribution) Count() uint64 {
	return atomic.LoadUint64(&d.count)
}

func (d *Distribution) Sum() int64 {
	return atomic.LoadInt64(&d.sum)
}

func (d *Distribution) Min() int64 {
	return atomic.LoadInt64(&d.min)
}

func (d *Distribution) Max() int64 {
	return atomic.LoadInt64(&d.max)
}

func (d *Distribution) Mean() int64 {
	return atomic.LoadInt64(&d.sum) / int64(atomic.LoadUint64(&d.count))
}

func (d *Distribution) Variance() float64 {
	count := atomic.LoadUint64(&d.count)
	if count <= 1 {
		return 0.0
	}
	v := (*variance)(atomic.LoadPointer(&d.variance))
	return v.s / float64(count-1)
}

func (d *Distribution) StdDev() float64 {
	return math.Sqrt(d.Variance())
}
