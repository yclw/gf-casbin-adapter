package gfadapter

import (
	"context"
	"errors"
	"fmt"

	"github.com/yclw/gf-casbin-adapter/dao"
	"github.com/yclw/gf-casbin-adapter/model/entity"

	"github.com/casbin/casbin/v2/model"
	"github.com/casbin/casbin/v2/persist"
	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/util/gconv"
)

// Filter filter conditions
type Filter struct {
	Ptype []string
	V0    []string
	V1    []string
	V2    []string
	V3    []string
	V4    []string
	V5    []string
}

var (
	// Check if the persist.*Adapter interface is implemented

	// Adapter implementation
	_ persist.Adapter = new(Adapter)

	// BatchAdapter batch adapter
	_ persist.BatchAdapter = new(Adapter)

	// ContextAdapter context adapter
	_ persist.ContextAdapter = new(Adapter)

	// ContextBatchAdapter context batch adapter
	_ persist.ContextBatchAdapter = new(Adapter)

	// ContextFilteredAdapter context filtered adapter
	_ persist.ContextFilteredAdapter = new(Adapter)

	// ContextUpdatableAdapter context updatable adapter
	_ persist.ContextUpdatableAdapter = new(Adapter)

	// FilteredAdapter filtered adapter
	_ persist.FilteredAdapter = new(Adapter)

	// UpdatableAdapter updatable adapter
	_ persist.UpdatableAdapter = new(Adapter)
)

// Adapter represents the GoFrame adapter for policy storage.
type Adapter struct {
	dao        *dao.CasbinRuleDao
	isFiltered bool
}

func NewAdapter(isFiltered bool) *Adapter {
	return &Adapter{
		dao:        dao.NewCasbinRuleDao(),
		isFiltered: isFiltered,
	}
}

// NewAdapterWithName creates a new Adapter with a custom table name.
func NewAdapterWithName(tableName string, isFiltered bool) *Adapter {
	return &Adapter{
		dao:        dao.NewCasbinRuleDaoWithName(tableName),
		isFiltered: isFiltered,
	}
}

// NewAdapterWithNameAndColumns creates a new Adapter with a custom table name and columns.
func NewAdapterWithNameAndColumns(tableName string, columns dao.CasbinRuleColumns, isFiltered bool) *Adapter {
	return &Adapter{
		dao:        dao.NewCasbinRuleDaoWithNameAndColumns(tableName, columns),
		isFiltered: isFiltered,
	}
}

// LoadPolicy loads all policy rules from the storage.
func (a *Adapter) LoadPolicy(model model.Model) error {
	return a.LoadPolicyCtx(context.Background(), model)
}

// LoadPolicyCtx loads policy from database.
func (a *Adapter) LoadPolicyCtx(ctx context.Context, model model.Model) error {
	var lines []entity.CasbinRule
	cols := a.dao.Columns()
	if err := a.dao.Ctx(ctx).Order(cols.Id).Scan(&lines); err != nil {
		return err
	}
	err := a.preview(&lines, model)
	if err != nil {
		return err
	}
	for _, line := range lines {
		err := loadPolicyLine(line, model)
		if err != nil {
			return err
		}
	}
	return nil
}

// LoadFilteredPolicy loads only policy rules that match the filter.
func (a *Adapter) LoadFilteredPolicy(model model.Model, filter interface{}) error {
	return a.LoadFilteredPolicyCtx(context.Background(), model, filter)
}

// LoadFilteredPolicyCtx loads only policy rules that match the filter.
func (a *Adapter) LoadFilteredPolicyCtx(ctx context.Context, model model.Model, filter interface{}) error {
	filterValue, ok := filter.(Filter)
	if !ok {
		return errors.New("invalid filter type")
	}

	cols := a.dao.Columns()
	// Build query conditions
	qs := a.dao.Ctx(ctx)

	// Apply filter conditions
	a.applyFilter(qs, filterValue)

	// Execute query and sort by ID
	var lines []entity.CasbinRule
	if err := qs.Order(cols.Id).Scan(&lines); err != nil {
		return err
	}

	// Pre-check filter results
	err := a.preview(&lines, model)
	if err != nil {
		return err
	}

	// Process query results
	for _, line := range lines {
		err := loadPolicyLine(line, model)
		if err != nil {
			return err
		}
	}

	a.isFiltered = true
	return nil
}

