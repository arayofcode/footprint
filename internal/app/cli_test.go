package app

import (
	"context"
	"strings"
	"testing"
)

func TestRunCLI_MissingUsername(t *testing.T) {
	t.Setenv("GITHUB_ACTOR", "")
	t.Setenv("GITHUB_TOKEN", "token")

	err := RunCLI(context.Background(), CLIConfig{
		Username: "",
	})

	if err == nil {
		t.Fatalf("expected error for missing username")
	}
	if !strings.Contains(err.Error(), "username is required") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRunCLI_MissingToken(t *testing.T) {
	t.Setenv("GITHUB_TOKEN", "")

	err := RunCLI(context.Background(), CLIConfig{
		Username: "ray",
	})

	if err == nil {
		t.Fatalf("expected error for missing token")
	}
	if !strings.Contains(err.Error(), "GITHUB_TOKEN is required") {
		t.Fatalf("unexpected error: %v", err)
	}
}
