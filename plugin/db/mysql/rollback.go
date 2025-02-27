package mysql

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/pkg/errors"
)

var (
	// The (?s) is a modifier that makes "." to match the new line character, which is a valid character in the MySQL database and name.
	reDatabaseTable = regexp.MustCompile("(?s)`(.+)`\\.`(.+)`")
)

// GetRollbackSQL generates the rollback SQL for the list of binlog events in the reversed order.
func (txn BinlogTransaction) GetRollbackSQL(tables map[string][]string) (string, error) {
	if len(txn) == 0 {
		return "", nil
	}
	var sqlList []string
	// Generate rollback SQL for each statement of the transaction in reversed order.
	// Each statement may have multiple affected rows in a single binlog event. The order between them is irrelevant.
	for i := len(txn) - 1; i >= 0; i-- {
		e := txn[i]
		if e.Type != WriteRowsEventType && e.Type != DeleteRowsEventType && e.Type != UpdateRowsEventType {
			continue
		}
		sql, err := e.getRollbackSQL(tables)
		if err != nil {
			return "", err
		}
		sqlList = append(sqlList, sql)
	}
	return strings.Join(sqlList, "\n"), nil
}

func (e *BinlogEvent) getRollbackSQL(tables map[string][]string) (string, error) {
	// 1. Remove the "### " prefix of each line.
	// mysqlbinlog output is separated by "\n", ref https://sourcegraph.com/github.com/mysql/mysql-server@a246bad76b9271cb4333634e954040a970222e0a/-/blob/sql/log_event.cc?L2398
	body := strings.Split(e.Body, "\n")
	body = replaceAllPrefix(body, "### ", "")

	matches := reDatabaseTable.FindStringSubmatch(e.Body)
	if len(matches) != 3 {
		return "", errors.Errorf("failed to match database and table names in binlog event %q", e.Body)
	}
	tableName := matches[2]
	columnNames, ok := tables[tableName]
	if !ok {
		return "", errors.Errorf("table %s does not exist in the provided table map", tableName)
	}

	// 2. Switch "DELETE FROM" and "INSERT INTO".
	// 3. Replace "WHERE" and "SET" with each other.
	// 4. Replace "@i" with the column names.
	// 5. Add a ";" at the end of each row.
	var err error
	switch e.Type {
	case WriteRowsEventType:
		body = replaceAllPrefix(body, "INSERT INTO", "DELETE FROM")
		body = replaceAllPrefix(body, "SET", "WHERE")
		body, err = replaceColumns(columnNames, body, "WHERE", " AND", ";")
	case DeleteRowsEventType:
		body = replaceAllPrefix(body, "DELETE FROM", "INSERT INTO")
		body = replaceAllPrefix(body, "WHERE", "SET")
		body, err = replaceColumns(columnNames, body, "SET", ",", ";")
	case UpdateRowsEventType:
		body = replaceAllPrefix(body, "WHERE", "OLDWHERE")
		body = replaceAllPrefix(body, "SET", "WHERE")
		body = replaceAllPrefix(body, "OLDWHERE", "SET")
		body, err = replaceColumns(columnNames, body, "SET", ",", "")
		if err != nil {
			return "", err
		}
		body, err = replaceColumns(columnNames, body, "WHERE", " AND", ";")
	default:
		return "", errors.Errorf("invalid binlog event type %s", e.Type.String())
	}

	return strings.Join(body, "\n"), err
}

func replaceAllPrefix(body []string, old, new string) []string {
	var ret []string
	for _, line := range body {
		ret = append(ret, replacePrefix(line, old, new))
	}
	return ret
}

func replacePrefix(line, old, new string) string {
	if strings.HasPrefix(line, old) {
		return new + strings.TrimPrefix(line, old)
	}
	return line
}

func replaceColumns(columnNames []string, body []string, sepLine, lineSuffix, sectionSuffix string) ([]string, error) {
	var ret []string
	for i := 0; i < len(body); {
		line := body[i]
		if line != sepLine {
			ret = append(ret, line)
			i++
			continue
		}
		// Found the "WHERE" or "SET" line
		ret = append(ret, line)
		i++
		if i+len(columnNames) > len(body) {
			return nil, errors.Errorf("binlog event body has a section with less columns than %d: %q", len(columnNames), strings.Join(body, "\n"))
		}
		for j := range columnNames {
			prefix := fmt.Sprintf("  @%d=", j+1)
			line := body[i+j]
			if !strings.HasPrefix(line, prefix) {
				return nil, errors.Errorf("invalid value line %q, must starts with %q", line, prefix)
			}
			if j == len(columnNames)-1 {
				ret = append(ret, fmt.Sprintf("  `%s`=%s%s", columnNames[j], strings.TrimPrefix(line, prefix), sectionSuffix))
			} else {
				ret = append(ret, fmt.Sprintf("  `%s`=%s%s", columnNames[j], strings.TrimPrefix(line, prefix), lineSuffix))
			}
		}
		i += len(columnNames)
	}
	return ret, nil
}

func (t BinlogEventType) String() string {
	switch t {
	case DeleteRowsEventType:
		return "DELETE"
	case UpdateRowsEventType:
		return "UPDATE"
	case WriteRowsEventType:
		return "INSERT"
	case QueryEventType:
		return "QUERY"
	case XidEventType:
		return "XID"
	default:
		return "UNKNOWN"
	}
}
