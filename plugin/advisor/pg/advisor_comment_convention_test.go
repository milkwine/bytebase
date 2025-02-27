package pg

// Framework code is generated by the generator.

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/bytebase/bytebase/plugin/advisor"
)

func TestCommentConvention(t *testing.T) {
	tests := []advisor.TestCase{
		{
			Statement: `comment on column public.name is 'is a column name'`,
			Want: []advisor.Advice{
				{
					Status: advisor.Success,
					Code:   advisor.Ok,
					Title:  "OK",
					Line:   0,
				},
			},
		},
		{
			Statement: `comment on column public.name is 'is a  to long column name'`,
			Want: []advisor.Advice{
				{
					Status:  advisor.Warn,
					Code:    advisor.CommentTooLong,
					Title:   "comment.length",
					Content: "The length of comment should be within 20 characters",
					Line:    1,
				},
			},
		},
		{
			Statement: `comment on table public is 'is a table name'`,
			Want: []advisor.Advice{
				{
					Status: advisor.Success,
					Code:   advisor.Ok,
					Title:  "OK",
					Line:   0,
				},
			},
		},
		{
			Statement: `comment on table public is 'is a  to long table name'`,
			Want: []advisor.Advice{
				{
					Status:  advisor.Warn,
					Code:    advisor.CommentTooLong,
					Title:   "comment.length",
					Content: "The length of comment should be within 20 characters",
					Line:    1,
				},
			},
		},
	}

	payload, err := json.Marshal(advisor.CommentConventionRulePayload{
		Required:  true,
		MaxLength: 20,
	})
	require.NoError(t, err)
	advisor.RunSQLReviewRuleTests(t, tests, &CommentConventionAdvisor{}, &advisor.SQLReviewRule{
		Type:    advisor.SchemaRuleCommentLength,
		Level:   advisor.SchemaRuleLevelWarning,
		Payload: string(payload),
	}, advisor.MockPostgreSQLDatabase)
}
