package user

// User roles
const (
	RoleAdmin    = "admin"
	RoleOperator = "operator"
	RoleViewer   = "viewer"
)

// Permissions
const (
	PermCreate = "create"
	PermRead   = "read"
	PermUpdate = "update"
	PermDelete = "delete"
	PermStart  = "start"
	PermStop   = "stop"
	PermExport = "export"
)

// RolePermissions maps roles to their permissions
var RolePermissions = map[string][]string{
	RoleAdmin: {
		PermCreate, PermRead, PermUpdate, PermDelete,
		PermStart, PermStop, PermExport,
	},
	RoleOperator: {
		PermRead, PermUpdate, PermStart, PermStop, PermExport,
	},
	RoleViewer: {
		PermRead,
	},
}

// GetRolePermissions returns all permissions for a given role
func GetRolePermissions(role string) []string {
	return RolePermissions[role]
}

// HasPermission checks if a role has a specific permission
func HasPermission(role, permission string) bool {
	permissions, exists := RolePermissions[role]
	if !exists {
		return false
	}

	for _, p := range permissions {
		if p == permission {
			return true
		}
	}
	return false
}

// GetUserPermissions returns all unique permissions for a user based on their roles
func GetUserPermissions(roles []string) []string {
	// Use a map to deduplicate permissions
	permMap := make(map[string]struct{})
	
	for _, role := range roles {
		for _, perm := range RolePermissions[role] {
			permMap[perm] = struct{}{}
		}
	}
	
	// Convert map keys to slice
	perms := make([]string, 0, len(permMap))
	for perm := range permMap {
		perms = append(perms, perm)
	}
	
	return perms
}

// UserHasPermission checks if a user with the given roles has a specific permission
func UserHasPermission(roles []string, permission string) bool {
	for _, role := range roles {
		if HasPermission(role, permission) {
			return true
		}
	}
	return false
}

// Roles returns all valid roles
func Roles() []string {
	return []string{RoleAdmin, RoleOperator, RoleViewer}
}

// Permissions returns all valid permissions
func Permissions() []string {
	return []string{
		PermCreate,
		PermRead,
		PermUpdate,
		PermDelete,
		PermStart,
		PermStop,
		PermExport,
	}
}

// IsValidRole checks if a role is valid
func IsValidRole(role string) bool {
	switch role {
	case RoleAdmin, RoleOperator, RoleViewer:
		return true
	default:
		return false
	}
}

// IsValidPermission checks if a permission is valid
func IsValidPermission(permission string) bool {
	switch permission {
	case PermCreate, PermRead, PermUpdate, PermDelete, PermStart, PermStop, PermExport:
		return true
	default:
		return false
	}
}
