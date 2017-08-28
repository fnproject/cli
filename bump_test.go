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
	ff, err := loadFuncfile()
	if err != nil {
		return err
	}
	if ff.Version != version {
		return fmt.Errorf("funcfile version %v does not match expected version %v", ff.Version, version)
	}
	return nil
}
