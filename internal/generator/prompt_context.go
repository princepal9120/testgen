package generator

import (
	"strings"
)

func augmentPromptWithTestContext(prompt string, framework string, ctx existingTestContext) string {
	var b strings.Builder
	b.WriteString(strings.TrimSpace(prompt))
	b.WriteString("\n\nProduction repo requirements:\n")
	b.WriteString("- Match the project's existing test framework, file layout, imports, naming, fixtures, mocks, and assertion style.\n")
	b.WriteString("- Test observable behavior and edge cases. Avoid brittle implementation-detail assertions unless the existing suite already does that.\n")
	b.WriteString("- Reuse existing helpers and factories only when they appear in the provided context. Do not invent unavailable helpers.\n")
	b.WriteString("- If the source is already covered by an existing test file, add compatible tests for the missing behavior instead of replacing the suite.\n")
	b.WriteString("- Output only test code. Do not include explanations, markdown, or shell commands.\n")
	if strings.TrimSpace(framework) != "" {
		b.WriteString("- Preferred detected framework: ")
		b.WriteString(strings.TrimSpace(framework))
		b.WriteString(".\n")
	}
	if len(ctx.Paths) > 0 && strings.TrimSpace(ctx.Snippet) != "" {
		b.WriteString("\nExisting test files considered:\n")
		for _, path := range ctx.Paths {
			b.WriteString("- ")
			b.WriteString(path)
			b.WriteString("\n")
		}
		b.WriteString("\nExisting test style context. Follow this style closely:\n")
		b.WriteString(ctx.Snippet)
		b.WriteString("\n")
	} else {
		b.WriteString("\nNo existing test file was found near this source file. Create the smallest idiomatic test file that a production team would accept.\n")
	}
	return b.String()
}
