package mysql

import (
	"bufio"
	"io"
	"regexp"
	"strings"

	"github.com/pkg/errors"
)

// BinlogEventType is the enumeration of binlog event types.
type BinlogEventType int

const (
	// UnknownEventType represents other types of event that are ignored.
	UnknownEventType BinlogEventType = iota
	// WriteRowsEventType is the binlog event for INSERT.
	WriteRowsEventType
	// UpdateRowsEventType is the binlog event for UPDATE.
	UpdateRowsEventType
	// DeleteRowsEventType is the binlog event for DELETE.
	DeleteRowsEventType
	// QueryEventType is the binlog event for QUERY.
	// The thread ID is parsed from QUERY events.
	QueryEventType
	// XidEventType is the binlog event for Xid.
	// It is the last event of a transaction.
	XidEventType
)

// BinlogEvent contains the raw string of a single binlog event from the mysqlbinlog output stream.
type BinlogEvent struct {
	Type   BinlogEventType
	Header string
	Body   string
}

// BinlogTransaction is a list of events, starting with Query (BEGIN) and ending with Xid (COMMIT).
type BinlogTransaction []BinlogEvent

// ParseBinlogStream splits the mysqlbinlog output stream to a list of transactions.
func ParseBinlogStream(stream io.Reader) ([]BinlogTransaction, error) {
	reader := bufio.NewReader(stream)
	var event BinlogEvent
	var txns []BinlogTransaction
	var txn BinlogTransaction
	var bodyBuf strings.Builder
	seenEvent := false
	for {
		line, err := reader.ReadString('\n')
		if err != nil && err != io.EOF {
			return nil, errors.Wrap(err, "failed to read line from the stream")
		}

		switch {
		case len(line) == 0 && err == io.EOF:
			// The last line is empty. Skip the state machine.
		case !seenEvent && !strings.HasPrefix(line, "# at "):
			// Skip the first non-binlog-event lines output of mysqlbinlog.
			continue
		case strings.HasPrefix(line, "# at "):
			seenEvent = true
		case strings.Contains(line, "server id"):
			// Parse the header line.
			// Examples:
			// - Query:       #221020 15:45:58 server id 1  end_log_pos 2828 CRC32 0x5445bc77 	Query	thread_id=62592	exec_time=0	error_code=0
			// - Write_rows:  #221017 14:25:24 server id 1  end_log_pos 1916 CRC32 0x896854fc 	Write_rows: table id 259 flags: STMT_END_F
			// - Update_rows: #221018 16:21:19 server id 1  end_log_pos 2044 CRC32 0x9dbbb766 	Update_rows: table id 259 flags: STMT_END_F
			// - Delete_rows: #221017 14:31:53 server id 1  end_log_pos 1685 CRC32 0x5ea4b2c4 	Delete_rows: table id 259 flags: STMT_END_F
			// - Xid:         #221026 15:35:51 server id 1  end_log_pos 1435 CRC32 0x3be8594f 	Xid = 46
			event.Type = getEventType(line)
			event.Header = line
			continue
		default:
			// Accumulate the body.
			_, _ = bodyBuf.WriteString(line)
			continue
		}

		if event.Type != UnknownEventType {
			event.Body = bodyBuf.String()
			if len(txn) == 0 {
				txn = append(txn, event)
			} else if event.Type == QueryEventType && txn[0].Type == QueryEventType {
				// A Query event without a corresponding Xid event is not a start of a transaction.
				// We should replace the existing Query event with the new one.
				txn[0] = event
			} else if event.Type == XidEventType {
				// The current transaction ends with an Xid event, which means it's a complete transaction.
				txn = append(txn, event)
				txns = append(txns, txn)
				txn = nil
			} else {
				// This is a DML event. Append it to the current transaction.
				txn = append(txn, event)
			}
		}
		event = BinlogEvent{}
		bodyBuf.Reset()
		if err == io.EOF {
			if len(txn) > 0 {
				txns = append(txns, txn)
			}
			break
		}
	}

	return txns, nil
}

// FilterBinlogTransactionsByThreadID filters the binlog transaction by thread ID.
func FilterBinlogTransactionsByThreadID(txnList []BinlogTransaction, threadID string) ([]BinlogTransaction, error) {
	var ret []BinlogTransaction
	for _, txn := range txnList {
		event := txn[0]
		if event.Type != QueryEventType {
			return nil, errors.Errorf("invalid binlog transaction, the first event must be an query event, but got %s", event.Type.String())
		}
		parsed, err := parseQueryEventThreadID(event.Header)
		if err != nil {
			return nil, errors.WithMessage(err, "failed to parse thread ID from query event")
		}
		if parsed == threadID {
			ret = append(ret, txn)
		}
	}
	return ret, nil
}

var (
	reThreadID = regexp.MustCompile(`thread_id=(\d+)`)
)

func parseQueryEventThreadID(header string) (string, error) {
	matches := reThreadID.FindStringSubmatch(header)
	if len(matches) != 2 {
		return "", errors.Errorf("invalid query header %q", header)
	}
	return matches[1], nil
}

func getEventType(header string) BinlogEventType {
	if strings.Contains(header, "Query") {
		return QueryEventType
	} else if strings.Contains(header, "Xid") {
		return XidEventType
	} else if strings.Contains(header, "Write_rows") {
		return WriteRowsEventType
	} else if strings.Contains(header, "Update_rows") {
		return UpdateRowsEventType
	} else if strings.Contains(header, "Delete_rows") {
		return DeleteRowsEventType
	} else {
		// Ignore other events.
		return UnknownEventType
	}
}
