package mysql

// Framework code is generated by the generator.

import (
	"testing"

	"github.com/bytebase/bytebase/plugin/advisor"
)

func TestDisallowLimit(t *testing.T) {
	tests := []advisor.TestCase{
		{
			Statement: ``,
			Want: []advisor.Advice{
				{
					Status:  advisor.Success,
					Code:    advisor.Ok,
					Title:   "OK",
					Content: "",
				},
			},
		},
		{
			Statement: `INSERT INTO tech_book SELECT * FROM tech_book LIMIT 1`,
			Want: []advisor.Advice{
				{
					Status:  advisor.Warn,
					Code:    advisor.InsertUseLimit,
					Title:   "statement.disallow-limit",
					Content: "LIMIT clause is forbidden in INSERT, UPDATE and DELETE statement, but \"INSERT INTO tech_book SELECT * FROM tech_book LIMIT 1\" uses",
					Line:    1,
				},
			},
		},
		{
			Statement: `INSERT INTO tech_book SELECT * FROM tech_book UNION SELECT * FROM tech_book LIMIT 1`,
			Want: []advisor.Advice{
				{
					Status:  advisor.Warn,
					Code:    advisor.InsertUseLimit,
					Title:   "statement.disallow-limit",
					Content: "LIMIT clause is forbidden in INSERT, UPDATE and DELETE statement, but \"INSERT INTO tech_book SELECT * FROM tech_book UNION SELECT * FROM tech_book LIMIT 1\" uses",
					Line:    1,
				},
			},
		},
		{
			Statement: `UPDATE tech_book SET name = 'my name' LIMIT 10`,
			Want: []advisor.Advice{
				{
					Status:  advisor.Warn,
					Code:    advisor.UpdateUseLimit,
					Title:   "statement.disallow-limit",
					Content: "LIMIT clause is forbidden in INSERT, UPDATE and DELETE statement, but \"UPDATE tech_book SET name = 'my name' LIMIT 10\" uses",
					Line:    1,
				},
			},
		},
		{
			Statement: `DELETE FROM tech_book LIMIT 10`,
			Want: []advisor.Advice{
				{
					Status:  advisor.Warn,
					Code:    advisor.DeleteUseLimit,
					Title:   "statement.disallow-limit",
					Content: "LIMIT clause is forbidden in INSERT, UPDATE and DELETE statement, but \"DELETE FROM tech_book LIMIT 10\" uses",
					Line:    1,
				},
			},
		},
	}

	advisor.RunSQLReviewRuleTests(t, tests, &DisallowLimitAdvisor{}, &advisor.SQLReviewRule{
		Type:    advisor.SchemaRuleStatementDisallowLimit,
		Level:   advisor.SchemaRuleLevelWarning,
		Payload: "",
	}, advisor.MockMySQLDatabase)
}
