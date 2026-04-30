package adapters

import "regexp"

func NewCPPAdapter() LanguageAdapter {
	return &regexLanguageAdapter{
		BaseAdapter:     BaseAdapter{language: "cpp", frameworks: []string{"googletest", "catch2", "doctest"}, defaultFW: "googletest"},
		extensions:      []string{".cpp", ".cc", ".cxx", ".c++", ".hpp", ".hh", ".hxx"},
		classPattern:    regexp.MustCompile(`\b(?:class|struct)\s+(\w+)`),
		importPattern:   regexp.MustCompile(`^\s*#include\s+[<"]([^>"]+)[>"]`),
		methodPattern:   regexp.MustCompile(`^\s*(?:template\s*<[^>]+>\s*)?(?:inline\s+|static\s+|virtual\s+|constexpr\s+)*([\w:<>&*\s]+)\s+(?:\w+::)?(\w+)\s*\(([^)]*)\)`),
		functionPattern: regexp.MustCompile(`^\s*(?:template\s*<[^>]+>\s*)?(?:inline\s+|static\s+|constexpr\s+)*([\w:<>&*\s]+)\s+(\w+)\s*\(([^)]*)\)`),
		testSuffix:      "_test",
		testExt:         ".cpp",
		promptName:      "C++",
		promptHints:     []string{"Use GoogleTest TEST/TEST_F by default.", "Cover value semantics, boundary inputs, and error returns/exceptions."},
	}
}
