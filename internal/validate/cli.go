package validate

import (
	"fmt"
	"os"
)

// CLI validates each directory argument, prints a report, and returns
// the process exit code: 0 when no directory has errors (warnings
// pass — CI for the default repository may hold a stricter line), 1
// otherwise.
func CLI(dirs []string) int {
	if len(dirs) == 0 {
		fmt.Fprintln(os.Stderr, "usage: orven validate <plugin-dir> [<plugin-dir>...]")
		return 2
	}
	exit := 0
	for _, dir := range dirs {
		fmt.Printf("orven validate %s\n\n", dir)
		findings := Dir(dir)
		errors, warnings := 0, 0
		for _, f := range findings {
			if f.Severity == "ERROR" {
				errors++
			} else {
				warnings++
			}
			fmt.Printf("  %-5s  %s: %s\n", f.Severity, f.Where, f.Message)
			if f.Suggestion != "" {
				fmt.Printf("         suggestion: %q\n", f.Suggestion)
			}
		}
		if len(findings) == 0 {
			fmt.Println("  no findings")
		}
		verdict := "OK"
		if errors > 0 {
			verdict = "validation failed"
			exit = 1
		}
		fmt.Printf("\n  %d error(s), %d warning(s) — %s\n\n", errors, warnings, verdict)
	}
	return exit
}
