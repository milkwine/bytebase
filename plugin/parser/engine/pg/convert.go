package pg

import (
	"strconv"
	"strings"

	pgquery "github.com/pganalyze/pg_query_go/v2"
	"github.com/pkg/errors"

	"github.com/bytebase/bytebase/plugin/parser"
	"github.com/bytebase/bytebase/plugin/parser/ast"
)

// convert converts the pg_query.Node to ast.Node.
func convert(node *pgquery.Node, statement parser.SingleSQL) (res ast.Node, err error) {
	defer func() {
		if err == nil && res != nil {
			res.SetText(strings.TrimSpace(statement.Text))
			res.SetLastLine(statement.LastLine)
			switch n := res.(type) {
			case *ast.CreateTableStmt:
				err = parser.SetLineForCreateTableStmt(parser.Postgres, n)
			case *ast.AlterTableStmt:
				for _, item := range n.AlterItemList {
					item.SetLastLine(n.LastLine())
				}
			}
		}
	}()
	switch in := node.Node.(type) {
	case *pgquery.Node_AlterTableStmt:
		alterTable := &ast.AlterTableStmt{
			Table:         convertRangeVarToTableName(in.AlterTableStmt.Relation, ast.TableTypeBaseTable),
			AlterItemList: []ast.Node{},
		}
		for _, cmd := range in.AlterTableStmt.Cmds {
			if cmdNode, ok := cmd.Node.(*pgquery.Node_AlterTableCmd); ok {
				alterCmd := cmdNode.AlterTableCmd

				switch alterCmd.Subtype {
				case pgquery.AlterTableType_AT_AddColumn:
					def, ok := alterCmd.Def.Node.(*pgquery.Node_ColumnDef)
					if !ok {
						return nil, parser.NewConvertErrorf("expected ColumnDef but found %t", alterCmd.Def.Node)
					}

					column, err := convertColumnDef(def)
					if err != nil {
						return nil, err
					}

					addColumn := &ast.AddColumnListStmt{
						Table:      alterTable.Table,
						ColumnList: []*ast.ColumnDef{column},
					}

					alterTable.AlterItemList = append(alterTable.AlterItemList, addColumn)
				case pgquery.AlterTableType_AT_DropColumn:
					dropColumn := &ast.DropColumnStmt{
						Table:      alterTable.Table,
						ColumnName: alterCmd.Name,
					}

					alterTable.AlterItemList = append(alterTable.AlterItemList, dropColumn)
				case pgquery.AlterTableType_AT_AddConstraint:
					def, ok := alterCmd.Def.Node.(*pgquery.Node_Constraint)
					if !ok {
						return nil, parser.NewConvertErrorf("expected Constraint but found %t", alterCmd.Def.Node)
					}
					constraint, err := convertConstraint(def)
					if err != nil {
						return nil, err
					}

					addConstraint := &ast.AddConstraintStmt{
						Table:      alterTable.Table,
						Constraint: constraint,
					}

					alterTable.AlterItemList = append(alterTable.AlterItemList, addConstraint)
				case pgquery.AlterTableType_AT_DropConstraint:
					dropConstraint := &ast.DropConstraintStmt{
						Table:          alterTable.Table,
						ConstraintName: alterCmd.Name,
					}

					alterTable.AlterItemList = append(alterTable.AlterItemList, dropConstraint)
				case pgquery.AlterTableType_AT_SetNotNull:
					setNotNull := &ast.SetNotNullStmt{
						Table:      alterTable.Table,
						ColumnName: alterCmd.Name,
					}

					alterTable.AlterItemList = append(alterTable.AlterItemList, setNotNull)
				case pgquery.AlterTableType_AT_DropNotNull:
					dropNotNull := &ast.DropNotNullStmt{
						Table:      alterTable.Table,
						ColumnName: alterCmd.Name,
					}

					alterTable.AlterItemList = append(alterTable.AlterItemList, dropNotNull)
				case pgquery.AlterTableType_AT_AlterColumnType:
					column, ok := alterCmd.Def.Node.(*pgquery.Node_ColumnDef)
					if !ok {
						return nil, parser.NewConvertErrorf("expected ColumnDef but found %t", alterCmd.Def.Node)
					}
					alterColumType := &ast.AlterColumnTypeStmt{
						Table:      alterTable.Table,
						ColumnName: alterCmd.Name,
						Type:       convertDataType(column.ColumnDef.TypeName),
					}

					alterTable.AlterItemList = append(alterTable.AlterItemList, alterColumType)
				case pgquery.AlterTableType_AT_ColumnDefault:
					if alterCmd.Def == nil {
						dropDefault := &ast.DropDefaultStmt{
							Table:      alterTable.Table,
							ColumnName: alterCmd.Name,
						}

						alterTable.AlterItemList = append(alterTable.AlterItemList, dropDefault)
					} else {
						var err error
						setDefault := &ast.SetDefaultStmt{
							Table:      alterTable.Table,
							ColumnName: alterCmd.Name,
						}
						if setDefault.Expression, _, _, err = convertExpressionNode(alterCmd.Def); err != nil {
							return nil, err
						}
						text, err := pgquery.DeparseNode(pgquery.DeparseTypeExpr, alterCmd.Def)
						if err != nil {
							return nil, err
						}
						setDefault.Expression.SetText(text)

						alterTable.AlterItemList = append(alterTable.AlterItemList, setDefault)
					}
				}
			}
		}
		return alterTable, nil
	case *pgquery.Node_CreateStmt:
		return convertCreateStmt(in.CreateStmt)
	case *pgquery.Node_RenameStmt:
		switch in.RenameStmt.RenameType {
		case pgquery.ObjectType_OBJECT_COLUMN:
			tableType, err := convertToTableType(in.RenameStmt.RelationType)
			if err != nil {
				return nil, err
			}
			table := convertRangeVarToTableName(in.RenameStmt.Relation, tableType)
			return &ast.AlterTableStmt{
				Table: table,
				AlterItemList: []ast.Node{
					&ast.RenameColumnStmt{
						Table:      table,
						ColumnName: in.RenameStmt.Subname,
						NewName:    in.RenameStmt.Newname,
					},
				},
			}, nil
		case pgquery.ObjectType_OBJECT_TABLE:
			table := convertRangeVarToTableName(in.RenameStmt.Relation, ast.TableTypeBaseTable)
			return &ast.AlterTableStmt{
				Table: table,
				AlterItemList: []ast.Node{
					&ast.RenameTableStmt{
						Table:   table,
						NewName: in.RenameStmt.Newname,
					},
				},
			}, nil
		case pgquery.ObjectType_OBJECT_TABCONSTRAINT:
			table := convertRangeVarToTableName(in.RenameStmt.Relation, ast.TableTypeBaseTable)
			return &ast.AlterTableStmt{
				Table: table,
				AlterItemList: []ast.Node{
					&ast.RenameConstraintStmt{
						Table:          table,
						ConstraintName: in.RenameStmt.Subname,
						NewName:        in.RenameStmt.Newname,
					},
				},
			}, nil
		case pgquery.ObjectType_OBJECT_VIEW:
			view := convertRangeVarToTableName(in.RenameStmt.Relation, ast.TableTypeView)
			return &ast.AlterTableStmt{
				Table: view,
				AlterItemList: []ast.Node{
					&ast.RenameTableStmt{
						Table:   view,
						NewName: in.RenameStmt.Newname,
					},
				},
			}, nil
		case pgquery.ObjectType_OBJECT_INDEX:
			return &ast.RenameIndexStmt{
				Table:     convertRangeVarToIndexTableName(in.RenameStmt.Relation, ast.TableTypeUnknown),
				IndexName: in.RenameStmt.Relation.Relname,
				NewName:   in.RenameStmt.Newname,
			}, nil
		}
	case *pgquery.Node_IndexStmt:
		indexDef := &ast.IndexDef{
			Table:  convertRangeVarToTableName(in.IndexStmt.Relation, ast.TableTypeUnknown),
			Name:   in.IndexStmt.Idxname,
			Unique: in.IndexStmt.Unique,
		}

		for _, key := range in.IndexStmt.IndexParams {
			index, ok := key.Node.(*pgquery.Node_IndexElem)
			if !ok {
				return nil, parser.NewConvertErrorf("expected IndexElem but found %t", key.Node)
			}
			// We only support index on columns now.
			// TODO(rebelice): support index on expressions.
			if index.IndexElem.Name != "" {
				indexDef.KeyList = append(indexDef.KeyList, &ast.IndexKeyDef{
					Type: ast.IndexKeyTypeColumn,
					Key:  index.IndexElem.Name,
				})
			} else {
				indexDef.KeyList = append(indexDef.KeyList, &ast.IndexKeyDef{
					Type: ast.IndexKeyTypeExpression,
				})
			}
		}

		return &ast.CreateIndexStmt{Index: indexDef}, nil
	case *pgquery.Node_DropStmt:
		switch in.DropStmt.RemoveType {
		case pgquery.ObjectType_OBJECT_INDEX:
			dropIndex := &ast.DropIndexStmt{}
			for _, object := range in.DropStmt.Objects {
				list, ok := object.Node.(*pgquery.Node_List)
				if !ok {
					return nil, parser.NewConvertErrorf("expected List but found %t", object.Node)
				}
				indexDef, err := convertListToIndexDef(list, ast.TableTypeUnknown)
				if err != nil {
					return nil, err
				}
				dropIndex.IndexList = append(dropIndex.IndexList, indexDef)
			}
			return dropIndex, nil
		case pgquery.ObjectType_OBJECT_TABLE:
			dropTable := &ast.DropTableStmt{}
			for _, object := range in.DropStmt.Objects {
				list, ok := object.Node.(*pgquery.Node_List)
				if !ok {
					return nil, parser.NewConvertErrorf("expected List but found %t", object.Node)
				}
				tableDef, err := convertListToTableDef(list, ast.TableTypeBaseTable)
				if err != nil {
					return nil, err
				}
				dropTable.TableList = append(dropTable.TableList, tableDef)
			}
			return dropTable, nil
		case pgquery.ObjectType_OBJECT_VIEW:
			dropView := &ast.DropTableStmt{}
			for _, object := range in.DropStmt.Objects {
				list, ok := object.Node.(*pgquery.Node_List)
				if !ok {
					return nil, parser.NewConvertErrorf("expected List but found %t", object.Node)
				}
				viewDef, err := convertListToTableDef(list, ast.TableTypeView)
				if err != nil {
					return nil, err
				}
				dropView.TableList = append(dropView.TableList, viewDef)
			}
			return dropView, nil
		case pgquery.ObjectType_OBJECT_SCHEMA:
			dropSchema := &ast.DropSchemaStmt{
				IfExists: in.DropStmt.MissingOk,
			}
			switch in.DropStmt.Behavior {
			case pgquery.DropBehavior_DROP_CASCADE:
				dropSchema.Behavior = ast.DropSchemaBehaviorCascade
			case pgquery.DropBehavior_DROP_RESTRICT:
				dropSchema.Behavior = ast.DropSchemaBehaviorRestrict
			default:
				dropSchema.Behavior = ast.DropSchemaBehaviorNone
			}
			for _, object := range in.DropStmt.Objects {
				strNode, ok := object.Node.(*pgquery.Node_String_)
				if !ok {
					return nil, parser.NewConvertErrorf("expected String but found %t", object.Node)
				}
				dropSchema.SchemaList = append(dropSchema.SchemaList, strNode.String_.Str)
			}
			return dropSchema, nil
		}
	case *pgquery.Node_DropdbStmt:
		return &ast.DropDatabaseStmt{
			DatabaseName: in.DropdbStmt.Dbname,
			IfExists:     in.DropdbStmt.MissingOk,
		}, nil
	case *pgquery.Node_SelectStmt:
		return convertSelectStmt(in.SelectStmt)
	case *pgquery.Node_UpdateStmt:
		update := &ast.UpdateStmt{
			Table: convertRangeVarToTableName(in.UpdateStmt.Relation, ast.TableTypeBaseTable),
		}
		// Convert FROM clause
		// Here we only find the SELECT stmt in FROM clause
		for _, item := range in.UpdateStmt.FromClause {
			if node, ok := item.Node.(*pgquery.Node_RangeSubselect); ok {
				subselect, err := convertRangeSubselect(node.RangeSubselect)
				if err != nil {
					return nil, err
				}
				update.SubqueryList = append(update.SubqueryList, subselect)
			}
		}
		// Convert WHERE clause
		if in.UpdateStmt.WhereClause != nil {
			var err error
			var subqueryList []*ast.SubqueryDef
			update.WhereClause, update.PatternLikeList, subqueryList, err = convertExpressionNode(in.UpdateStmt.WhereClause)
			if err != nil {
				return nil, err
			}
			update.SubqueryList = append(update.SubqueryList, subqueryList...)
		}
		return update, nil
	case *pgquery.Node_DeleteStmt:
		deleteStmt := &ast.DeleteStmt{
			Table: convertRangeVarToTableName(in.DeleteStmt.Relation, ast.TableTypeBaseTable),
		}
		if in.DeleteStmt.WhereClause != nil {
			var err error
			if deleteStmt.WhereClause, deleteStmt.PatternLikeList, deleteStmt.SubqueryList, err = convertExpressionNode(in.DeleteStmt.WhereClause); err != nil {
				return nil, err
			}
		}
		return deleteStmt, nil
	case *pgquery.Node_AlterObjectSchemaStmt:
		switch in.AlterObjectSchemaStmt.ObjectType {
		case pgquery.ObjectType_OBJECT_TABLE:
			table := convertRangeVarToTableName(in.AlterObjectSchemaStmt.Relation, ast.TableTypeBaseTable)
			return &ast.AlterTableStmt{
				Table: table,
				AlterItemList: []ast.Node{
					&ast.SetSchemaStmt{
						Table:     table,
						NewSchema: in.AlterObjectSchemaStmt.Newschema,
					},
				},
			}, nil
		case pgquery.ObjectType_OBJECT_VIEW:
			view := convertRangeVarToTableName(in.AlterObjectSchemaStmt.Relation, ast.TableTypeView)
			return &ast.AlterTableStmt{
				Table: view,
				AlterItemList: []ast.Node{
					&ast.SetSchemaStmt{
						Table:     view,
						NewSchema: in.AlterObjectSchemaStmt.Newschema,
					},
				},
			}, nil
		}
	case *pgquery.Node_ExplainStmt:
		explainStmt := &ast.ExplainStmt{}
		if query, ok := in.ExplainStmt.Query.Node.(*pgquery.Node_SelectStmt); ok {
			if explainStmt.Statement, err = convertSelectStmt(query.SelectStmt); err != nil {
				return nil, err
			}
		}
		return explainStmt, nil
	case *pgquery.Node_InsertStmt:
		insertStmt := &ast.InsertStmt{
			Table: convertRangeVarToTableName(in.InsertStmt.Relation, ast.TableTypeBaseTable),
		}

		if in.InsertStmt.SelectStmt != nil {
			if selectNode, ok := in.InsertStmt.SelectStmt.Node.(*pgquery.Node_SelectStmt); ok {
				// PG parser will parse the value list as a SELECT statement.
				if selectNode.SelectStmt.ValuesLists == nil {
					if insertStmt.Select, err = convertSelectStmt(selectNode.SelectStmt); err != nil {
						return nil, err
					}
				} else {
					for _, list := range selectNode.SelectStmt.ValuesLists {
						var valueList []ast.ExpressionNode
						listNode, ok := list.Node.(*pgquery.Node_List)
						if !ok {
							return nil, parser.NewConvertErrorf("expected Node_List but found %t", list.Node)
						}
						for _, item := range listNode.List.Items {
							value, _, _, err := convertExpressionNode(item)
							if err != nil {
								return nil, err
							}
							valueList = append(valueList, value)
						}
						insertStmt.ValueList = append(insertStmt.ValueList, valueList)
					}
				}
			} else {
				return nil, parser.NewConvertErrorf("expected SelectStmt but found %t", in.InsertStmt.SelectStmt.Node)
			}
		}
		return insertStmt, nil
	case *pgquery.Node_CopyStmt:
		copyStmt := ast.CopyStmt{
			Table:    convertRangeVarToTableName(in.CopyStmt.Relation, ast.TableTypeBaseTable),
			FilePath: in.CopyStmt.Filename,
		}

		return &copyStmt, nil
	case *pgquery.Node_CommentStmt:
		commentStmt := ast.CommentStmt{
			Comment: in.CommentStmt.Comment,
		}

		return &commentStmt, nil
	case *pgquery.Node_CreatedbStmt:
		createDatabaseStmt := ast.CreateDatabaseStmt{
			Name: in.CreatedbStmt.Dbname,
		}
		for _, option := range in.CreatedbStmt.Options {
			if item, ok := option.Node.(*pgquery.Node_DefElem); ok && item.DefElem.Defname == "encoding" {
				value, ok := item.DefElem.Arg.Node.(*pgquery.Node_String_)
				if !ok {
					return nil, parser.NewConvertErrorf("expected String but found %t", item.DefElem.Arg.Node)
				}
				createDatabaseStmt.OptionList = append(createDatabaseStmt.OptionList, &ast.DatabaseOptionDef{
					Type:  ast.DatabaseOptionEncoding,
					Value: value.String_.Str,
				})
			}
		}

		return &createDatabaseStmt, nil
	case *pgquery.Node_CreateSchemaStmt:
		createSchemaStmt := ast.CreateSchemaStmt{
			Name:        in.CreateSchemaStmt.Schemaname,
			IfNotExists: in.CreateSchemaStmt.IfNotExists,
		}
		roleSpec, err := convertRoleSpec(in.CreateSchemaStmt.Authrole)
		if err != nil {
			return nil, err
		}
		createSchemaStmt.RoleSpec = roleSpec
		for _, elt := range in.CreateSchemaStmt.SchemaElts {
			switch stmt := elt.Node.(type) {
			// Currently, only CREATE TABLE, CREATE VIEW, CREATE INDEX, CREATE SEQUENCE,
			// CREATE TRIGGER and GRANT are accepted as clauses within CREATE SCHEMA.
			// TODO(zp): support other statement list above.
			case *pgquery.Node_CreateStmt:
				createStmt, err := convertCreateStmt(stmt.CreateStmt)
				if err != nil {
					return nil, err
				}
				createSchemaStmt.SchemaElementList = append(createSchemaStmt.SchemaElementList, createStmt)
			default:
			}
		}
		return &createSchemaStmt, nil
	default:
		return &ast.UnconvertedStmt{}, nil
	}

	return nil, nil
}

