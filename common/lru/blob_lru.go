// Copyright 2022 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

package lru

import (
	"math"
	"sync"
)

// blobType is the type constraint for values stored in SizeConstrainedCache.
// blobType 是存储在 SizeConstrainedCache 中的值的类型约束。
type blobType interface {
	~[]byte | ~string
}

// SizeConstrainedCache is a cache where capacity is in bytes (instead of item count). When the cache
// is at capacity, and a new item is added, older items are evicted until the size
// constraint is met.
//
// OBS: This cache assumes that items are content-addressed: keys are unique per content.
// In other words: two Add(..) with the same key K, will always have the same value V.
// SizeConstrainedCache 是一种以字节为单位（而不是以项目数量为单位）计算容量的缓存。当缓存已满并且添加新项目时，较旧的项目会被逐出，直到满足大小约束。
//
// 注意：此缓存假设项目是内容寻址的：每个内容的键都是唯一的。换句话说：使用相同键 K 的两个 Add(..) 操作，其值 V 始终相同。
type SizeConstrainedCache[K comparable, V blobType] struct {
	size    uint64
	maxSize uint64
	lru     BasicLRU[K, V]
	lock    sync.Mutex
}

// NewSizeConstrainedCache creates a new size-constrained LRU cache.
func NewSizeConstrainedCache[K comparable, V blobType](maxSize uint64) *SizeConstrainedCache[K, V] {
	return &SizeConstrainedCache[K, V]{
		size:    0,
		maxSize: maxSize,
		lru:     NewBasicLRU[K, V](math.MaxInt),
	}
}

// Add adds a value to the cache.  Returns true if an eviction occurred.
// OBS: This cache assumes that items are content-addressed: keys are unique per content.
// In other words: two Add(..) with the same key K, will always have the same value V.
// OBS: The value is _not_ copied on Add, so the caller must not modify it afterwards.
// Add 将值添加到缓存中。如果发生了逐出操作，则返回 true。
// 注意：此缓存假设项目是内容寻址的：每个内容的键都是唯一的。换句话说：使用相同键 K 的两个 Add(..) 操作，其值 V 始终相同。
// 注意：在调用 Add 时，值不会被复制，因此调用方在之后不得修改该值。
func (c *SizeConstrainedCache[K, V]) Add(key K, value V) (evicted bool) {
	c.lock.Lock()
	defer c.lock.Unlock()

	// Unless it is already present, might need to evict something.
	// OBS: If it is present, we still call Add internally to bump the recentness.
	if !c.lru.Contains(key) {
		targetSize := c.size + uint64(len(value))
		for targetSize > c.maxSize {
			evicted = true
			_, v, ok := c.lru.RemoveOldest()
			if !ok {
				// list is now empty. Break
				break
			}
			targetSize -= uint64(len(v))
		}
		c.size = targetSize
	}
	c.lru.Add(key, value)
	return evicted
}

// Get looks up a key's value from the cache.
func (c *SizeConstrainedCache[K, V]) Get(key K) (V, bool) {
	c.lock.Lock()
	defer c.lock.Unlock()

	return c.lru.Get(key)
}
