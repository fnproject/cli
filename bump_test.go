package main

import (
	"fmt"
	"os/exec"
	"testing"
)

func TestBump(t *testing.T) {
	tmp := cdToTmp(t)

	// create a func.yaml
	res, err := exec.Command(fnTestBin, "init", "--runtime", "go").CombinedOutput()
	if err != nil {
		fmt.Println(string(res))
		t.Fatal(err)
	}
	err = verifyVersion(tmp, "0.0.1")
	if err != nil {
		t.Fatal(err)
	}

	res, err = exec.Command(fnTestBin, "bump").CombinedOutput()
	if err != nil {
		fmt.Println(string(res))
		t.Fatal(err)
	}
	err = verifyVersion(tmp, "0.0.2")
	if err != nil {
		t.Fatal(err)
	}
	res, err = exec.Command(fnTestBin, "bump", "--minor").CombinedOutput()
	if err != nil {
		fmt.Println(string(res))
		t.Fatal(err)
	}
	err = verifyVersion(tmp, "0.1.0")
	if err != nil {
		t.Fatal(err)
	}
	res, err = exec.Command(fnTestBin, "bump", "--major").CombinedOutput()
	if err != nil {
		fmt.Println(string(res))
		t.Fatal(err)
	}
	err = verifyVersion(tmp, "1.0.0")
	if err != nil {
		t.Fatal(err)
	}

}

func verifyVersion(tmp, version string) error {
	_, ff, err := loadFuncfile()
	if err != nil {
		return err
	}
	if ff.Version != version {
		return fmt.Errorf("funcfile version %v does not match expected version %v", ff.Version, version)
	}
	return nil
}

func TestCleanImageName(t *testing.T) {
	testCases := []struct {
		input    string
		expected string
	}{
		{"someimage:latest", "someimage"},
		{"repository/image/name:latest", "repository/image/name"},
		{"repository:port/image/name:latest", "repository:port/image/name"},
	}
	for _, c := range testCases {
		t.Run(c.input, func(t *testing.T) {
			output := cleanImageName(c.input)
			if output != c.expected {
				t.Fatalf("Expected '%s' but got '%s'", c.expected, output)
			}
		})
	}
}