func convertRoleSpec(in *pgquery.RoleSpec) (*ast.RoleSpec, error) {
	if in == nil {
		return nil, nil
	}
	switch in.Roletype {
	case pgquery.RoleSpecType_ROLE_SPEC_TYPE_UNDEFINED:
		return &ast.RoleSpec{
			Type:  ast.RoleSpecTypeNone,
			Value: "",
		}, nil
	case pgquery.RoleSpecType_ROLESPEC_CSTRING:
		return &ast.RoleSpec{
			Type:  ast.RoleSpecTypeUser,
			Value: in.Rolename,
		}, nil
	case pgquery.RoleSpecType_ROLESPEC_CURRENT_USER:
		return &ast.RoleSpec{
			Type:  ast.RoleSpecTypeCurrentUser,
			Value: "",
		}, nil
	case pgquery.RoleSpecType_ROLESPEC_SESSION_USER:
		return &ast.RoleSpec{
			Type:  ast.RoleSpecTypeSessionUser,
			Value: "",
		}, nil
	}
	return nil, parser.NewConvertErrorf("unexpected role spec type: %q", in.Roletype.String())
}

func convertExpressionNode(node *pgquery.Node) (ast.ExpressionNode, []*ast.PatternLikeDef, []*ast.SubqueryDef, error) {
	if node == nil || node.Node == nil {
		return &ast.UnconvertedExpressionDef{}, nil, nil, nil
	}
	switch in := node.Node.(type) {
	case *pgquery.Node_AConst:
		return convertExpressionNode(in.AConst.Val)
	case *pgquery.Node_String_:
		return &ast.StringDef{Value: in.String_.Str}, nil, nil, nil
	case *pgquery.Node_ResTarget:
		return convertExpressionNode(in.ResTarget.Val)
	case *pgquery.Node_TypeCast:
		_, likeList, subqueryList, err := convertExpressionNode(in.TypeCast.Arg)
		if err != nil {
			return nil, nil, nil, err
		}
		return &ast.UnconvertedExpressionDef{}, likeList, subqueryList, nil
	case *pgquery.Node_ColumnRef:
		columnName := &ast.ColumnNameDef{Table: &ast.TableDef{}}
		list := in.ColumnRef.Fields
		// There are three cases for column name:
		//   1. schemaName.tableName.columnName
		//   2. tableName.columnName
		//   3. columnName
		// The pg parser will split them by ".", and use a list to define it.
		// So we need to consider this three cases.
		switch len(in.ColumnRef.Fields) {
		// schemaName.tableName.columName
		case 3:
			schema, ok := list[0].Node.(*pgquery.Node_String_)
			if !ok {
				return nil, nil, nil, parser.NewConvertErrorf("expected String but found %t", in.ColumnRef.Fields[2].Node)
			}
			columnName.Table.Schema = schema.String_.Str
			// need to convert tableName.columnName
			list = list[1:]
			fallthrough
		// tableName.columnName
		case 2:
			table, ok := list[0].Node.(*pgquery.Node_String_)
			if !ok {
				return nil, nil, nil, parser.NewConvertErrorf("expected String but found %t", in.ColumnRef.Fields[1].Node)
			}
			columnName.Table.Name = table.String_.Str
			// need to convert columnName
			list = list[1:]
			fallthrough
		// columnName
		case 1:
			switch column := list[0].Node.(type) {
			// column name
			case *pgquery.Node_String_:
				columnName.ColumnName = column.String_.Str
			// e.g. SELECT * FROM t;
			case *pgquery.Node_AStar:
				columnName.ColumnName = "*"
			default:
				return nil, nil, nil, parser.NewConvertErrorf("expected String or AStar but found %t", in.ColumnRef.Fields[0].Node)
			}
		default:
			return nil, nil, nil, parser.NewConvertErrorf("failed to convert ColumnRef, column name contains unexpected components: %v", in)
		}
		return columnName, nil, nil, nil
	case *pgquery.Node_FuncCall:
		var likeList []*ast.PatternLikeDef
		var subqueryList []*ast.SubqueryDef
		for _, arg := range in.FuncCall.Args {
			_, interLike, interSubquery, err := convertExpressionNode(arg)
			if err != nil {
				return nil, nil, nil, err
			}
			likeList = append(likeList, interLike...)
			subqueryList = append(subqueryList, interSubquery...)
		}
		return &ast.UnconvertedExpressionDef{}, likeList, subqueryList, nil
	case *pgquery.Node_AExpr:
		var likeList, interLike []*ast.PatternLikeDef
		var subqueryList, interSubquery []*ast.SubqueryDef
		var lExpr, rExpr ast.ExpressionNode
		var err error
		if in.AExpr.Lexpr != nil {
			if lExpr, interLike, interSubquery, err = convertExpressionNode(in.AExpr.Lexpr); err != nil {
				return nil, nil, nil, err
			}
			likeList = append(likeList, interLike...)
			subqueryList = append(subqueryList, interSubquery...)
		}
		if in.AExpr.Rexpr != nil {
			if rExpr, interLike, interSubquery, err = convertExpressionNode(in.AExpr.Rexpr); err != nil {
				return nil, nil, nil, err
			}
			likeList = append(likeList, interLike...)
			subqueryList = append(subqueryList, interSubquery...)
		}
		if len(in.AExpr.Name) == 1 {
			name, ok := in.AExpr.Name[0].Node.(*pgquery.Node_String_)
			if !ok {
				return nil, nil, nil, parser.NewConvertErrorf("expected String but found %t", in.AExpr.Name[0].Node)
			}
			switch name.String_.Str {
			// LIKE
			case operatorLike, operatorNotLike:
				like := &ast.PatternLikeDef{
					Not:        (name.String_.Str == operatorNotLike),
					Expression: lExpr,
					Pattern:    rExpr,
				}
				likeList = append(likeList, like)
				return like, likeList, interSubquery, nil
			}
		}
		return &ast.UnconvertedExpressionDef{}, likeList, subqueryList, nil
	case *pgquery.Node_BoolExpr:
		var likeList []*ast.PatternLikeDef
		var subqueryList []*ast.SubqueryDef
		for _, arg := range in.BoolExpr.Args {
			_, interLike, interSubquery, err := convertExpressionNode(arg)
			if err != nil {
				return nil, nil, nil, err
			}
			likeList = append(likeList, interLike...)
			subqueryList = append(subqueryList, interSubquery...)
		}
		return &ast.UnconvertedExpressionDef{}, likeList, subqueryList, nil
	case *pgquery.Node_SubLink:
		if subselectNode, ok := in.SubLink.Subselect.Node.(*pgquery.Node_SelectStmt); ok {
			subselect, err := convertSelectStmt(subselectNode.SelectStmt)
			if err != nil {
				return nil, nil, nil, err
			}
			subQuery := &ast.SubqueryDef{Select: subselect}
			return subQuery, nil, []*ast.SubqueryDef{subQuery}, nil
		}
	}
	return &ast.UnconvertedExpressionDef{}, nil, nil, nil
}

