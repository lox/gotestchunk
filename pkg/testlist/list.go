package testlist

import (
	"bufio"
	"fmt"
	"os/exec"
	"sort"
	"strings"
)

// Test represents a discovered test
type Test struct {
	Package string
	Name    string
}

func (t Test) String() string {
	return t.Package + "." + t.Name
}

// ModuleName returns the name of the module, e.g. github.com/lox/gotestchunk
func ModuleName() (string, error) {
	modCmd := exec.Command("go", "list", "-m")
	modOutput, err := modCmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get module name: %w", err)
	}
	return strings.TrimSpace(string(modOutput)), nil
}

// List returns all tests in the given package path
func List(pkgPath ...string) ([]Test, error) {
	// Get module name first
	moduleName, err := ModuleName()
	if err != nil {
		return nil, fmt.Errorf("failed to get module name: %w", err)
	}

	args := []string{"list"}
	args = append(args, pkgPath...)

	// Use go list to get all packages matching the pattern
	listCmd := exec.Command("go", args...)
	listCmd.Dir = "."
	output, err := listCmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("failed to list packages: %s", output)
	}

	var allTests []Test
	scanner := bufio.NewScanner(strings.NewReader(string(output)))

	// For each package, list its tests
	for scanner.Scan() {
		pkg := scanner.Text()

		// Get all top-level tests using the full package path
		cmd := exec.Command("go", "test", "-list", ".", pkg)
		cmd.Dir = "."
		output, err := cmd.CombinedOutput()
		if err != nil {
			return nil, fmt.Errorf("failed to list tests for package %s: %s", pkg, output)
		}

		// Get relative package path by removing module prefix
		relPkg := strings.TrimPrefix(pkg, moduleName+"/")

		testScanner := bufio.NewScanner(strings.NewReader(string(output)))
		for testScanner.Scan() {
			testName := testScanner.Text()
			// Only include Test functions, skip empty lines and other patterns
			if strings.HasPrefix(testName, "Test") {
				allTests = append(allTests, Test{
					Package: relPkg,
					Name:    testName,
				})
			}
		}
	}

	return allTests, nil
}

// Sort sorts a slice of tests by package name and test name
func Sort(tests []Test) {
	sort.Slice(tests, func(i, j int) bool {
		// First compare packages
		if tests[i].Package != tests[j].Package {
			return tests[i].Package < tests[j].Package
		}
		// If packages are equal, compare test names
		return tests[i].Name < tests[j].Name
	})
}

// Packages returns a sorted slice of unique package names from the given tests
func Packages(tests []Test) []string {
	// Use a map to track unique packages
	pkgMap := make(map[string]struct{})
	for _, test := range tests {
		pkgMap[test.Package] = struct{}{}
	}

	// Convert map keys to slice
	packages := make([]string, 0, len(pkgMap))
	for pkg := range pkgMap {
		packages = append(packages, pkg)
	}

	// Sort packages for consistent output
	sort.Strings(packages)
	return packages
}
