package authz

import "testing"

func TestRoleRanks(t *testing.T) {
	if roleRank[RoleViewer] >= roleRank[RoleMember] {
		t.Fatal("viewer should rank below member")
	}
	if roleRank[RoleMember] >= roleRank[RoleAdmin] {
		t.Fatal("member should rank below admin")
	}
}