// convertCreateStmt convert pgquery create stmt to Bytebase create table stmt node.
func convertCreateStmt(in *pgquery.CreateStmt) (*ast.CreateTableStmt, error) {
	table := &ast.CreateTableStmt{
		IfNotExists: in.IfNotExists,
		Name:        convertRangeVarToTableName(in.Relation, ast.TableTypeBaseTable),
	}

	for _, elt := range in.TableElts {
		switch item := elt.Node.(type) {
		case *pgquery.Node_ColumnDef:
			column, err := convertColumnDef(item)
			if err != nil {
				return nil, err
			}
			table.ColumnList = append(table.ColumnList, column)
		case *pgquery.Node_Constraint:
			cons, err := convertConstraint(item)
			if err != nil {
				return nil, err
			}
			table.ConstraintList = append(table.ConstraintList, cons)
		}
	}
	return table, nil
}

func convertSelectStmt(in *pgquery.SelectStmt) (*ast.SelectStmt, error) {
	selectStmt := &ast.SelectStmt{}

	setOperation, err := convertSetOperation(in.Op)
	if err != nil {
		return nil, err
	}

	selectStmt.SetOperation = setOperation
	if setOperation != ast.SetOperationTypeNone {
		lQuery, err := convertSelectStmt(in.Larg)
		if err != nil {
			return nil, err
		}
		rQuery, err := convertSelectStmt(in.Rarg)
		if err != nil {
			return nil, err
		}
		selectStmt.LQuery = lQuery
		selectStmt.RQuery = rQuery
		return selectStmt, nil
	}

	// Convert target list
	for _, node := range in.TargetList {
		convertedNode, _, _, err := convertExpressionNode(node)
		if err != nil {
			return nil, err
		}
		selectStmt.FieldList = append(selectStmt.FieldList, convertedNode)
	}
	// Convert FROM clause
	// Here we only find the SELECT stmt in FROM clause
	for _, item := range in.FromClause {
		if node, ok := item.Node.(*pgquery.Node_RangeSubselect); ok {
			subselect, err := convertRangeSubselect(node.RangeSubselect)
			if err != nil {
				return nil, err
			}
			selectStmt.SubqueryList = append(selectStmt.SubqueryList, subselect)
		}
	}
	// Convert WHERE clause
	if in.WhereClause != nil {
		var err error
		var subqueryList []*ast.SubqueryDef
		selectStmt.WhereClause, selectStmt.PatternLikeList, subqueryList, err = convertExpressionNode(in.WhereClause)
		if err != nil {
			return nil, err
		}
		selectStmt.SubqueryList = append(selectStmt.SubqueryList, subqueryList...)
	}
	return selectStmt, nil
}

