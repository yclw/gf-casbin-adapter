package gfadapter

import (
	"log"
	"testing"

	"github.com/casbin/casbin/v2"
	"github.com/casbin/casbin/v2/util"
	"github.com/stretchr/testify/assert"

	_ "github.com/gogf/gf/contrib/drivers/mysql/v2"
)

func testGetPolicy(t *testing.T, e *casbin.Enforcer, res [][]string) {
	myRes, err := e.GetPolicy()
	if err != nil {
		panic(err)
	}

	log.Print("Policy: ", myRes)

	if !arrayEqualsWithoutOrder(myRes, res) {
		t.Error("Policy: ", myRes, ", supposed to be ", res)
	}
}

func arrayEqualsWithoutOrder(a [][]string, b [][]string) bool {
	if len(a) != len(b) {
		return false
	}

	setA := make(map[string]int)
	setB := make(map[string]int)

	for _, policy := range a {
		key := util.ArrayToString(policy)
		setA[key]++
	}

	for _, policy := range b {
		key := util.ArrayToString(policy)
		setB[key]++
	}

	if len(setA) != len(setB) {
		return false
	}

	for key, countA := range setA {
		if countB, exists := setB[key]; !exists || countA != countB {
			return false
		}
	}

	return true
}

func initPolicy(t *testing.T, a *Adapter) {
	// Because the DB is empty at first,
	// so we need to load the policy from the file adapter (.CSV) first.
	e, err := casbin.NewEnforcer("examples/rbac_model.conf", "examples/rbac_policy.csv")
	if err != nil {
		panic(err)
	}

	// This is a trick to save the current policy to the DB.
	// We can't call e.SavePolicy() because the adapter in the enforcer is still the file adapter.
	// The current policy means the policy in the Casbin enforcer (aka in memory).
	err = a.SavePolicy(e.GetModel())
	if err != nil {
		panic(err)
	}

	// Clear the current policy.
	e.ClearPolicy()
	testGetPolicy(t, e, [][]string{})

	// Load the policy from DB.
	err = a.LoadPolicy(e.GetModel())
	if err != nil {
		panic(err)
	}
	testGetPolicy(t, e, [][]string{{"alice", "data1", "read"}, {"bob", "data2", "write"}, {"data2_admin", "data2", "read"}, {"data2_admin", "data2", "write"}})
}

func testSaveLoad(t *testing.T, a *Adapter) {
	// Initialize some policy in DB.
	initPolicy(t, a)
	// Note: you don't need to look at the above code
	// if you already have a working DB with policy inside.

	// Now the DB has policy, so we can provide a normal use case.
	// Create an adapter and an enforcer.
	// NewEnforcer() will load the policy automatically.
	e, _ := casbin.NewEnforcer("examples/rbac_model.conf", a)
	testGetPolicy(t, e, [][]string{{"alice", "data1", "read"}, {"bob", "data2", "write"}, {"data2_admin", "data2", "read"}, {"data2_admin", "data2", "write"}})
}

func initAdapter(t *testing.T) *Adapter {
	// Create an adapter
	a, err := NewAdapter()
	if err != nil {
		panic(err)
	}

	// Initialize some policy in DB.
	initPolicy(t, a)
	// Now the DB has policy, so we can provide a normal use case.
	// Note: you don't need to look at the above code
	// if you already have a working DB with policy inside.

	return a
}

func initAdapterWithName(t *testing.T, tableName string) *Adapter {
	// Create an adapter with custom table name
	a, err := NewAdapterWithName(tableName, false)
	if err != nil {
		panic(err)
	}

	// Initialize some policy in DB.
	initPolicy(t, a)
	// Now the DB has policy, so we can provide a normal use case.
	// Note: you don't need to look at the above code
	// if you already have a working DB with policy inside.

	return a
}

func TestNilField(t *testing.T) {
	a, err := NewAdapter()
	assert.Nil(t, err)

	e, err := casbin.NewEnforcer("examples/rbac_model.conf", a)
	assert.Nil(t, err)
	e.EnableAutoSave(false)

	_, err = e.AddPolicy("", "data1", "write")
	assert.Nil(t, err)
	e.SavePolicy()
	assert.Nil(t, e.LoadPolicy())

	ok, err := e.Enforce("", "data1", "write")
	assert.Nil(t, err)
	assert.Equal(t, ok, true)
}

