package user

import (
	"reflect"
	"sort"
	"testing"
)

func TestGetRolePermissions(t *testing.T) {
	tests := []struct {
		name string
		role string
		want []string
	}{
		{
			name: "Admin permissions",
			role: RoleAdmin,
			want: []string{PermCreate, PermRead, PermUpdate, PermDelete, PermStart, PermStop, PermExport},
		},
		{
			name: "Operator permissions",
			role: RoleOperator,
			want: []string{PermRead, PermUpdate, PermStart, PermStop, PermExport},
		},
		{
			name: "Viewer permissions",
			role: RoleViewer,
			want: []string{PermRead},
		},
		{
			name: "Non-existent role",
			role: "nonexistent",
			want: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetRolePermissions(tt.role)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetRolePermissions() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHasPermission(t *testing.T) {
	tests := []struct {
		name       string
		role       string
		permission string
		want       bool
	}{
		{
			name:       "Admin has create permission",
			role:       RoleAdmin,
			permission: PermCreate,
			want:       true,
		},
		{
			name:       "Operator has read permission",
			role:       RoleOperator,
			permission: PermRead,
			want:       true,
		},
		{
			name:       "Operator does not have create permission",
			role:       RoleOperator,
			permission: PermCreate,
			want:       false,
		},
		{
			name:       "Viewer has read permission",
			role:       RoleViewer,
			permission: PermRead,
			want:       true,
		},
		{
			name:       "Viewer does not have create permission",
			role:       RoleViewer,
			permission: PermCreate,
			want:       false,
		},
		{
			name:       "Non-existent role",
			role:       "nonexistent",
			permission: PermRead,
			want:       false,
		},
		{
			name:       "Non-existent permission",
			role:       RoleAdmin,
			permission: "nonexistent",
			want:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := HasPermission(tt.role, tt.permission)
			if got != tt.want {
				t.Errorf("HasPermission() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetUserPermissions(t *testing.T) {
	tests := []struct {
		name  string
		roles []string
		want  []string
	}{
		{
			name:  "Admin only",
			roles: []string{RoleAdmin},
			want:  []string{PermCreate, PermRead, PermUpdate, PermDelete, PermStart, PermStop, PermExport},
		},
		{
			name:  "Operator only",
			roles: []string{RoleOperator},
			want:  []string{PermRead, PermUpdate, PermStart, PermStop, PermExport},
		},
		{
			name:  "Viewer only",
			roles: []string{RoleViewer},
			want:  []string{PermRead},
		},
		{
			name:  "Admin and Operator",
			roles: []string{RoleAdmin, RoleOperator},
			want:  []string{PermCreate, PermRead, PermUpdate, PermDelete, PermStart, PermStop, PermExport},
		},
		{
			name:  "Admin, Operator, and Viewer",
			roles: []string{RoleAdmin, RoleOperator, RoleViewer},
			want:  []string{PermCreate, PermRead, PermUpdate, PermDelete, PermStart, PermStop, PermExport},
		},
		{
			name:  "Operator and Viewer",
			roles: []string{RoleOperator, RoleViewer},
			want:  []string{PermRead, PermUpdate, PermStart, PermStop, PermExport},
		},
		{
			name:  "No roles",
			roles: []string{},
			want:  []string{},
		},
		{
			name:  "Non-existent role",
			roles: []string{"nonexistent"},
			want:  []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetUserPermissions(tt.roles)

			// Sort both slices for comparison since map iteration order is not guaranteed
			sort.Strings(got)
			sort.Strings(tt.want)

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetUserPermissions() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUserHasPermission(t *testing.T) {
	tests := []struct {
		name       string
		roles      []string
		permission string
		want       bool
	}{
		{
			name:       "Admin has create permission",
			roles:      []string{RoleAdmin},
			permission: PermCreate,
			want:       true,
		},
		{
			name:       "Operator has read permission",
			roles:      []string{RoleOperator},
			permission: PermRead,
			want:       true,
		},
		{
			name:       "Operator does not have create permission",
			roles:      []string{RoleOperator},
			permission: PermCreate,
			want:       false,
		},
		{
			name:       "Viewer and Operator together have update permission",
			roles:      []string{RoleViewer, RoleOperator},
			permission: PermUpdate,
			want:       true,
		},
		{
			name:       "Viewer and Operator together do not have create permission",
			roles:      []string{RoleViewer, RoleOperator},
			permission: PermCreate,
			want:       false,
		},
		{
			name:       "No roles",
			roles:      []string{},
			permission: PermRead,
			want:       false,
		},
		{
			name:       "Non-existent role",
			roles:      []string{"nonexistent"},
			permission: PermRead,
			want:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := UserHasPermission(tt.roles, tt.permission)
			if got != tt.want {
				t.Errorf("UserHasPermission() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRoles(t *testing.T) {
	roles := Roles()
	expected := []string{RoleAdmin, RoleOperator, RoleViewer}

	if len(roles) != len(expected) {
		t.Errorf("Roles() returned %d roles, expected %d", len(roles), len(expected))
	}

	for _, r := range expected {
		found := false
		for _, role := range roles {
			if role == r {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Roles() did not include %s", r)
		}
	}
}

func TestPermissions(t *testing.T) {
	perms := Permissions()
	expected := []string{
		PermCreate,
		PermRead,
		PermUpdate,
		PermDelete,
		PermStart,
		PermStop,
		PermExport,
	}

	if len(perms) != len(expected) {
		t.Errorf("Permissions() returned %d permissions, expected %d", len(perms), len(expected))
	}

	for _, p := range expected {
		found := false
		for _, perm := range perms {
			if perm == p {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Permissions() did not include %s", p)
		}
	}
}

func TestIsValidRole(t *testing.T) {
	tests := []struct {
		name string
		role string
		want bool
	}{
		{
			name: "Valid admin role",
			role: RoleAdmin,
			want: true,
		},
		{
			name: "Valid operator role",
			role: RoleOperator,
			want: true,
		},
		{
			name: "Valid viewer role",
			role: RoleViewer,
			want: true,
		},
		{
			name: "Invalid role",
			role: "invalid",
			want: false,
		},
		{
			name: "Empty role",
			role: "",
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsValidRole(tt.role); got != tt.want {
				t.Errorf("IsValidRole() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsValidPermission(t *testing.T) {
	tests := []struct {
		name       string
		permission string
		want       bool
	}{
		{
			name:       "Valid create permission",
			permission: PermCreate,
			want:       true,
		},
		{
			name:       "Valid read permission",
			permission: PermRead,
			want:       true,
		},
		{
			name:       "Valid update permission",
			permission: PermUpdate,
			want:       true,
		},
		{
			name:       "Valid delete permission",
			permission: PermDelete,
			want:       true,
		},
		{
			name:       "Valid start permission",
			permission: PermStart,
			want:       true,
		},
		{
			name:       "Valid stop permission",
			permission: PermStop,
			want:       true,
		},
		{
			name:       "Valid export permission",
			permission: PermExport,
			want:       true,
		},
		{
			name:       "Invalid permission",
			permission: "invalid",
			want:       false,
		},
		{
			name:       "Empty permission",
			permission: "",
			want:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsValidPermission(tt.permission); got != tt.want {
				t.Errorf("IsValidPermission() = %v, want %v", got, tt.want)
			}
		})
	}
}