func convertRangeSubselect(node *pgquery.RangeSubselect) (*ast.SubqueryDef, error) {
	subselect, ok := node.Subquery.Node.(*pgquery.Node_SelectStmt)
	if !ok {
		return nil, parser.NewConvertErrorf("expected SELECT but found %t", node.Subquery.Node)
	}
	res, err := convertSelectStmt(subselect.SelectStmt)
	if err != nil {
		return nil, err
	}
	return &ast.SubqueryDef{Select: res}, nil
}

func convertSetOperation(t pgquery.SetOperation) (ast.SetOperationType, error) {
	switch t {
	case pgquery.SetOperation_SETOP_NONE:
		return ast.SetOperationTypeNone, nil
	case pgquery.SetOperation_SETOP_UNION:
		return ast.SetOperationTypeUnion, nil
	case pgquery.SetOperation_SETOP_INTERSECT:
		return ast.SetOperationTypeIntersect, nil
	case pgquery.SetOperation_SETOP_EXCEPT:
		return ast.SetOperationTypeExcept, nil
	default:
		return 0, errors.Errorf("failed to parse set operation: unknown type %s", t)
	}
}

func convertListToTableDef(in *pgquery.Node_List, tableType ast.TableType) (*ast.TableDef, error) {
	stringList, err := convertListToStringList(in)
	if err != nil {
		return nil, err
	}
	switch len(in.List.Items) {
	case 2:
		return &ast.TableDef{
			Type:   tableType,
			Schema: stringList[0],
			Name:   stringList[1],
		}, nil
	case 1:
		return &ast.TableDef{
			Type: tableType,
			Name: stringList[0],
		}, nil
	default:
		return nil, parser.NewConvertErrorf("expected length is 1 or 2, but found %d", len(in.List.Items))
	}
}