func testAutoSave(t *testing.T, a *Adapter) {

	// NewEnforcer() will load the policy automatically.
	e, _ := casbin.NewEnforcer("examples/rbac_model.conf", a)
	// AutoSave is enabled by default.
	// Now we disable it.
	e.EnableAutoSave(false)

	// Because AutoSave is disabled, the policy change only affects the policy in Casbin enforcer,
	// it doesn't affect the policy in the storage.
	e.AddPolicy("alice", "data1", "write")
	// Reload the policy from the storage to see the effect.
	e.LoadPolicy()
	// This is still the original policy.
	testGetPolicy(t, e, [][]string{{"alice", "data1", "read"}, {"bob", "data2", "write"}, {"data2_admin", "data2", "read"}, {"data2_admin", "data2", "write"}})

	// Now we enable the AutoSave.
	e.EnableAutoSave(true)

	// Because AutoSave is enabled, the policy change not only affects the policy in Casbin enforcer,
	// but also affects the policy in the storage.
	e.AddPolicy("alice", "data1", "write")
	// Reload the policy from the storage to see the effect.
	e.LoadPolicy()
	// The policy has a new rule: {"alice", "data1", "write"}.
	testGetPolicy(t, e, [][]string{{"alice", "data1", "read"}, {"bob", "data2", "write"}, {"data2_admin", "data2", "read"}, {"data2_admin", "data2", "write"}, {"alice", "data1", "write"}})

	// Remove the added rule.
	e.RemovePolicy("alice", "data1", "write")
	e.LoadPolicy()
	testGetPolicy(t, e, [][]string{{"alice", "data1", "read"}, {"bob", "data2", "write"}, {"data2_admin", "data2", "read"}, {"data2_admin", "data2", "write"}})

	// Remove "data2_admin" related policy rules via a filter.
	// Two rules: {"data2_admin", "data2", "read"}, {"data2_admin", "data2", "write"} are deleted.
	e.RemoveFilteredPolicy(0, "data2_admin")
	e.LoadPolicy()
	testGetPolicy(t, e, [][]string{{"alice", "data1", "read"}, {"bob", "data2", "write"}})
}

func testFilteredPolicy(t *testing.T, a *Adapter) {
	// NewEnforcer() without an adapter will not auto load the policy
	e, _ := casbin.NewEnforcer("examples/rbac_model.conf")
	// Now set the adapter
	e.SetAdapter(a)

	// Load only alice's policies
	assert.Nil(t, e.LoadFilteredPolicy(Filter{V0: []string{"alice"}}))
	testGetPolicy(t, e, [][]string{{"alice", "data1", "read"}})

	// Load only bob's policies
	assert.Nil(t, e.LoadFilteredPolicy(Filter{V0: []string{"bob"}}))
	testGetPolicy(t, e, [][]string{{"bob", "data2", "write"}})

	// Load policies for data2_admin
	assert.Nil(t, e.LoadFilteredPolicy(Filter{V0: []string{"data2_admin"}}))
	testGetPolicy(t, e, [][]string{{"data2_admin", "data2", "read"}, {"data2_admin", "data2", "write"}})

	// Load policies for alice and bob
	assert.Nil(t, e.LoadFilteredPolicy(Filter{V0: []string{"alice", "bob"}}))
	testGetPolicy(t, e, [][]string{{"alice", "data1", "read"}, {"bob", "data2", "write"}})

	// Test combined filters (alice OR data2)
	assert.Nil(t, e.LoadFilteredPolicy(Filter{V0: []string{"alice", "data2_admin"}}))
	testGetPolicy(t, e, [][]string{
		{"alice", "data1", "read"},
		{"data2_admin", "data2", "read"},
		{"data2_admin", "data2", "write"},
	})
}

func testUpdatePolicy(t *testing.T, a *Adapter) {
	// NewEnforcer() will load the policy automatically.
	e, _ := casbin.NewEnforcer("examples/rbac_model.conf", a)

	e.EnableAutoSave(true)
	e.UpdatePolicy([]string{"alice", "data1", "read"}, []string{"alice", "data1", "write"})
	e.LoadPolicy()
	testGetPolicy(t, e, [][]string{{"alice", "data1", "write"}, {"bob", "data2", "write"}, {"data2_admin", "data2", "read"}, {"data2_admin", "data2", "write"}})
}

