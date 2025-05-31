package jwt

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/threatflux/libgo/internal/models/user"
)

func TestNewClaims(t *testing.T) {
	// Create a user for testing
	testUser := &user.User{
		ID:       "test-id",
		Username: "testuser",
		Roles:    []string{user.RoleAdmin, user.RoleOperator},
		Active:   true,
	}

	// Create registered claims
	issuer := "test-issuer"
	audience := jwt.ClaimStrings{"test-audience"}
	subject := "test-subject"
	issuedAt := jwt.NewNumericDate(time.Now())
	expiration := jwt.NewNumericDate(time.Now().Add(15 * time.Minute))
	registeredClaims := jwt.RegisteredClaims{
		Issuer:    issuer,
		Subject:   subject,
		Audience:  audience,
		ExpiresAt: expiration,
		IssuedAt:  issuedAt,
	}

	// Create claims
	claims := NewClaims(testUser, registeredClaims)

	// Check registered claims
	if claims.Issuer != issuer {
		t.Errorf("Expected Issuer to be %q, got %q", issuer, claims.Issuer)
	}

	if claims.Subject != subject {
		t.Errorf("Expected Subject to be %q, got %q", subject, claims.Subject)
	}

	found := false
	for _, aud := range claims.Audience {
		if aud == audience[0] {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("Expected Audience to contain %q", audience[0])
	}

	if claims.ExpiresAt == nil || !claims.ExpiresAt.Equal(expiration.Time) {
		t.Errorf("Expected ExpiresAt to be %v, got %v", expiration, claims.ExpiresAt)
	}

	if claims.IssuedAt == nil || !claims.IssuedAt.Equal(issuedAt.Time) {
		t.Errorf("Expected IssuedAt to be %v, got %v", issuedAt, claims.IssuedAt)
	}

	// Check user claims
	if claims.UserID != testUser.ID {
		t.Errorf("Expected UserID to be %q, got %q", testUser.ID, claims.UserID)
	}

	if claims.Username != testUser.Username {
		t.Errorf("Expected Username to be %q, got %q", testUser.Username, claims.Username)
	}

	if len(claims.Roles) != len(testUser.Roles) {
		t.Errorf("Expected Roles length to be %d, got %d", len(testUser.Roles), len(claims.Roles))
	}

	for i, role := range testUser.Roles {
		if claims.Roles[i] != role {
			t.Errorf("Expected Role[%d] to be %q, got %q", i, role, claims.Roles[i])
		}
	}
}

func TestClaims_Valid(t *testing.T) {
	now := time.Now()
	expiry := now.Add(15 * time.Minute)

	validRegisteredClaims := jwt.RegisteredClaims{
		Issuer:    "test-issuer",
		Subject:   "test-subject",
		Audience:  jwt.ClaimStrings{"test-audience"},
		ExpiresAt: jwt.NewNumericDate(expiry),
		IssuedAt:  jwt.NewNumericDate(now),
	}

	tests := []struct {
		name    string
		claims  *Claims
		wantErr bool
	}{
		{
			name: "Valid claims",
			claims: &Claims{
				RegisteredClaims: validRegisteredClaims,
				UserID:           "test-id",
				Username:         "testuser",
				Roles:            []string{user.RoleAdmin},
			},
			wantErr: false,
		},
		{
			name: "Expired token",
			claims: &Claims{
				RegisteredClaims: jwt.RegisteredClaims{
					ExpiresAt: jwt.NewNumericDate(now.Add(-1 * time.Hour)),
				},
				UserID:   "test-id",
				Username: "testuser",
				Roles:    []string{user.RoleAdmin},
			},
			wantErr: true,
		},
		{
			name: "Missing user ID",
			claims: &Claims{
				RegisteredClaims: validRegisteredClaims,
				UserID:           "",
				Username:         "testuser",
				Roles:            []string{user.RoleAdmin},
			},
			wantErr: true,
		},
		{
			name: "Missing username",
			claims: &Claims{
				RegisteredClaims: validRegisteredClaims,
				UserID:           "test-id",
				Username:         "",
				Roles:            []string{user.RoleAdmin},
			},
			wantErr: true,
		},
		{
			name: "Empty roles",
			claims: &Claims{
				RegisteredClaims: validRegisteredClaims,
				UserID:           "test-id",
				Username:         "testuser",
				Roles:            []string{},
			},
			wantErr: true,
		},
		{
			name: "Invalid role",
			claims: &Claims{
				RegisteredClaims: validRegisteredClaims,
				UserID:           "test-id",
				Username:         "testuser",
				Roles:            []string{"invalid-role"},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.claims.Valid()
			if (err != nil) != tt.wantErr {
				t.Errorf("Claims.Valid() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestClaims_HasPermission(t *testing.T) {
	tests := []struct {
		name       string
		roles      []string
		permission string
		want       bool
	}{
		{
			name:       "Admin has create permission",
			roles:      []string{user.RoleAdmin},
			permission: user.PermCreate,
			want:       true,
		},
		{
			name:       "Operator has read permission",
			roles:      []string{user.RoleOperator},
			permission: user.PermRead,
			want:       true,
		},
		{
			name:       "Operator does not have create permission",
			roles:      []string{user.RoleOperator},
			permission: user.PermCreate,
			want:       false,
		},
		{
			name:       "Multiple roles with permission",
			roles:      []string{user.RoleViewer, user.RoleOperator},
			permission: user.PermUpdate,
			want:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			claims := &Claims{
				RegisteredClaims: jwt.RegisteredClaims{},
				UserID:           "test-id",
				Username:         "testuser",
				Roles:            tt.roles,
			}

			if got := claims.HasPermission(tt.permission); got != tt.want {
				t.Errorf("Claims.HasPermission() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestClaims_HasRole(t *testing.T) {
	tests := []struct {
		name  string
		roles []string
		role  string
		want  bool
	}{
		{
			name:  "Has admin role",
			roles: []string{user.RoleAdmin, user.RoleOperator},
			role:  user.RoleAdmin,
			want:  true,
		},
		{
			name:  "Has operator role",
			roles: []string{user.RoleAdmin, user.RoleOperator},
			role:  user.RoleOperator,
			want:  true,
		},
		{
			name:  "Does not have viewer role",
			roles: []string{user.RoleAdmin, user.RoleOperator},
			role:  user.RoleViewer,
			want:  false,
		},
		{
			name:  "Does not have non-existent role",
			roles: []string{user.RoleAdmin, user.RoleOperator},
			role:  "non-existent",
			want:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			claims := &Claims{
				RegisteredClaims: jwt.RegisteredClaims{},
				UserID:           "test-id",
				Username:         "testuser",
				Roles:            tt.roles,
			}

			if got := claims.HasRole(tt.role); got != tt.want {
				t.Errorf("Claims.HasRole() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestClaims_ToUser(t *testing.T) {
	claims := &Claims{
		RegisteredClaims: jwt.RegisteredClaims{},
		UserID:           "test-id",
		Username:         "testuser",
		Roles:            []string{user.RoleAdmin, user.RoleOperator},
	}

	userModel := claims.ToUser()

	// Check user properties
	if userModel.ID != claims.UserID {
		t.Errorf("Expected ID to be %q, got %q", claims.UserID, userModel.ID)
	}

	if userModel.Username != claims.Username {
		t.Errorf("Expected Username to be %q, got %q", claims.Username, userModel.Username)
	}

	if len(userModel.Roles) != len(claims.Roles) {
		t.Errorf("Expected Roles length to be %d, got %d", len(claims.Roles), len(userModel.Roles))
	}

	for i, role := range claims.Roles {
		if userModel.Roles[i] != role {
			t.Errorf("Expected Role[%d] to be %q, got %q", i, role, userModel.Roles[i])
		}
	}

	if !userModel.Active {
		t.Error("Expected user to be active")
	}

	// Password should not be set
	if userModel.Password != "" {
		t.Errorf("Expected Password to be empty, got %q", userModel.Password)
	}
}