// IsFiltered returns true if the loaded policy has been filtered.
func (a *Adapter) IsFiltered() bool {
	return a.isFiltered
}

// IsFilteredCtx returns true if the loaded policy has been filtered.
func (a *Adapter) IsFilteredCtx(ctx context.Context) bool {
	return a.isFiltered
}

// SavePolicy saves all policy rules to the storage.
func (a *Adapter) SavePolicy(model model.Model) error {
	return a.SavePolicyCtx(context.Background(), model)
}

// SavePolicyCtx saves policy to database.
func (a *Adapter) SavePolicyCtx(ctx context.Context, model model.Model) error {
	var err error

	tx, err := a.dao.DB().Begin(ctx)
	if err != nil {
		return err
	}

	err = a.truncateTable(ctx)

	if err != nil {
		tx.Rollback()
		return err
	}

	var lines []entity.CasbinRule
	flushEvery := 1000
	for ptype, ast := range model["p"] {
		for _, rule := range ast.Policy {
			lines = append(lines, a.savePolicyLine(ptype, rule))
			if len(lines) > flushEvery {
				if _, err := a.dao.Ctx(ctx).Data(lines).InsertIgnore(); err != nil {
					tx.Rollback()
					return err
				}
				lines = nil
			}
		}
	}

	for ptype, ast := range model["g"] {
		for _, rule := range ast.Policy {
			lines = append(lines, a.savePolicyLine(ptype, rule))
			if len(lines) > flushEvery {
				if _, err := a.dao.Ctx(ctx).Data(lines).InsertIgnore(); err != nil {
					tx.Rollback()
					return err
				}
				lines = nil
			}
		}
	}
	if len(lines) > 0 {
		if _, err := a.dao.Ctx(ctx).Data(lines).InsertIgnore(); err != nil {
			tx.Rollback()
			return err
		}
	}

	return tx.Commit()
}

// AddPolicy adds a policy rule to the storage.
// This is part of the Auto-Save feature.
func (a *Adapter) AddPolicy(sec string, ptype string, rule []string) error {
	return a.AddPolicyCtx(context.Background(), sec, ptype, rule)
}

// AddPolicyCtx adds a policy rule to the storage.
func (a *Adapter) AddPolicyCtx(ctx context.Context, sec string, ptype string, rule []string) error {
	line := a.savePolicyLine(ptype, rule)
	_, err := a.dao.Ctx(ctx).Data(line).InsertIgnore()
	return err
}

// AddPolicies adds policy rules to the storage.
// This is part of the Auto-Save feature.
func (a *Adapter) AddPolicies(sec string, ptype string, rules [][]string) error {
	return a.AddPoliciesCtx(context.Background(), sec, ptype, rules)
}

// AddPoliciesCtx adds policy rules to the storage.
// This is part of the Auto-Save feature.
func (a *Adapter) AddPoliciesCtx(ctx context.Context, sec string, ptype string, rules [][]string) error {
	var lines []entity.CasbinRule
	for _, rule := range rules {
		line := a.savePolicyLine(ptype, rule)
		lines = append(lines, line)
	}
	_, err := a.dao.Ctx(ctx).Data(lines).InsertIgnore()
	return err
}

// RemovePolicy removes a policy rule from the storage.
// This is part of the Auto-Save feature.
func (a *Adapter) RemovePolicy(sec string, ptype string, rule []string) error {
	return a.RemovePolicyCtx(context.Background(), sec, ptype, rule)
}

// RemovePolicyCtx removes a policy rule from the storage with context.
// This is part of the Auto-Save feature.
func (a *Adapter) RemovePolicyCtx(ctx context.Context, sec string, ptype string, rule []string) error {
	line := a.savePolicyLine(ptype, rule)
	_, err := a.dao.Ctx(ctx).Where(line).OmitEmpty().Delete()
	return err
}

// RemovePolicies removes policy rules from the storage.
// This is part of the Auto-Save feature.
func (a *Adapter) RemovePolicies(sec string, ptype string, rules [][]string) error {
	return a.RemovePoliciesCtx(context.Background(), sec, ptype, rules)
}

