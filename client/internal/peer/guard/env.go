package guard

import (
	"os"
	"strconv"
	"time"

	log "github.com/sirupsen/logrus"
)

const (
	envICEMonitorPeriod = "NB_ICE_MONITOR_PERIOD"
	envICERetryInterval = "NB_ICE_RETRY_INTERVAL"
)

func GetICEMonitorPeriod() time.Duration {
	if envVal := os.Getenv(envICEMonitorPeriod); envVal != "" {
		if seconds, err := strconv.Atoi(envVal); err == nil && seconds > 0 {
			return time.Duration(seconds) * time.Second
		}
	}
	return defaultCandidatesMonitorPeriod
}

// getICERetryInterval returns how often ICE is retried once the retry budget is spent,
// read in seconds from NB_ICE_RETRY_INTERVAL.
func getICERetryInterval() time.Duration {
	envVal := os.Getenv(envICERetryInterval)
	if envVal == "" {
		return defaultICERetryInterval
	}
	seconds, err := strconv.Atoi(envVal)
	if err != nil || seconds <= 0 {
		log.Warnf("invalid %s %q, using %s", envICERetryInterval, envVal, defaultICERetryInterval)
		return defaultICERetryInterval
	}
	return time.Duration(seconds) * time.Second
}
