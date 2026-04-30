package cmd

import (
	"encoding/json"
	"testing"

	"github.com/princepal9120/testgen-cli/internal/app"
)

func TestBuildCapabilitiesResponseIsAgentReadable(t *testing.T) {
	resp := buildCapabilitiesResponse()

	if resp.APIVersion != app.APIVersion {
		t.Fatalf("expected API version %q, got %q", app.APIVersion, resp.APIVersion)
	}
	if !resp.Success {
		t.Fatal("expected success response")
	}
	if resp.SchemaVersion == "" {
		t.Fatal("expected schema version")
	}
	if len(resp.Languages) == 0 {
		t.Fatal("expected language metadata")
	}
	if len(resp.Commands) == 0 {
		t.Fatal("expected command metadata")
	}
	if !resp.DryRunSupported || !resp.PatchSupported {
		t.Fatal("expected dry-run patch support in manifest")
	}

	seenDoctor := false
	seenGenerate := false
	for _, command := range resp.Commands {
		switch command.Name {
		case "doctor":
			seenDoctor = true
		case "generate":
			seenGenerate = true
			if !command.WritesFiles {
				t.Fatal("generate should be marked as file-writing capable")
			}
		}
	}
	if !seenDoctor || !seenGenerate {
		t.Fatalf("expected doctor and generate commands, got %#v", resp.Commands)
	}

	encoded, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("marshal capabilities response: %v", err)
	}
	var decoded capabilitiesResponse
	if err := json.Unmarshal(encoded, &decoded); err != nil {
		t.Fatalf("decode capabilities response: %v", err)
	}
}
