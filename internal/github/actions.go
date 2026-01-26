package github

import (
	"fmt"
	"os"
)

// Actions integration helper for GitHub Actions runtime environment
type Actions struct {
	StepSummaryPath string
	OutputPath      string
}

func NewActions() *Actions {
	return &Actions{
		StepSummaryPath: os.Getenv("GITHUB_STEP_SUMMARY"),
		OutputPath:      os.Getenv("GITHUB_OUTPUT"),
	}
}

// WriteSummary appends markdown to the Job Summary page in GitHub Actions UI
func (a *Actions) WriteSummary(data []byte) error {
	if a.StepSummaryPath == "" {
		return nil
	}
	f, err := os.OpenFile(a.StepSummaryPath, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return fmt.Errorf("opening GITHUB_STEP_SUMMARY: %w", err)
	}
	defer f.Close() //nolint:errcheck

	if _, err := f.Write(data); err != nil {
		return fmt.Errorf("writing to GITHUB_STEP_SUMMARY: %w", err)
	}
	return nil
}

// SetOutput sets an output parameter for the current step
func (a *Actions) SetOutput(name, value string) error {
	if a.OutputPath == "" {
		return nil
	}
	f, err := os.OpenFile(a.OutputPath, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return fmt.Errorf("opening GITHUB_OUTPUT: %w", err)
	}
	defer f.Close() //nolint:errcheck

	if _, err := fmt.Fprintf(f, "%s=%s\n", name, value); err != nil {
		return fmt.Errorf("writing to GITHUB_OUTPUT: %w", err)
	}
	return nil
}