func convertListToIndexDef(in *pgquery.Node_List, tableType ast.TableType) (*ast.IndexDef, error) {
	stringList, err := convertListToStringList(in)
	if err != nil {
		return nil, err
	}
	indexDef := &ast.IndexDef{}
	switch len(in.List.Items) {
	case 2:
		indexDef.Table = &ast.TableDef{
			Type:   tableType,
			Schema: stringList[0],
		}
		indexDef.Name = stringList[1]
	case 1:
		indexDef.Name = stringList[0]
	default:
		return nil, parser.NewConvertErrorf("expected length is 1 or 2, but found %d", len(in.List.Items))
	}
	return indexDef, nil
}

func convertListToStringList(in *pgquery.Node_List) ([]string, error) {
	var res []string
	for _, item := range in.List.Items {
		s, ok := item.Node.(*pgquery.Node_String_)
		if !ok {
			return nil, parser.NewConvertErrorf("expected String but found %t", item.Node)
		}
		res = append(res, s.String_.Str)
	}
	return res, nil
}

func convertRangeVarToTableName(in *pgquery.RangeVar, tableType ast.TableType) *ast.TableDef {
	return &ast.TableDef{
		Type:     tableType,
		Database: in.Catalogname,
		Schema:   in.Schemaname,
		Name:     in.Relname,
	}
}