func testUpdatePolicies(t *testing.T, a *Adapter) {
	// NewEnforcer() will load the policy automatically.
	e, _ := casbin.NewEnforcer("examples/rbac_model.conf", a)

	e.EnableAutoSave(true)
	e.UpdatePolicies([][]string{{"alice", "data1", "write"}, {"bob", "data2", "write"}}, [][]string{{"alice", "data1", "read"}, {"bob", "data2", "read"}})
	e.LoadPolicy()
	testGetPolicy(t, e, [][]string{{"alice", "data1", "read"}, {"bob", "data2", "read"}, {"data2_admin", "data2", "read"}, {"data2_admin", "data2", "write"}})
}

func testUpdateFilteredPolicies(t *testing.T, a *Adapter) {
	// NewEnforcer() will load the policy automatically.
	e, _ := casbin.NewEnforcer("examples/rbac_model.conf", a)

	e.EnableAutoSave(true)
	e.UpdateFilteredPolicies([][]string{{"alice", "data1", "write"}}, 0, "alice", "data1", "read")
	e.UpdateFilteredPolicies([][]string{{"bob", "data2", "read"}}, 0, "bob", "data2", "write")
	e.LoadPolicy()
	testGetPolicy(t, e, [][]string{{"alice", "data1", "write"}, {"data2_admin", "data2", "read"}, {"data2_admin", "data2", "write"}, {"bob", "data2", "read"}})
}

func TestAdapterWithCustomTable(t *testing.T) {
	a := initAdapterWithName(t, "test_casbin_rule")
	testAutoSave(t, a)
	testSaveLoad(t, a)

	a = initAdapterWithName(t, "test_casbin_rule")
	testFilteredPolicy(t, a)
}

func TestAdapters(t *testing.T) {
	// Test basic adapter
	a := initAdapter(t)
	testAutoSave(t, a)
	testSaveLoad(t, a)

	// Test filtered policy
	a = initAdapter(t)
	testFilteredPolicy(t, a)

	// Test with custom table name
	a = initAdapterWithName(t, "casbin_rule")
	testAutoSave(t, a)
	testSaveLoad(t, a)

	a = initAdapterWithName(t, "casbin_rule")
	testFilteredPolicy(t, a)

	// Test update operations
	a = initAdapter(t)
	testUpdatePolicy(t, a)
	testUpdatePolicies(t, a)
	testUpdateFilteredPolicies(t, a)
}

func TestAddPolicies(t *testing.T) {
	a := initAdapter(t)
	e, _ := casbin.NewEnforcer("examples/rbac_model.conf", a)
	e.AddPolicies([][]string{{"jack", "data1", "read"}, {"jack2", "data1", "read"}})
	e.LoadPolicy()

	testGetPolicy(t, e, [][]string{{"alice", "data1", "read"}, {"bob", "data2", "write"}, {"data2_admin", "data2", "read"}, {"data2_admin", "data2", "write"}, {"jack", "data1", "read"}, {"jack2", "data1", "read"}})
}

func TestAddPolicy(t *testing.T) {
	a := initAdapter(t)
	e1, _ := casbin.NewEnforcer("examples/rbac_model.conf", a)
	e2, _ := casbin.NewEnforcer("examples/rbac_model.conf", a)

	policy := []string{"alice", "data1", "TestAddPolicy"}

	ok, err := e1.AddPolicy(policy)
	if err != nil {
		t.Errorf("e1.AddPolicy() got err %v", err)
	}
	if !ok {
		t.Errorf("e1.AddPolicy() got false, want true")
	}

	// Reload policy for e2 to sync with database
	err = e2.LoadPolicy()
	if err != nil {
		t.Errorf("e2.LoadPolicy() got err %v", err)
	}

	ok, err = e2.AddPolicy(policy)
	if err != nil {
		t.Errorf("e2.AddPolicy() got err %v", err)
	}
	if ok {
		t.Errorf("e2.AddPolicy() got true, want false (policy already exists)")
	}
}
