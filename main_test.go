package main

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path"
	"strconv"
	"testing"

	"github.com/BurntSushi/toml"
)

type testCase struct {
	Name   *string
	Source string
	Stdout *string
	Stderr *string
	Js     string
}

func TestWooo(t *testing.T) {
	var testCasesMap map[string][]testCase
	_, err := toml.DecodeFile("testdata/testcase.toml", &testCasesMap)
	if err != nil {
		t.Fatalf("open test cases file: %v", err)
	}

	testCases := testCasesMap["test"]

	tempDirName := t.TempDir()

	for i, tc := range testCases {
		name := strconv.Itoa(i)
		if tc.Name != nil {
			name = fmt.Sprintf("%d. %s", i, *tc.Name)
		}

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			watFilePath, ok := execWooo(t, i, tempDirName, tc.Source)
			if !ok {
				return
			}

			wasmFilePath, ok := execWatToWasm(t, i, watFilePath)
			if !ok {
				return
			}

			jsCmd := exec.CommandContext(t.Context(), "node", path.Join("testdata", tc.Js), wasmFilePath)

			stdoutR, stdoutW := io.Pipe()
			stderrR, stderrW := io.Pipe()

			jsCmd.Stdout = stdoutW
			jsCmd.Stderr = stderrW

			go func() {
				err := jsCmd.Run()
				if err != nil {
					t.Errorf("run wasm2js cmd: %v", err)
				}
				stdoutW.Close()
				stderrW.Close()
			}()

			stdout, err := io.ReadAll(stdoutR)
			if err != nil {
				t.Errorf("read stdout: %v", err)
			}
			if tc.Stdout != nil && string(stdout) != *tc.Stdout {
				t.Errorf("stdout: \n"+"expected:\n%s\n"+"actual:\n%s\n", *tc.Stdout, string(stdout))
			}

			stderr, err := io.ReadAll(stderrR)
			if err != nil {
				t.Errorf("read stderr: %v", err)
			}
			if tc.Stderr != nil && string(stderr) != *tc.Stderr {
				t.Errorf("stderr: \n"+"expected:\n%s\n"+"actual:\n%s\n", *tc.Stderr, string(stderr))
			}
		})
	}

}

func execWooo(t *testing.T, testNumber int, tempDirName string, source string) (string, bool) {
	t.Helper()

	testGoFilePath := path.Join(tempDirName, fmt.Sprintf("%d.go", testNumber))

	watFilePath := path.Join(tempDirName, fmt.Sprintf("%d.wat", testNumber))
	watFile, err := os.Create(watFilePath)
	if err != nil {
		t.Fatalf("open wat file for write: %v", err)
	}
	defer watFile.Close()

	err = CompileBytes(watFile, testGoFilePath, []byte(source), false)
	if err != nil {
		t.Errorf("unexpected compile error: %v", err)
	}

	return watFilePath, true
}

func execWatToWasm(t *testing.T, testNumber int, watFilePath string) (string, bool) {
	t.Helper()

	wasmFilePath := path.Join(path.Dir(watFilePath), fmt.Sprintf("%d.wasm", testNumber))
	watCmd := exec.CommandContext(t.Context(), "wat2wasm", watFilePath, "-o", wasmFilePath)
	err := watCmd.Run()
	if err != nil {
		t.Errorf("run wat2wasm cmd: %v", err)
		return "", false
	}

	return wasmFilePath, true
}
