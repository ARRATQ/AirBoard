package utils

import (
	"testing"
	"time"
)

func newTestLockout(max int, lockout time.Duration, clock *time.Time) *AuthSecurityManager {
	asm := NewAuthSecurityManagerWithLimits(max, lockout)
	asm.now = func() time.Time { return *clock }
	return asm
}

func TestLockoutAfterMaxAttempts(t *testing.T) {
	clock := time.Date(2026, 11, 19, 9, 0, 0, 0, time.UTC)
	asm := newTestLockout(3, 15*time.Minute, &clock)

	for i := 1; i <= 2; i++ {
		if locked, _ := asm.RecordFailedLogin("alice"); locked {
			t.Fatalf("locked after %d failures, want 3", i)
		}
		// Une vérification entre deux échecs ne doit pas effacer le compteur.
		if locked, _ := asm.CheckFailedLogin("alice"); locked {
			t.Fatalf("CheckFailedLogin reports a lock after %d failures", i)
		}
	}

	locked, duration := asm.RecordFailedLogin("alice")
	if !locked || duration != 15*time.Minute {
		t.Fatalf("third failure: locked=%v duration=%v, want locked for 15m", locked, duration)
	}
	if locked, _ := asm.CheckFailedLogin("alice"); !locked {
		t.Fatal("CheckFailedLogin must report the lock")
	}
	if locked, _ := asm.CheckFailedLogin("bob"); locked {
		t.Fatal("another identifier must not be locked")
	}

	clock = clock.Add(16 * time.Minute)
	if locked, _ := asm.CheckFailedLogin("alice"); locked {
		t.Fatal("the lock must expire")
	}
	if locked, _ := asm.RecordFailedLogin("alice"); locked {
		t.Fatal("the counter must restart after the lock expires")
	}
}

func TestFailuresOutsideWindowAreForgotten(t *testing.T) {
	clock := time.Date(2026, 11, 19, 9, 0, 0, 0, time.UTC)
	asm := newTestLockout(3, 15*time.Minute, &clock)

	asm.RecordFailedLogin("alice")
	asm.RecordFailedLogin("alice")
	clock = clock.Add(20 * time.Minute)

	if locked, _ := asm.RecordFailedLogin("alice"); locked {
		t.Fatal("failures older than the window must not count")
	}
}

func TestSuccessfulLoginClearsFailures(t *testing.T) {
	clock := time.Date(2026, 11, 19, 9, 0, 0, 0, time.UTC)
	asm := newTestLockout(3, 15*time.Minute, &clock)

	asm.RecordFailedLogin("alice")
	asm.RecordFailedLogin("alice")
	asm.RecordSuccessfulLogin("alice")

	if locked, _ := asm.RecordFailedLogin("alice"); locked {
		t.Fatal("a successful login must reset the counter")
	}
}

func TestHasRecentFailures(t *testing.T) {
	clock := time.Date(2026, 11, 19, 9, 0, 0, 0, time.UTC)
	asm := newTestLockout(3, 15*time.Minute, &clock)

	if asm.HasRecentFailures("1.2.3.4|id:1") {
		t.Fatal("no failure recorded yet")
	}
	asm.RecordFailedLogin("1.2.3.4|id:1")
	if !asm.HasRecentFailures("1.2.3.4|id:1") {
		t.Fatal("a fresh failure must count")
	}
	clock = clock.Add(16 * time.Minute)
	if asm.HasRecentFailures("1.2.3.4|id:1") {
		t.Fatal("a failure outside the window must not count")
	}
}
