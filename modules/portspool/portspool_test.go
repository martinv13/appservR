package portspool

import (
	"strconv"
	"sync"
	"testing"
)

// resetPool clears global pool state so each test starts from a known baseline.
func resetPool(t *testing.T) {
	t.Helper()
	portsPool.Lock()
	portsPool.inUse = make(map[string]bool)
	portsPool.Unlock()
}

func TestGetNextReturnsFirstFreePort(t *testing.T) {
	resetPool(t)

	port, err := GetNext()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if port != strconv.Itoa(portsPool.rangeStart) {
		t.Errorf("expected first allocated port to be %d, got %s", portsPool.rangeStart, port)
	}

	portsPool.Lock()
	inUse := portsPool.inUse[port]
	portsPool.Unlock()
	if !inUse {
		t.Errorf("expected port %s to be marked in use", port)
	}
}

func TestGetNextSkipsPortsAlreadyMarkedInUse(t *testing.T) {
	resetPool(t)

	first := strconv.Itoa(portsPool.rangeStart)
	portsPool.Lock()
	portsPool.inUse[first] = true
	portsPool.Unlock()

	port, err := GetNext()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if port == first {
		t.Errorf("expected GetNext to skip already in-use port %s", first)
	}
}

func TestReleaseAllowsPortReuse(t *testing.T) {
	resetPool(t)

	port, err := GetNext()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	Release(port)

	portsPool.Lock()
	_, stillTracked := portsPool.inUse[port]
	portsPool.Unlock()
	if stillTracked {
		t.Errorf("expected port %s to be removed from inUse map after Release", port)
	}

	again, err := GetNext()
	if err != nil {
		t.Fatalf("unexpected error on second GetNext: %v", err)
	}
	if again != port {
		t.Errorf("expected released port %s to be reallocated, got %s", port, again)
	}
}

func TestReleaseUnknownPortIsNoop(t *testing.T) {
	resetPool(t)
	// Should not panic even though this port was never allocated.
	Release("9999")
}

func TestGetNextConcurrentAllocatesDistinctPorts(t *testing.T) {
	resetPool(t)

	const n = 20
	ports := make([]string, n)
	errs := make([]error, n)
	var wg sync.WaitGroup
	wg.Add(n)
	for i := 0; i < n; i++ {
		i := i
		go func() {
			defer wg.Done()
			ports[i], errs[i] = GetNext()
		}()
	}
	wg.Wait()

	seen := map[string]bool{}
	for i, p := range ports {
		if errs[i] != nil {
			t.Fatalf("goroutine %d: unexpected error: %v", i, errs[i])
		}
		if seen[p] {
			t.Errorf("port %s was allocated more than once", p)
		}
		seen[p] = true
	}
	if len(seen) != n {
		t.Errorf("expected %d distinct ports, got %d", n, len(seen))
	}
}

func TestGetNextReturnsErrorWhenPoolExhausted(t *testing.T) {
	resetPool(t)

	portsPool.Lock()
	for port := portsPool.rangeStart; port < portsPool.rangeStart+1000; port++ {
		portsPool.inUse[strconv.Itoa(port)] = true
	}
	portsPool.Unlock()

	_, err := GetNext()
	if err == nil {
		t.Fatal("expected error when pool is exhausted")
	}
}
