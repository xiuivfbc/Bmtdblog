package system

import (
	"regexp"

	"github.com/pkg/errors"
)

func parseMySQLDSN(dsn string) (host, port, username, password, database string, err error) {
	var (
		regex *regexp.Regexp
		match []string
	)
	// 解析 DSN 的正则表达式
	regex = regexp.MustCompile(`^([^:]+):([^@]+)@tcp\(([^:]+):(\d+)\)\/([^?]+)(\?.*)?$`)
	match = regex.FindStringSubmatch(dsn)
	if len(match) < 6 {
		err = errors.New("invalid MySQL DSN format")
		return
	}
	username = match[1]
	password = match[2]
	host = match[3]
	port = match[4]
	database = match[5]
	return
}
