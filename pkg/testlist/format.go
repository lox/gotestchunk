package testlist

import (
	"fmt"
	"strings"
)

// Format returns a formatted string representation of tests based on the format type
func Format(tests []Test, format string) (string, error) {
	switch format {
	case "listTests":
		// Simply join test names with newlines
		var names []string
		for _, test := range tests {
			names = append(names, test.Name)
		}
		return strings.Join(names, "\n"), nil

	case "listPackages":
		// Join package paths with newlines
		var paths []string
		for _, pkg := range Packages(tests) {
			paths = append(paths, "./"+pkg)
		}
		return strings.Join(paths, "\n"), nil

	case "runPattern":
		// Create go test -run pattern
		var testNames []string
		for _, test := range tests {
			parts := strings.Split(test.String(), ".")
			if len(parts) == 2 {
				testNames = append(testNames, parts[1])
			}
		}
		if len(testNames) == 0 {
			return "", nil
		}
		return fmt.Sprintf("^(%s)$", strings.Join(testNames, "|")), nil

	default:
		return "", fmt.Errorf("unknown format: %s", format)
	}
}
