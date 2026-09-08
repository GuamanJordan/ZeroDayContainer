package cgroups

import (
	"testing"
)

func TestParseMemoryBytes(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
		hasErr   bool
	}{
		{"", 0, false},
		{"1024", 1024, false},
		{"4k", 4 * 1024, false},
		{"4kb", 4 * 1024, false},
		{"50m", 50 * 1024 * 1024, false},
		{"50MB", 50 * 1024 * 1024, false},
		{"1g", 1024 * 1024 * 1024, false},
		{"2GB", 2 * 1024 * 1024 * 1024, false},
		{"invalid", 0, true},
	}

	for _, tt := range tests {
		got, err := ParseMemoryBytes(tt.input)
		if tt.hasErr && err == nil {
			t.Errorf("para '%s' se esperaba error", tt.input)
		}
		if !tt.hasErr && err != nil {
			t.Errorf("para '%s' no se esperaba error: %v", tt.input, err)
		}
		if got != tt.expected {
			t.Errorf("para '%s' se esperaba %d, se obtuvo %d", tt.input, tt.expected, got)
		}
	}
}

func TestFormatCPUQuotaPeriod(t *testing.T) {
	tests := []struct {
		cpus     float64
		expected string
		hasErr   bool
	}{
		{0.5, "50000 100000", false},
		{1.0, "100000 100000", false},
		{2.0, "200000 100000", false},
		{0.0, "", true},
		{-1.0, "", true},
	}

	for _, tt := range tests {
		got, err := FormatCPUQuotaPeriod(tt.cpus)
		if tt.hasErr && err == nil {
			t.Errorf("para cpus=%f se esperaba error", tt.cpus)
		}
		if !tt.hasErr && err != nil {
			t.Errorf("para cpus=%f no se esperaba error: %v", tt.cpus, err)
		}
		if got != tt.expected {
			t.Errorf("para cpus=%f se esperaba '%s', se obtuvo '%s'", tt.cpus, tt.expected, got)
		}
	}
}