// RemovePoliciesCtx removes policy rules from the storage.
// This is part of the Auto-Save feature.
func (a *Adapter) RemovePoliciesCtx(ctx context.Context, sec string, ptype string, rules [][]string) error {
	err := a.dao.DB().Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		for _, rule := range rules {
			line := a.savePolicyLine(ptype, rule)
			_, err := a.dao.Ctx(ctx).Where(line).OmitEmpty().Delete()
			if err != nil {
				return err
			}
		}
		return nil
	})
	return err
}

// RemoveFilteredPolicy removes policy rules that match the filter from the storage.
// This is part of the Auto-Save feature.
func (a *Adapter) RemoveFilteredPolicy(sec string, ptype string, fieldIndex int, fieldValues ...string) error {
	return a.RemoveFilteredPolicyCtx(context.Background(), sec, ptype, fieldIndex, fieldValues...)
}

// RemoveFilteredPolicyCtx removes policy rules that match the filter from the storage with context.
// This is part of the Auto-Save feature.
func (a *Adapter) RemoveFilteredPolicyCtx(ctx context.Context, sec string, ptype string, fieldIndex int, fieldValues ...string) error {
	line := &entity.CasbinRule{}
	line.Ptype = ptype

	// If fieldIndex is -1, delete all policies with the specified ptype
	if fieldIndex == -1 {
		_, err := a.dao.Ctx(ctx).Where(line).OmitEmpty().Delete()
		return err
	}

	// Check if all query fields are empty
	err := a.checkQueryField(fieldValues)
	if err != nil {
		return err
	}

	// Set filter conditions based on fieldIndex and fieldValues
	idx := fieldIndex + len(fieldValues)
	if fieldIndex <= 0 && 0 < idx {
		line.V0 = fieldValues[0-fieldIndex]
	}
	if fieldIndex <= 1 && 1 < idx {
		line.V1 = fieldValues[1-fieldIndex]
	}
	if fieldIndex <= 2 && 2 < idx {
		line.V2 = fieldValues[2-fieldIndex]
	}
	if fieldIndex <= 3 && 3 < idx {
		line.V3 = fieldValues[3-fieldIndex]
	}
	if fieldIndex <= 4 && 4 < idx {
		line.V4 = fieldValues[4-fieldIndex]
	}
	if fieldIndex <= 5 && 5 < idx {
		line.V5 = fieldValues[5-fieldIndex]
	}

	// Execute delete operation
	_, err = a.dao.Ctx(ctx).Where(line).OmitEmpty().Delete()
	return err
}

// This is part of the Auto-Save feature.
func (a *Adapter) UpdatePolicy(sec string, ptype string, oldRule, newRule []string) error {
	return a.UpdatePolicyCtx(context.Background(), sec, ptype, oldRule, newRule)
}

// UpdatePolicyCtx updates a policy rule from storage.
// This is part of the Auto-Save feature.
func (a *Adapter) UpdatePolicyCtx(ctx context.Context, sec string, ptype string, oldRule, newRule []string) error {
	oldLine := a.savePolicyLine(ptype, oldRule)
	newLine := a.savePolicyLine(ptype, newRule)
	_, err := a.dao.Ctx(ctx).Where(oldLine).OmitEmpty().Data(newLine).Update()
	return err
}

// UpdatePolicies updates some policy rules to storage, like db, redis.
func (a *Adapter) UpdatePolicies(sec string, ptype string, oldRules, newRules [][]string) error {
	return a.UpdatePoliciesCtx(context.Background(), sec, ptype, oldRules, newRules)
}

