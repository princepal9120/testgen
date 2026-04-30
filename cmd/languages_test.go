package cmd

import (
	"encoding/json"
	"testing"

	"github.com/princepal9120/testgen-cli/internal/scanner"
)

func TestSupportedLanguageManifestIncludesAllLanguages(t *testing.T) {
	resp := languagesResponse{
		APIVersion: "v1",
		Success:    true,
		Languages:  supportedLanguageInfo(),
	}
	resp.Count = len(resp.Languages)

	encoded, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("marshal languages response: %v", err)
	}
	var decoded languagesResponse
	if err := json.Unmarshal(encoded, &decoded); err != nil {
		t.Fatalf("decode languages json: %v", err)
	}

	if !decoded.Success {
		t.Fatal("expected success response")
	}
	if decoded.Count < 10 {
		t.Fatalf("expected at least 10 language entries, got %d", decoded.Count)
	}

	seen := map[string]languageInfo{}
	for _, lang := range decoded.Languages {
		seen[lang.Language] = lang
	}
	for _, required := range []string{
		scanner.LangJavaScript,
		scanner.LangTypeScript,
		scanner.LangPython,
		scanner.LangGo,
		scanner.LangRust,
		scanner.LangJava,
		scanner.LangCSharp,
		scanner.LangPHP,
		scanner.LangRuby,
		scanner.LangCPP,
		scanner.LangKotlin,
	} {
		info, ok := seen[required]
		if !ok {
			t.Fatalf("missing language %s in manifest", required)
		}
		if len(info.Extensions) == 0 {
			t.Fatalf("expected extensions for %s", required)
		}
		if info.DefaultFramework == "" || len(info.Frameworks) == 0 {
			t.Fatalf("expected framework metadata for %s", required)
		}
	}
}

func TestGetExtensionsForLanguageSortsAndNormalizesAliases(t *testing.T) {
	exts := scanner.GetExtensionsForLanguage("c++")
	if len(exts) == 0 {
		t.Fatal("expected C++ extensions")
	}
	for i := 1; i < len(exts); i++ {
		if exts[i-1] > exts[i] {
			t.Fatalf("expected sorted extensions, got %#v", exts)
		}
	}
}
