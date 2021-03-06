package main

import (
	"errors"
	"log"
	"math"
	"sync/atomic"

	blocks "github.com/ipfs/go-block-format"
	blockstore "github.com/ipfs/go-ipfs-blockstore"
)

var ErrLimitReached = errors.New("blockstore limit reached")

// LimitBlockstore allows a limited set of put operations on the underlying
// blockstore. Put operations beyond the limit are rejected with ErrLimitReached.
type LimitedBlockstore struct {
	blockstore.Blockstore

	puts  int64
	limit int64
}

var _ blockstore.Blockstore = (*LimitedBlockstore)(nil)

// LimitBlockstore intercepts a blockstore and limits the amount of puts that
// it will take, returning ErrLimitReached when its limit is reached. A zero
// limit means unbounded.
func LimitBlockstore(inner blockstore.Blockstore, limit int64) *LimitedBlockstore {
	if limit == 0 {
		limit = math.MaxInt64
	}
	return &LimitedBlockstore{Blockstore: inner, limit: limit}
}

func (l *LimitedBlockstore) Put(block blocks.Block) error {
	curr := atomic.AddInt64(&l.puts, 1)
	if curr > l.limit {
		return ErrLimitReached
	}
	if curr%100000 == 0 {
		log.Printf("import: progress: %d blocks imported", curr)
	}
	return l.Blockstore.Put(block)
}

func (l *LimitedBlockstore) PutMany(blocks []blocks.Block) error {
	cnt := int64(len(blocks))
	curr := atomic.AddInt64(&l.puts, cnt)
	if curr > l.limit {
		return ErrLimitReached
	}

	if int(curr/100000) != int((curr-cnt)/100000) {
		// crossed a 1000 multiple.
		log.Printf("import: progress: %d blocks imported", curr)
	}

	return l.Blockstore.PutMany(blocks)
}