// UpdatePoliciesCtx updates some policy rules to storage, like db, redis.
func (a *Adapter) UpdatePoliciesCtx(ctx context.Context, sec string, ptype string, oldRules, newRules [][]string) error {
	tx, err := a.dao.DB().Begin(ctx)
	if err != nil {
		return err
	}

	oldP := make([]entity.CasbinRule, 0, len(oldRules))
	for _, oldRule := range oldRules {
		oldP = append(oldP, a.savePolicyLine(ptype, oldRule))
	}

	newP := make([]entity.CasbinRule, 0, len(newRules))
	for _, newRule := range newRules {
		newP = append(newP, a.savePolicyLine(ptype, newRule))
	}

	cols := a.dao.Columns()

	// Batch delete old policies - first query IDs, then batch delete
	if len(oldP) > 0 {
		var idsToDelete []int64
		for _, oldLine := range oldP {
			var ruleIds []int64
			var arr []gdb.Value
			if arr, err = a.dao.Ctx(ctx).Where(oldLine).OmitEmpty().Array(cols.Id); err != nil {
				tx.Rollback()
				return err
			}
			ruleIds = gconv.Int64s(arr)
			idsToDelete = append(idsToDelete, ruleIds...)
		}

		// Batch delete using IDs
		if len(idsToDelete) > 0 {
			if _, err := a.dao.Ctx(ctx).WhereIn(cols.Id, idsToDelete).Delete(); err != nil {
				tx.Rollback()
				return err
			}
		}
	}

	// Then add new policies
	if len(newP) > 0 {
		if _, err := a.dao.Ctx(ctx).Data(newP).InsertIgnore(); err != nil {
			tx.Rollback()
			return err
		}
	}
	return tx.Commit()
}

// UpdateFilteredPolicies deletes old rules and adds new rules.
func (a *Adapter) UpdateFilteredPolicies(sec string, ptype string, newRules [][]string, fieldIndex int, fieldValues ...string) ([][]string, error) {
	return a.UpdateFilteredPoliciesCtx(context.Background(), sec, ptype, newRules, fieldIndex, fieldValues...)
}

// UpdateFilteredPoliciesCtx deletes old rules and adds new rules.
func (a *Adapter) UpdateFilteredPoliciesCtx(ctx context.Context, sec string, ptype string, newRules [][]string, fieldIndex int, fieldValues ...string) ([][]string, error) {
	// Build filter conditions
	line := &entity.CasbinRule{}
	line.Ptype = ptype

	if fieldIndex <= 0 && 0 < fieldIndex+len(fieldValues) {
		line.V0 = fieldValues[0-fieldIndex]
	}
	if fieldIndex <= 1 && 1 < fieldIndex+len(fieldValues) {
		line.V1 = fieldValues[1-fieldIndex]
	}
	if fieldIndex <= 2 && 2 < fieldIndex+len(fieldValues) {
		line.V2 = fieldValues[2-fieldIndex]
	}
	if fieldIndex <= 3 && 3 < fieldIndex+len(fieldValues) {
		line.V3 = fieldValues[3-fieldIndex]
	}
	if fieldIndex <= 4 && 4 < fieldIndex+len(fieldValues) {
		line.V4 = fieldValues[4-fieldIndex]
	}
	if fieldIndex <= 5 && 5 < fieldIndex+len(fieldValues) {
		line.V5 = fieldValues[5-fieldIndex]
	}

	// Prepare new policy data
	newP := make([]entity.CasbinRule, 0, len(newRules))
	for _, newRule := range newRules {
		newP = append(newP, a.savePolicyLine(ptype, newRule))
	}

	// Begin transaction
	tx, err := a.dao.DB().Begin(ctx)
	if err != nil {
		return nil, err
	}

	// Query old policies to be deleted
	var oldP []entity.CasbinRule
	if err := a.dao.Ctx(ctx).Where(line).OmitEmpty().Scan(&oldP); err != nil {
		tx.Rollback()
		return nil, err
	}

	// Delete old policies
	if _, err := a.dao.Ctx(ctx).Where(line).OmitEmpty().Delete(); err != nil {
		tx.Rollback()
		return nil, err
	}

	// Batch add new policies
	if len(newP) > 0 {
		if _, err := a.dao.Ctx(ctx).Data(newP).InsertIgnore(); err != nil {
			tx.Rollback()
			return nil, err
		}
	}

	// Build list of deleted policies to return
	oldPolicies := make([][]string, 0)
	for _, v := range oldP {
		oldPolicy := a.toStringPolicy(v)
		oldPolicies = append(oldPolicies, oldPolicy)
	}

	return oldPolicies, tx.Commit()
}

// ClearPolicy clears all current policy in all instances
func (a *Adapter) ClearPolicy() error {
	return a.truncateTable(context.Background())
}

