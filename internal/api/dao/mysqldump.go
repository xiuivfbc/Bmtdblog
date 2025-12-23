package dao

import (
	"bytes"
	"os/exec"

	"github.com/pkg/errors"
)

func Mysqldump(host, port, username, password, database string) (output []byte, err error) {
	var (
		cmd    *exec.Cmd
		buf    bytes.Buffer
		errBuf bytes.Buffer
	)
	// 构建 mysqldump 命令
	args := []string{
		"-h", host,
		"-P", port,
		"-u", username,
		"--password=" + password,
		"--default-character-set=utf8mb4",
		"--single-transaction",
		"--add-drop-table",
		"--quick",
		database,
	}
	cmd = exec.Command("mysqldump", args...)
	cmd.Stdout = &buf
	cmd.Stderr = &errBuf
	err = cmd.Run()
	if err != nil {
		err = errors.WithMessage(err, errBuf.String())
		return
	}
	output = buf.Bytes()
	return
}
