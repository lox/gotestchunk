package commands

import (
	"bufio"
	"fmt"
	"os/exec"
	"strings"
)

type ListCmd struct {
	Packages []string `arg:"" optional:"" help:"Packages to list tests from" type:"path"`
}

func (cmd *ListCmd) Run() error {
	// Default to current directory if no package specified
	pkgPath := "."
	if len(cmd.Packages) > 0 {
		pkgPath = cmd.Packages[0]
	}

	tests, err := listTests(pkgPath)
	if err != nil {
		return fmt.Errorf("error listing tests: %w", err)
	}

	// Print tests sorted by package
	for _, test := range tests {
		fmt.Println(test)
	}
	return nil
}

func listTests(pkgPath string) ([]string, error) {
	// Get module name first
	modCmd := exec.Command("go", "list", "-m")
	modOutput, err := modCmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get module name: %w", err)
	}
	moduleName := strings.TrimSpace(string(modOutput))

	// Use go list to get all packages matching the pattern
	listCmd := exec.Command("go", "list", pkgPath)
	listOutput, err := listCmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to list packages: %w", err)
	}

	var allTests []string
	scanner := bufio.NewScanner(strings.NewReader(string(listOutput)))

	// For each package, list its tests
	for scanner.Scan() {
		pkg := scanner.Text()
		// Get all top-level tests
		cmd := exec.Command("go", "test", "-list", ".", pkg)
		output, err := cmd.Output()
		if err != nil {
			return nil, fmt.Errorf("failed to list tests for package %s: %w", pkg, err)
		}

		// Get relative package path by removing module prefix
		relPkg := strings.TrimPrefix(pkg, moduleName+"/")

		testScanner := bufio.NewScanner(strings.NewReader(string(output)))
		for testScanner.Scan() {
			testName := testScanner.Text()
			// Only include Test functions, skip empty lines and other patterns
			if strings.HasPrefix(testName, "Test") {
				// Prefix test names with relative package path if not in root package
				if relPkg != "." {
					testName = relPkg + "." + testName
				}
				allTests = append(allTests, testName)
			}
		}
	}

	return allTests, nil
}
