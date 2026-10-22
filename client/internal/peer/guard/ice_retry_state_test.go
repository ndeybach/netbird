package guard

import (
	"os"
	"testing"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func newTestRetryState() *iceRetryState {
	return &iceRetryState{log: log.NewEntry(log.StandardLogger())}
}

func TestICERetryState_AllowsInitialBudget(t *testing.T) {
	s := newTestRetryState()

	for i := 1; i <= maxICERetries; i++ {
		if !s.shouldRetry() {
			t.Fatalf("shouldRetry returned false on attempt %d, want true (budget = %d)", i, maxICERetries)
		}
	}
}

func TestICERetryState_ExhaustsAfterBudget(t *testing.T) {
	s := newTestRetryState()

	for i := 0; i < maxICERetries; i++ {
		_ = s.shouldRetry()
	}

	if s.shouldRetry() {
		t.Fatalf("shouldRetry returned true after budget exhausted, want false")
	}
}

func TestICERetryState_HourlyCNilBeforeEnterHourlyMode(t *testing.T) {
	s := newTestRetryState()

	if s.hourlyC() != nil {
		t.Fatalf("hourlyC returned non-nil channel before enterHourlyMode")
	}
}

func TestICERetryState_EnterHourlyModeArmsTicker(t *testing.T) {
	s := newTestRetryState()
	for i := 0; i < maxICERetries+1; i++ {
		_ = s.shouldRetry()
	}

	s.enterHourlyMode()
	defer s.reset()

	if s.hourlyC() == nil {
		t.Fatalf("hourlyC returned nil after enterHourlyMode")
	}
}

func TestICERetryState_ShouldRetryTrueInHourlyMode(t *testing.T) {
	s := newTestRetryState()
	s.enterHourlyMode()
	defer s.reset()

	if !s.shouldRetry() {
		t.Fatalf("shouldRetry returned false in hourly mode, want true")
	}

	// Subsequent calls also return true — we keep retrying on each hourly tick.
	if !s.shouldRetry() {
		t.Fatalf("second shouldRetry returned false in hourly mode, want true")
	}
}

func TestICERetryState_ResetRestoresBudget(t *testing.T) {
	s := newTestRetryState()
	for i := 0; i < maxICERetries+1; i++ {
		_ = s.shouldRetry()
	}
	s.enterHourlyMode()

	s.reset()

	if s.hourlyC() != nil {
		t.Fatalf("hourlyC returned non-nil channel after reset")
	}
	if s.retries != 0 {
		t.Fatalf("retries = %d after reset, want 0", s.retries)
	}

	for i := 1; i <= maxICERetries; i++ {
		if !s.shouldRetry() {
			t.Fatalf("shouldRetry returned false on attempt %d after reset, want true", i)
		}
	}
}

func TestICERetryState_ResetIsIdempotent(t *testing.T) {
	s := newTestRetryState()
	s.reset()
	s.reset() // second call must not panic or re-stop a nil ticker

	if s.hourlyC() != nil {
		t.Fatalf("hourlyC non-nil after double reset")
	}
}

func TestGetICERetryInterval(t *testing.T) {
	tests := []struct {
		name  string
		value string
		want  time.Duration
	}{
		{name: "seconds", value: "120", want: 2 * time.Minute},
		{name: "zero", value: "0", want: defaultICERetryInterval},
		{name: "negative", value: "-5", want: defaultICERetryInterval},
		{name: "not a number", value: "5m", want: defaultICERetryInterval},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv(envICERetryInterval, tt.value)
			assert.Equal(t, tt.want, getICERetryInterval(), "interval for %q", tt.value)
		})
	}

	t.Run("unset", func(t *testing.T) {
		t.Setenv(envICERetryInterval, "")
		os.Unsetenv(envICERetryInterval)
		assert.Equal(t, defaultICERetryInterval, getICERetryInterval(), "interval when unset")
	})
}

func TestICERetryState_EnterHourlyModeUsesConfiguredInterval(t *testing.T) {
	t.Setenv(envICERetryInterval, "1")
	s := newTestRetryState()
	s.enterHourlyMode()
	defer s.reset()

	select {
	case <-s.hourlyC():
	case <-time.After(3 * time.Second):
		t.Fatalf("no retry tick within 3s with %s=1", envICERetryInterval)
	}
}