func convertRangeVarToIndexTableName(in *pgquery.RangeVar, tableType ast.TableType) *ast.TableDef {
	return &ast.TableDef{
		Type:   tableType,
		Schema: in.Schemaname,
	}
}

func convertConstraint(in *pgquery.Node_Constraint) (*ast.ConstraintDef, error) {
	cons := &ast.ConstraintDef{
		Name:           in.Constraint.Conname,
		Type:           convertConstraintType(in.Constraint.Contype, in.Constraint.Indexname != ""),
		SkipValidation: in.Constraint.SkipValidation,
		Deferrable:     in.Constraint.Deferrable,
		Initdeferred:   in.Constraint.Initdeferred,
	}

	switch cons.Type {
	case ast.ConstraintTypePrimary, ast.ConstraintTypeUnique:
		for _, key := range in.Constraint.Keys {
			name, ok := key.Node.(*pgquery.Node_String_)
			if !ok {
				return nil, parser.NewConvertErrorf("expected String but found %t", key.Node)
			}
			cons.KeyList = append(cons.KeyList, name.String_.Str)
		}
		for _, col := range in.Constraint.Including {
			name, ok := col.Node.(*pgquery.Node_String_)
			if !ok {
				return nil, parser.NewConvertErrorf("expected String but found %t", col.Node)
			}
			cons.Including = append(cons.Including, name.String_.Str)
		}
		cons.IndexTableSpace = in.Constraint.Indexspace
	case ast.ConstraintTypeForeign:
		cons.Foreign = &ast.ForeignDef{
			Table: convertRangeVarToTableName(in.Constraint.Pktable, ast.TableTypeBaseTable),
		}

		for _, item := range in.Constraint.PkAttrs {
			name, ok := item.Node.(*pgquery.Node_String_)
			if !ok {
				return nil, parser.NewConvertErrorf("expected String but found %t", item.Node)
			}
			cons.Foreign.ColumnList = append(cons.Foreign.ColumnList, name.String_.Str)
		}

		for _, item := range in.Constraint.FkAttrs {
			name, ok := item.Node.(*pgquery.Node_String_)
			if !ok {
				return nil, parser.NewConvertErrorf("expected String but found %t", item.Node)
			}
			cons.KeyList = append(cons.KeyList, name.String_.Str)
		}
	case ast.ConstraintTypePrimaryUsingIndex, ast.ConstraintTypeUniqueUsingIndex:
		cons.IndexName = in.Constraint.Indexname
	case ast.ConstraintTypeCheck, ast.ConstraintTypeDefault:
		expression, _, _, err := convertExpressionNode(in.Constraint.RawExpr)
		if err != nil {
			return nil, err
		}
		text, err := pgquery.DeparseNode(pgquery.DeparseTypeExpr, in.Constraint.RawExpr)
		if err != nil {
			return nil, err
		}
		expression.SetText(text)
		cons.Expression = expression
	}

	return cons, nil
}

