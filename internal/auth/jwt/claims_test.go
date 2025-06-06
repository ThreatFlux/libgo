package jwt

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/threatflux/libgo/internal/models/user"
)

const claimsTestIssuer = "test-issuer"

func TestNewClaims(t *testing.T) {
	// Create a user for testing
	testUser := &user.User{
		ID:       "test-id",
		Username: "testuser",
		Roles:    []string{user.RoleAdmin, user.RoleOperator},
		Active:   true,
	}

	// Create registered claims
	issuer := claimsTestIssuer
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
		Issuer:    claimsTestIssuer,
		Subject:   "test-subject",
		Audience:  jwt.ClaimStrings{"test-audience"},
		ExpiresAt: jwt.NewNumericDate(expiry),
		IssuedAt:  jwt.NewNumericDate(now),
	}

	tests := []struct {
		claims  *Claims // 8 bytes (pointer)
		name    string  // 16 bytes (string header)
		wantErr bool    // 1 byte (bool)
	}{
		{
			claims: &Claims{
				RegisteredClaims: validRegisteredClaims,
				UserID:           "test-id",
				Username:         "testuser",
				Roles:            []string{user.RoleAdmin},
			},
			name:    "Valid claims",
			wantErr: false,
		},
		{
			claims: &Claims{
				RegisteredClaims: jwt.RegisteredClaims{
					ExpiresAt: jwt.NewNumericDate(now.Add(-1 * time.Hour)),
				},
				UserID:   "test-id",
				Username: "testuser",
				Roles:    []string{user.RoleAdmin},
			},
			name:    "Expired token",
			wantErr: true,
		},
		{
			claims: &Claims{
				RegisteredClaims: validRegisteredClaims,
				UserID:           "",
				Username:         "testuser",
				Roles:            []string{user.RoleAdmin},
			},
			name:    "Missing user ID",
			wantErr: true,
		},
		{
			claims: &Claims{
				RegisteredClaims: validRegisteredClaims,
				UserID:           "test-id",
				Username:         "",
				Roles:            []string{user.RoleAdmin},
			},
			name:    "Missing username",
			wantErr: true,
		},
		{
			claims: &Claims{
				RegisteredClaims: validRegisteredClaims,
				UserID:           "test-id",
				Username:         "testuser",
				Roles:            []string{},
			},
			name:    "Empty roles",
			wantErr: true,
		},
		{
			claims: &Claims{
				RegisteredClaims: validRegisteredClaims,
				UserID:           "test-id",
				Username:         "testuser",
				Roles:            []string{"invalid-role"},
			},
			name:    "Invalid role",
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
		roles      []string // 24 bytes (slice header)
		name       string   // 16 bytes (string header)
		permission string   // 16 bytes (string header)
		want       bool     // 1 byte (bool)
	}{
		{
			roles:      []string{user.RoleAdmin},
			name:       "Admin has create permission",
			permission: user.PermCreate,
			want:       true,
		},
		{
			roles:      []string{user.RoleOperator},
			name:       "Operator has read permission",
			permission: user.PermRead,
			want:       true,
		},
		{
			roles:      []string{user.RoleOperator},
			name:       "Operator does not have create permission",
			permission: user.PermCreate,
			want:       false,
		},
		{
			roles:      []string{user.RoleViewer, user.RoleOperator},
			name:       "Multiple roles with permission",
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
		roles []string // 24 bytes (slice header)
		name  string   // 16 bytes (string header)
		role  string   // 16 bytes (string header)
		want  bool     // 1 byte (bool)
	}{
		{
			roles: []string{user.RoleAdmin, user.RoleOperator},
			name:  "Has admin role",
			role:  user.RoleAdmin,
			want:  true,
		},
		{
			roles: []string{user.RoleAdmin, user.RoleOperator},
			name:  "Has operator role",
			role:  user.RoleOperator,
			want:  true,
		},
		{
			roles: []string{user.RoleAdmin, user.RoleOperator},
			name:  "Does not have viewer role",
			role:  user.RoleViewer,
			want:  false,
		},
		{
			roles: []string{user.RoleAdmin, user.RoleOperator},
			name:  "Does not have non-existent role",
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
