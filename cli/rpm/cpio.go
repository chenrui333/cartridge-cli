package rpm

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/tarantool/cartridge-cli/cli/context"
)

func packCpio(relPaths []string, resFileName string, ctx *context.Ctx) error {
	filesBuffer := bytes.Buffer{}
	filesBuffer.WriteString(strings.Join(relPaths, "\n"))

	cpioFile, err := os.Create(resFileName)
	if err != nil {
		return err
	}
	defer cpioFile.Close()

	cpioFileWriter := bufio.NewWriter(cpioFile)
	defer cpioFileWriter.Flush()

	var stderrBuf bytes.Buffer

	cmd := exec.Command("cpio", "-o", "-H", "newc")
	cmd.Stdin = &filesBuffer
	cmd.Stdout = cpioFileWriter
	cmd.Stderr = &stderrBuf
	cmd.Dir = ctx.Pack.PackageFilesDir

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("Failed to run \n%s\n\nStderr: %s", cmd.String(), stderrBuf.String())
	}

	return nil
}
