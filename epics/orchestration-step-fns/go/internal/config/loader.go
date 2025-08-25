package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	b "github.com/Sreeram-ganesan/jaunt-data-scout/epics/orchestration-step-fns/go/internal/budget"
	"gopkg.in/yaml.v3"
)

// RawDefaults mirrors config/defaults.yaml
type RawDefaults struct {
	Version      int `yaml:"version"`
	CityDefaults struct {
		EarlyStop struct {
			MinNewUniqueRate float64 `yaml:"min_new_unique_rate"`
			Window           int     `yaml:"window"`
		} `yaml:"early_stop"`
		Budgets struct {
			MaxAPICalls       int `yaml:"max_api_calls"`
			MaxWallClockHours int `yaml:"max_wall_clock_hours"`
		} `yaml:"budgets"`
	} `yaml:"city_defaults"`
	Budgets map[string]struct {
		Capacity int64         `yaml:"capacity"`
		Refill   int64         `yaml:"refill"`
		Period   time.Duration `yaml:"period"`
	} `yaml:"budgets"`
	SplitRatio  float64        `yaml:"split_ratio"`
	Concurrency map[string]int `yaml:"concurrency"`
}

func LoadDefaults(path string) (RawDefaults, error) {
	var rd RawDefaults
	data, err := os.ReadFile(path)
	if err != nil {
		return rd, err
	}
	if err := yaml.Unmarshal(data, &rd); err != nil {
		return rd, err
	}
	return rd, nil
}

func ApplyEnvOverrides(rd *RawDefaults) {
	// Global ratios and caps
	if v := os.Getenv("BUDGET_SPLIT_RATIO"); v != "" {
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			rd.SplitRatio = f
		}
	}
	if v := os.Getenv("EARLY_STOP_MIN_NEW_UNIQUE_RATE"); v != "" {
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			rd.CityDefaults.EarlyStop.MinNewUniqueRate = f
		}
	}
	if v := os.Getenv("EARLY_STOP_WINDOW"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			rd.CityDefaults.EarlyStop.Window = n
		}
	}
	if v := os.Getenv("BUDGET_MAX_API_CALLS"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			rd.CityDefaults.Budgets.MaxAPICalls = n
		}
	}
	if v := os.Getenv("BUDGET_MAX_WALL_CLOCK_HOURS"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			rd.CityDefaults.Budgets.MaxWallClockHours = n
		}
	}

	// Concurrency overrides: CONCURRENCY_<KEY>
	for k := range rd.Concurrency {
		envKey := "CONCURRENCY_" + normalizeKey(k)
		if v := os.Getenv(envKey); v != "" {
			if n, err := strconv.Atoi(v); err == nil {
				rd.Concurrency[k] = n
			}
		}
	}

	// Budget token buckets: BUDGET_<TOKEN>_<FIELD>
	for token, cfg := range rd.Budgets {
		base := "BUDGET_" + normalizeKey(token)
		if v := os.Getenv(base + "_CAPACITY"); v != "" {
			if n, err := strconv.ParseInt(v, 10, 64); err == nil {
				cfg.Capacity = n
			}
		}
		if v := os.Getenv(base + "_REFILL"); v != "" {
			if n, err := strconv.ParseInt(v, 10, 64); err == nil {
				cfg.Refill = n
			}
		}
		if v := os.Getenv(base + "_PERIOD"); v != "" {
			if d, err := time.ParseDuration(v); err == nil {
				cfg.Period = d
			}
		}
		rd.Budgets[token] = cfg
	}
}

func normalizeKey(s string) string {
	return strings.ToUpper(strings.ReplaceAll(s, ".", "_"))
}

// BuildBudgetConfig adapts RawDefaults to internal/budget.Config
func BuildBudgetConfig(rd RawDefaults) b.Config {
	out := b.Config{
		Budgets: make(map[b.Connector]struct {
			Capacity int64         `yaml:"capacity"`
			Refill   int64         `yaml:"refill"`
			Period   time.Duration `yaml:"period"`
		}),
		SplitRatio: rd.SplitRatio,
	}
	for token, cfg := range rd.Budgets {
		out.Budgets[b.Connector(token)] = struct {
			Capacity int64         `yaml:"capacity"`
			Refill   int64         `yaml:"refill"`
			Period   time.Duration `yaml:"period"`
		}{Capacity: cfg.Capacity, Refill: cfg.Refill, Period: cfg.Period}
	}
	return out
}

// Human summary for logs/debugging
func (rd RawDefaults) String() string {
	return fmt.Sprintf("split_ratio=%.2f, early_stop{rate=%.3f, window=%d}, caps{api_calls=%d, wall_clock_h=%d}, budgets=%d tokens",
		rd.SplitRatio,
		rd.CityDefaults.EarlyStop.MinNewUniqueRate,
		rd.CityDefaults.EarlyStop.Window,
		rd.CityDefaults.Budgets.MaxAPICalls,
		rd.CityDefaults.Budgets.MaxWallClockHours,
		len(rd.Budgets),
	)
}
