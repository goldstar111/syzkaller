// Copyright 2017 syzkaller project authors. All rights reserved.
// Use of this source code is governed by Apache 2 LICENSE that can be found in the LICENSE file.

package repro

import (
	"math/rand"
	"testing"
	"time"

	"github.com/google/syzkaller/pkg/csource"
	"github.com/google/syzkaller/prog"
)

func initTest(t *testing.T) (*rand.Rand, int) {
	iters := 1000
	if testing.Short() {
		iters = 100
	}
	seed := int64(time.Now().UnixNano())
	rs := rand.NewSource(seed)
	t.Logf("seed=%v", seed)
	return rand.New(rs), iters
}

func TestBisect(t *testing.T) {
	ctx := &context{}

	rd, iters := initTest(t)
	for n := 0; n < iters; n++ {
		var progs []*prog.LogEntry
		numTotal := rd.Intn(300)
		numGuilty := 0
		for i := 0; i < numTotal; i++ {
			var prog prog.LogEntry
			if rd.Intn(30) == 0 {
				prog.Proc = 42
				numGuilty += 1
			}
			progs = append(progs, &prog)
		}
		if numGuilty == 0 {
			var prog prog.LogEntry
			prog.Proc = 42
			progs = append(progs, &prog)
			numGuilty += 1
		}
		progs, _ = ctx.bisectProgs(progs, func(p []*prog.LogEntry) (bool, error) {
			guilty := 0
			for _, prog := range p {
				if prog.Proc == 42 {
					guilty += 1
				}
			}
			return guilty == numGuilty, nil
		})
		if len(progs) != numGuilty {
			t.Fatalf("bisect test failed: wrong number of guilty progs: got: %v, want: %v", len(progs), numGuilty)
		}
		for _, prog := range progs {
			if prog.Proc != 42 {
				t.Fatalf("bisect test failed: wrong program is guilty: progs: %v", progs)
			}
		}
	}
}

func TestSimplifies(t *testing.T) {
	opts := csource.Options{
		Threaded:   true,
		Collide:    true,
		Repeat:     true,
		Procs:      10,
		Sandbox:    "namespace",
		EnableTun:  true,
		UseTmpDir:  true,
		HandleSegv: true,
		WaitRepeat: true,
		Repro:      true,
	}
	if err := opts.Check(); err != nil {
		t.Fatalf("initial opts are invalid: %v", err)
	}
	for i, fn := range cSimplifies {
		fn(&opts)
		if err := opts.Check(); err != nil {
			t.Fatalf("opts %v are invalid: %v", i, err)
		}
	}
}