// truncateTable clears the table
func (a *Adapter) truncateTable(ctx context.Context) error {
	tableName := a.dao.Table()
	dbType := a.dao.DB().GetConfig().Type

	var sql string
	switch dbType {
	case "sqlite":
		sql = fmt.Sprintf("delete from %s", tableName)
	case "sqlite3":
		sql = fmt.Sprintf("delete from %s", tableName)
	case "postgres":
		sql = fmt.Sprintf("truncate table %s RESTART IDENTITY", tableName)
	case "sqlserver":
		sql = fmt.Sprintf("truncate table %s", tableName)
	case "mysql":
		sql = fmt.Sprintf("truncate table %s", tableName)
	default:
		sql = fmt.Sprintf("truncate table %s", tableName)
	}
	_, err := a.dao.DB().Exec(ctx, sql)
	return err
}

// loadPolicyLine loads policy line
func loadPolicyLine(line entity.CasbinRule, model model.Model) error {
	var p = []string{line.Ptype,
		line.V0, line.V1, line.V2,
		line.V3, line.V4, line.V5}

	index := len(p) - 1
	for p[index] == "" {
		index--
	}
	index += 1
	p = p[:index]
	err := persist.LoadPolicyArray(p, model)
	if err != nil {
		return err
	}
	return nil
}

// preview Pre-checking to avoid causing partial load success and partial failure deep
func (a *Adapter) preview(rules *[]entity.CasbinRule, model model.Model) error {
	j := 0
	for i, rule := range *rules {
		r := []string{rule.Ptype,
			rule.V0, rule.V1, rule.V2,
			rule.V3, rule.V4, rule.V5}
		index := len(r) - 1
		for r[index] == "" {
			index--
		}
		index += 1
		p := r[:index]
		key := p[0]
		sec := key[:1]
		ok, err := model.HasPolicyEx(sec, key, p[1:])
		if err != nil {
			return err
		}
		if ok {
			(*rules)[j], (*rules)[i] = rule, (*rules)[j]
			j++
		}
	}
	(*rules) = (*rules)[j:]
	return nil
}

func (a *Adapter) savePolicyLine(ptype string, rule []string) entity.CasbinRule {
	line := &entity.CasbinRule{}

	line.Ptype = ptype
	if len(rule) > 0 {
		line.V0 = rule[0]
	}
	if len(rule) > 1 {
		line.V1 = rule[1]
	}
	if len(rule) > 2 {
		line.V2 = rule[2]
	}
	if len(rule) > 3 {
		line.V3 = rule[3]
	}
	if len(rule) > 4 {
		line.V4 = rule[4]
	}
	if len(rule) > 5 {
		line.V5 = rule[5]
	}

	return *line
}

// checkQueryField ensures that query fields are not all empty strings
func (a *Adapter) checkQueryField(fieldValues []string) error {
	for _, fieldValue := range fieldValues {
		if fieldValue != "" {
			return nil
		}
	}
	return errors.New("the query field cannot all be empty string (\"\"), please check")
}

// toStringPolicy converts CasbinRule to string policy array
func (a *Adapter) toStringPolicy(c entity.CasbinRule) []string {
	policy := make([]string, 0)
	if c.Ptype != "" {
		policy = append(policy, c.Ptype)
	}
	if c.V0 != "" {
		policy = append(policy, c.V0)
	}
	if c.V1 != "" {
		policy = append(policy, c.V1)
	}
	if c.V2 != "" {
		policy = append(policy, c.V2)
	}
	if c.V3 != "" {
		policy = append(policy, c.V3)
	}
	if c.V4 != "" {
		policy = append(policy, c.V4)
	}
	if c.V5 != "" {
		policy = append(policy, c.V5)
	}
	return policy
}

// applyFilter
func (a *Adapter) applyFilter(qs *gdb.Model, filter Filter) {
	cols := a.dao.Columns()

	// Apply ptype filtering
	if len(filter.Ptype) > 0 {
		qs.WhereIn(cols.Ptype, filter.Ptype)
	}

	// Apply V0-V5 filter conditions
	filterFields := []string{cols.V0, cols.V1, cols.V2, cols.V3, cols.V4, cols.V5}
	filterValues := [][]string{filter.V0, filter.V1, filter.V2, filter.V3, filter.V4, filter.V5}

	for i, values := range filterValues {
		if len(values) > 0 && i < len(filterFields) {
			qs.WhereIn(filterFields[i], values)
		}
	}
}