func convertConstraintType(in pgquery.ConstrType, usingIndex bool) ast.ConstraintType {
	switch in {
	case pgquery.ConstrType_CONSTR_PRIMARY:
		if usingIndex {
			return ast.ConstraintTypePrimaryUsingIndex
		}
		return ast.ConstraintTypePrimary
	case pgquery.ConstrType_CONSTR_UNIQUE:
		if usingIndex {
			return ast.ConstraintTypeUniqueUsingIndex
		}
		return ast.ConstraintTypeUnique
	case pgquery.ConstrType_CONSTR_FOREIGN:
		if usingIndex {
			return ast.ConstraintTypeUndefined
		}
		return ast.ConstraintTypeForeign
	case pgquery.ConstrType_CONSTR_NOTNULL:
		return ast.ConstraintTypeNotNull
	case pgquery.ConstrType_CONSTR_CHECK:
		return ast.ConstraintTypeCheck
	case pgquery.ConstrType_CONSTR_DEFAULT:
		return ast.ConstraintTypeDefault
	}
	return ast.ConstraintTypeUndefined
}

func convertColumnDef(in *pgquery.Node_ColumnDef) (*ast.ColumnDef, error) {
	column := &ast.ColumnDef{
		ColumnName: in.ColumnDef.Colname,
	}
	columnType := convertDataType(in.ColumnDef.TypeName)
	column.Type = columnType

	for _, cons := range in.ColumnDef.Constraints {
		constraint, ok := cons.Node.(*pgquery.Node_Constraint)
		if !ok {
			return nil, parser.NewConvertErrorf("expected Constraint but found %t", cons.Node)
		}
		columnCons, err := convertConstraint(constraint)
		if err != nil {
			return nil, err
		}
		columnCons.KeyList = append(columnCons.KeyList, in.ColumnDef.Colname)
		column.ConstraintList = append(column.ConstraintList, columnCons)
	}

	return column, nil
}

