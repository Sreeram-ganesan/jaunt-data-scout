package config

import (
    "os"
    "path/filepath"
    "testing"
    "time"
)

func TestLoadAndEnvOverride(t *testing.T) {
    path := filepath.Join("..", "..", "..", "..", "..", "config", "defaults.yaml")
    rd, err := LoadDefaults(path)
    if err != nil {
        t.Fatalf("load defaults: %v", err)
    }
    
    // Test token bucket override
    os.Setenv("BUDGET_GOOGLE_TEXT_CAPACITY", "1234")
    defer os.Unsetenv("BUDGET_GOOGLE_TEXT_CAPACITY")
    
    // Test global setting override  
    os.Setenv("BUDGET_MAX_API_CALLS", "7500")
    defer os.Unsetenv("BUDGET_MAX_API_CALLS")
    
    // Test split ratio override
    os.Setenv("BUDGET_SPLIT_RATIO", "0.8")
    defer os.Unsetenv("BUDGET_SPLIT_RATIO")
    
    ApplyEnvOverrides(&rd)
    
    if rd.Budgets["google.text"].Capacity != 1234 {
        t.Fatalf("expected override capacity 1234, got %d", rd.Budgets["google.text"].Capacity)
    }
    
    if rd.CityDefaults.Budgets.MaxAPICalls != 7500 {
        t.Fatalf("expected override max_api_calls 7500, got %d", rd.CityDefaults.Budgets.MaxAPICalls)
    }
    
    if rd.SplitRatio != 0.8 {
        t.Fatalf("expected override split_ratio 0.8, got %f", rd.SplitRatio)
    }
}

func TestBuildBudgetConfig(t *testing.T) {
    rd := RawDefaults{
        SplitRatio: 0.75,
        Budgets: map[string]struct {
            Capacity int64         `yaml:"capacity"`
            Refill   int64         `yaml:"refill"`
            Period   time.Duration `yaml:"period"`
        }{
            "google.text": {Capacity: 1000, Refill: 100, Period: time.Minute},
        },
    }
    
    cfg := BuildBudgetConfig(rd)
    if cfg.SplitRatio != 0.75 {
        t.Fatalf("expected split ratio 0.75, got %f", cfg.SplitRatio)
    }
    
    if len(cfg.Budgets) != 1 {
        t.Fatalf("expected 1 budget, got %d", len(cfg.Budgets))
    }
}
