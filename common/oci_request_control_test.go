package common

import "testing"

func TestNormalizeWaitSettings(t *testing.T) {
	maxWait, interval := NormalizeWaitSettings(0, 0)
	if maxWait != 1200 || interval != 5 {
		t.Fatalf("unexpected defaults: maxWait=%d interval=%d", maxWait, interval)
	}
	maxWait, interval = NormalizeWaitSettings(30, 2)
	if maxWait != 30 || interval != 2 {
		t.Fatalf("expected explicit values to be preserved, got maxWait=%d interval=%d", maxWait, interval)
	}
}
