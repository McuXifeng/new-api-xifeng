package service

import (
	"testing"
	"time"
)

func TestMemoryRiskMetricStoreReadPathDoesNotCreateState(t *testing.T) {
	store := newMemoryRiskMetricStore()
	now := time.Unix(1713175200, 0)

	block, err := store.GetBlock(RiskSubjectTypeToken, 1001)
	if err != nil {
		t.Fatalf("GetBlock returned error: %v", err)
	}
	if block != nil {
		t.Fatalf("GetBlock returned unexpected block: %#v", block)
	}
	if err = store.ClearBlock(RiskSubjectTypeToken, 1001); err != nil {
		t.Fatalf("ClearBlock returned error: %v", err)
	}
	inflight, err := store.RecordFinish(RiskSubjectTypeToken, 1001, now)
	if err != nil {
		t.Fatalf("RecordFinish returned error: %v", err)
	}
	if inflight != 0 {
		t.Fatalf("RecordFinish returned unexpected inflight: %d", inflight)
	}
	hitCount, err := store.GetRuleHitCount(RiskSubjectTypeToken, 1001, now)
	if err != nil {
		t.Fatalf("GetRuleHitCount returned error: %v", err)
	}
	if hitCount != 0 {
		t.Fatalf("GetRuleHitCount returned unexpected count: %d", hitCount)
	}
	if got := len(store.subject); got != 0 {
		t.Fatalf("read-only paths should not create memory states, got %d", got)
	}
}

func TestMemoryRiskMetricStoreSweepsIdleSubjectState(t *testing.T) {
	store := newMemoryRiskMetricStore()
	now := time.Unix(1713175200, 0)

	metrics, err := store.RecordStart(RiskSubjectTypeToken, 1001, "ip-1", "ua-1", now)
	if err != nil {
		t.Fatalf("RecordStart returned error: %v", err)
	}
	if metrics.RequestCount1M != 1 {
		t.Fatalf("RecordStart returned unexpected metrics: %#v", metrics)
	}
	if _, err = store.RecordFinish(RiskSubjectTypeToken, 1001, now); err != nil {
		t.Fatalf("RecordFinish returned error: %v", err)
	}
	if got := len(store.subject); got != 1 {
		t.Fatalf("expected one active memory state, got %d", got)
	}

	// Trigger a global sweep from another key before the original subject goes idle.
	if _, err = store.GetRuleHitCount(RiskSubjectTypeToken, 2002, now); err != nil {
		t.Fatalf("GetRuleHitCount before idle window returned error: %v", err)
	}
	if got := len(store.subject); got != 1 {
		t.Fatalf("state should still exist before retention windows expire, got %d", got)
	}

	later := now.Add(2 * time.Hour)
	if _, err = store.GetRuleHitCount(RiskSubjectTypeToken, 3003, later); err != nil {
		t.Fatalf("GetRuleHitCount during sweep returned error: %v", err)
	}
	if got := len(store.subject); got != 0 {
		t.Fatalf("idle subject state should be swept after retention windows expire, got %d", got)
	}
}
