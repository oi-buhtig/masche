package memsearch

import (
	"fmt"
	"github.com/mozilla/masche/memaccess"
	"os"
	"os/exec"
	"testing"
)

var needle []byte = []byte("Find This!")

var buffersToFind = [][]byte{
	[]byte{0xc, 0xa, 0xf, 0xe},
	[]byte{0xd, 0xe, 0xa, 0xd, 0xb, 0xe, 0xe, 0xf},
	[]byte{0xb, 0xe, 0xb, 0xe, 0xf, 0xe, 0x0},
}

var notPresent = []byte("this string should generate a list of bytes not present in the process")

func printSoftErrors(softerrors []error) {
	for _, err := range softerrors {
		fmt.Fprintln(os.Stderr, err.Error())
	}
}

// launchProcess starts a process and waits until it writes something to stdout: that way we know it has been initialized.
// the process launched should write to stdout once it has been fully initialized.
func launchProcess(file string) (*exec.Cmd, error) {
	cmd := exec.Command(file)

	childout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	defer childout.Close()

	if err := cmd.Start(); err != nil {
		return nil, err
	}

	// Wait until the process writes something to stdout, so we know it has initialized all its memory.
	if read, err := childout.Read(make([]byte, 1)); err != nil || read != 1 {
		closeProcess(cmd)
		return nil, err
	}

	return cmd, nil
}

func closeProcess(cmd *exec.Cmd) {
	cmd.Process.Kill()
}

func TestSearchInOtherProcess(t *testing.T) {
	//TODO(mvanotti): Right now the command is hardcoded. We should decide how to fix this.
	cmd, err := launchProcess("../test/tools/test.exe")
	if err != nil {
		t.Fatal(err)
	}
	defer closeProcess(cmd)

	pid := uint(cmd.Process.Pid)
	reader, err, softerrors := memaccess.NewProcessMemoryReader(pid)
	printSoftErrors(softerrors)
	if err != nil {
		t.Fatal(err)
	}
	defer reader.Close()

	for i, buf := range buffersToFind {
		found, _, err, softerrors := FindNext(reader, 0, buf)
		printSoftErrors(softerrors)
		if err != nil {
			t.Fatal(err)
		} else if !found {
			t.Fatalf("memoryGrep failed for case %d, the following buffer should be found: %+v", i, buf)
		}
	}

	// This must not be present
	found, _, err, softerrors := FindNext(reader, 0, notPresent)
	printSoftErrors(softerrors)
	if err != nil {
		t.Fatal(err)
	} else if found {
		t.Fatalf("memoryGrep failed, it found a sequense of bytes that it shouldn't")
	}
}