func convertToTableType(relationType pgquery.ObjectType) (ast.TableType, error) {
	switch relationType {
	case pgquery.ObjectType_OBJECT_TABLE:
		return ast.TableTypeBaseTable, nil
	case pgquery.ObjectType_OBJECT_VIEW:
		return ast.TableTypeView, nil
	default:
		return ast.TableTypeUnknown, parser.NewConvertErrorf("expected TABLE or VIEW but found %s", relationType)
	}
}

func stripPgCatalogPrefix(tp *pgquery.TypeName) *pgquery.TypeName {
	// The built-in data type may have the "pg_catalog" prefix.
	if len(tp.Names) > 0 {
		if first, ok := tp.Names[0].Node.(*pgquery.Node_String_); ok && first.String_.Str == "pg_catalog" {
			tp.Names = tp.Names[1:]
		}
	}
	return tp
}

func convertDataType(tp *pgquery.TypeName) ast.DataType {
	tp = stripPgCatalogPrefix(tp)
	if len(tp.Names) == 1 {
		name, ok := tp.Names[0].Node.(*pgquery.Node_String_)
		if !ok {
			return &ast.UnconvertedDataType{}
		}
		s := name.String_.Str
		switch {
		case strings.HasPrefix(s, "int"):
			size, err := strconv.Atoi(s[3:])
			if err != nil {
				return convertToUnconvertedDataType(tp)
			}
			return &ast.Integer{Size: size}
		case strings.HasPrefix(s, "float"):
			size, err := strconv.Atoi(s[5:])
			if err != nil {
				return convertToUnconvertedDataType(tp)
			}
			return &ast.Float{Size: size}
		case s == "serial":
			return &ast.Serial{Size: 4}
		case s == "smallserial":
			return &ast.Serial{Size: 2}
		case s == "bigserial":
			return &ast.Serial{Size: 8}
		case strings.HasPrefix(s, "serial"):
			size, err := strconv.Atoi(s[6:])
			if err != nil {
				return convertToUnconvertedDataType(tp)
			}
			return &ast.Serial{Size: size}
		case s == "numeric":
			return convertToDecimal(tp.Typmods)
		case s == "bpchar":
			return convertToCharacter(tp.Typmods)
		case s == "varchar":
			return convertToVarchar(tp.Typmods)
		case s == "text":
			return &ast.Text{}
		}
	}
	return convertToUnconvertedDataType(tp)
}

func convertToUnconvertedDataType(tp *pgquery.TypeName) ast.DataType {
	res := &ast.UnconvertedDataType{}
	for _, name := range tp.Names {
		s, ok := name.Node.(*pgquery.Node_String_)
		if !ok {
			return &ast.UnconvertedDataType{}
		}
		res.Name = append(res.Name, s.String_.Str)
	}
	return res
}

func convertToDecimal(typmods []*pgquery.Node) ast.DataType {
	ok := false
	decimal := &ast.Decimal{}
	switch len(typmods) {
	case 0:
		return decimal
	case 1:
		if decimal.Precision, ok = convertToInteger(typmods[0]); !ok {
			return &ast.UnconvertedDataType{}
		}
		return decimal
	case 2:
		if decimal.Precision, ok = convertToInteger(typmods[0]); !ok {
			return &ast.UnconvertedDataType{}
		}
		if decimal.Scale, ok = convertToInteger(typmods[1]); !ok {
			return &ast.UnconvertedDataType{}
		}
		return decimal
	default:
		return &ast.UnconvertedDataType{}
	}
}

func convertToVarchar(typmods []*pgquery.Node) ast.DataType {
	if len(typmods) != 1 {
		return &ast.UnconvertedDataType{}
	}
	size, ok := convertToInteger(typmods[0])
	if !ok {
		return &ast.UnconvertedDataType{}
	}
	return &ast.CharacterVarying{Size: size}
}

func convertToCharacter(typmods []*pgquery.Node) ast.DataType {
	if len(typmods) != 1 {
		return &ast.UnconvertedDataType{}
	}
	size, ok := convertToInteger(typmods[0])
	if !ok {
		return &ast.UnconvertedDataType{}
	}
	return &ast.Character{Size: size}
}

func convertToInteger(in *pgquery.Node) (int, bool) {
	aConst, ok := in.Node.(*pgquery.Node_AConst)
	if !ok {
		return 0, false
	}
	integer, ok := aConst.AConst.Val.Node.(*pgquery.Node_Integer)
	if !ok {
		return 0, false
	}
	return int(integer.Integer.Ival), true
}
