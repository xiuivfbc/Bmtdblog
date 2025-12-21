package system

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/pkg/errors"
)

func mysqlRestore(sqlContent string, dsn string) error {
	host, port, username, password, database, err := parseMySQLDSN(dsn)
	if err != nil {
		return err
	}
	cmd := exec.Command("mysql",
		"-h", host,
		"-P", port,
		"-u", username,
		fmt.Sprintf("-p%s", password),
		database,
	)
	cmd.Stdin = strings.NewReader(sqlContent)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return errors.Wrapf(err, "mysql restore failed: %s", string(output))
	}
	return nil
}
