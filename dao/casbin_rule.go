// ==========================================================================
// Code generated and maintained by GoFrame CLI tool.
// ==========================================================================

package dao

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// casbinRuleDao is the data access object for the table t_casbin_rule.
// You can define custom methods on it to extend its functionality as needed.
type casbinRuleDao struct {
	*CasbinRuleDao
}

var (
	// CasbinRule is a globally accessible object for table t_casbin_rule operations.
	CasbinRule = casbinRuleDao{NewCasbinRuleDao()}
)

// CasbinRuleDao is the data access object for the table t_casbin_rule.
type CasbinRuleDao struct {
	table    string             // table is the underlying table name of the DAO.
	group    string             // group is the database configuration group name of the current DAO.
	columns  CasbinRuleColumns  // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler // handlers for customized model modification.
}

// CasbinRuleColumns defines and stores column names for the table t_casbin_rule.
type CasbinRuleColumns struct {
	Id    string //
	Ptype string //
	V0    string //
	V1    string //
	V2    string //
	V3    string //
	V4    string //
	V5    string //
}

// casbinRuleColumns holds the columns for the table t_casbin_rule.
var casbinRuleColumns = CasbinRuleColumns{
	Id:    "id",
	Ptype: "ptype",
	V0:    "v0",
	V1:    "v1",
	V2:    "v2",
	V3:    "v3",
	V4:    "v4",
	V5:    "v5",
}

// NewCasbinRuleDao creates and returns a new DAO object for table data access.
func NewCasbinRuleDao(handlers ...gdb.ModelHandler) *CasbinRuleDao {
	return &CasbinRuleDao{
		group:    "default",
		table:    "casbin_rule",
		columns:  casbinRuleColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *CasbinRuleDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *CasbinRuleDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *CasbinRuleDao) Columns() CasbinRuleColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *CasbinRuleDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *CasbinRuleDao) Ctx(ctx context.Context) *gdb.Model {
	model := dao.DB().Model(dao.table)
	for _, handler := range dao.handlers {
		model = handler(model)
	}
	return model.Safe().Ctx(ctx)
}

// Transaction wraps the transaction logic using function f.
// It rolls back the transaction and returns the error if function f returns a non-nil error.
// It commits the transaction and returns nil if function f returns nil.
//
// Note: Do not commit or roll back the transaction in function f,
// as it is automatically handled by this function.
func (dao *CasbinRuleDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
