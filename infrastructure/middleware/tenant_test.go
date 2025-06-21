package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DimaJoyti/go-coffee/domain/shared"
	"github.com/DimaJoyti/go-coffee/domain/tenant"
)

// MockTenantRepository for testing
type MockTenantRepository struct {
	tenants map[string]*tenant.Tenant
}

func NewMockTenantRepository() *MockTenantRepository {
	return &MockTenantRepository{
		tenants: make(map[string]*tenant.Tenant),
	}
}

func (m *MockTenantRepository) Save(ctx context.Context, tenant *tenant.Tenant) error {
	m.tenants[tenant.GetTenantID().Value()] = tenant
	return nil
}

func (m *MockTenantRepository) FindByID(ctx context.Context, id shared.AggregateID) (*tenant.Tenant, error) {
	for _, t := range m.tenants {
		if t.ID() == id {
			return t, nil
		}
	}
	return nil, nil
}

func (m *MockTenantRepository) FindByTenantID(ctx context.Context, tenantID shared.TenantID) (*tenant.Tenant, error) {
	if t, exists := m.tenants[tenantID.Value()]; exists {
		return t, nil
	}
	return nil, nil
}

func (m *MockTenantRepository) FindByEmail(ctx context.Context, email shared.Email) (*tenant.Tenant, error) {
	return nil, nil
}

func (m *MockTenantRepository) FindByOwnerID(ctx context.Context, ownerID shared.AggregateID) ([]*tenant.Tenant, error) {
	return nil, nil
}

func (m *MockTenantRepository) FindByStatus(ctx context.Context, status tenant.TenantStatus) ([]*tenant.Tenant, error) {
	return nil, nil
}

func (m *MockTenantRepository) FindBySubscriptionPlan(ctx context.Context, plan shared.SubscriptionPlan) ([]*tenant.Tenant, error) {
	return nil, nil
}

func (m *MockTenantRepository) ExistsByEmail(ctx context.Context, email shared.Email) (bool, error) {
	return false, nil
}

func (m *MockTenantRepository) Delete(ctx context.Context, id shared.AggregateID) error {
	return nil
}

func (m *MockTenantRepository) FindAll(ctx context.Context, offset, limit int) ([]*tenant.Tenant, int64, error) {
	return nil, 0, nil
}

func (m *MockTenantRepository) FindActiveTenants(ctx context.Context) ([]*tenant.Tenant, error) {
	return nil, nil
}

func (m *MockTenantRepository) CountByStatus(ctx context.Context, status tenant.TenantStatus) (int64, error) {
	return 0, nil
}

func (m *MockTenantRepository) FindExpiredSubscriptions(ctx context.Context) ([]*tenant.Tenant, error) {
	return nil, nil
}

// Helper function to create a test tenant
func createTestTenant(tenantID string) *tenant.Tenant {
	id := shared.NewAggregateID("test-aggregate-id")
	tid, _ := shared.NewTenantID(tenantID)
	email, _ := shared.NewEmail("test@example.com")
	phone, _ := shared.NewPhoneNumber("+1234567890")
	address, _ := shared.NewAddress("123 Test St", "Test City", "Test State", "12345", "Test Country")
	ownerID := shared.NewAggregateID("test-owner-id")

	testTenant, _ := tenant.NewTenant(
		id,
		tid,
		"Test Tenant",
		tenant.TenantTypeRestaurant,
		email,
		phone,
		address,
		ownerID,
	)

	// Activate the tenant
	testTenant.Activate()

	return testTenant
}

func TestHTTPTenantMiddleware_TenantContext(t *testing.T) {
	// Setup
	mockRepo := NewMockTenantRepository()
	testTenant := createTestTenant("test-tenant")
	mockRepo.Save(context.Background(), testTenant)

	middleware := NewHTTPTenantMiddleware(mockRepo, nil)

	// Create a test handler
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tenantCtx, err := GetHTTPTenantContext(r)
		if err != nil {
			t.Errorf("Expected tenant context, got error: %v", err)
			return
		}

		if tenantCtx.TenantID().Value() != "test-tenant" {
			t.Errorf("Expected tenant ID 'test-tenant', got '%s'", tenantCtx.TenantID().Value())
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// Wrap with middleware
	handler := middleware.TenantContext(testHandler)

	// Test with X-Tenant-ID header
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("X-Tenant-ID", "test-tenant")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

func TestHTTPTenantMiddleware_RequireFeature(t *testing.T) {
	// Setup
	mockRepo := NewMockTenantRepository()
	testTenant := createTestTenant("test-tenant")
	mockRepo.Save(context.Background(), testTenant)

	middleware := NewHTTPTenantMiddleware(mockRepo, nil)

	// Create a test handler
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Feature accessed"))
	})

	// Wrap with middleware
	handler := middleware.RequireFeature("test-feature")(
		middleware.TenantContext(testHandler),
	)

	// Test without feature (should fail)
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("X-Tenant-ID", "test-tenant")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("Expected status 403 (feature not available), got %d", w.Code)
	}
}

func TestTenantContextMiddleware_ExtractFromSubdomain(t *testing.T) {
	middleware := NewTenantContextMiddleware(NewMockTenantRepository())

	tests := []struct {
		host     string
		expected string
	}{
		{"tenant1.api.example.com", "tenant1"},
		{"tenant2.api.example.com:8080", "tenant2"},
		{"api.example.com", ""},
		{"localhost", ""},
		{"", ""},
	}

	for _, test := range tests {
		req := httptest.NewRequest("GET", "/test", nil)
		req.Host = test.host

		result := middleware.extractFromSubdomain(req)
		if result != test.expected {
			t.Errorf("For host '%s', expected '%s', got '%s'", test.host, test.expected, result)
		}
	}
}

func TestTenantContextMiddleware_ExtractFromPath(t *testing.T) {
	middleware := NewTenantContextMiddleware(NewMockTenantRepository())

	tests := []struct {
		path     string
		expected string
	}{
		{"/api/v1/tenants/tenant1/orders", "tenant1"},
		{"/tenants/tenant2/locations", "tenant2"},
		{"/api/orders", ""},
		{"/tenants/", ""},
		{"", ""},
	}

	for _, test := range tests {
		path := test.path
		if path == "" {
			path = "/"
		}
		req := httptest.NewRequest("GET", path, nil)

		result := middleware.extractFromPath(req)
		if result != test.expected {
			t.Errorf("For path '%s', expected '%s', got '%s'", test.path, test.expected, result)
		}
	}
}

func TestTenantJWTExtractor_ExtractTokenFromHeader(t *testing.T) {
	extractor := &TenantJWTExtractor{}

	tests := []struct {
		header      string
		expectError bool
		expected    string
	}{
		{"Bearer valid-token", false, "valid-token"},
		{"Bearer ", true, ""},
		{"Invalid format", true, ""},
		{"", true, ""},
	}

	for _, test := range tests {
		req := httptest.NewRequest("GET", "/test", nil)
		if test.header != "" {
			req.Header.Set("Authorization", test.header)
		}

		result, err := extractor.extractTokenFromHeader(req)

		if test.expectError && err == nil {
			t.Errorf("Expected error for header '%s', but got none", test.header)
		}

		if !test.expectError && err != nil {
			t.Errorf("Unexpected error for header '%s': %v", test.header, err)
		}

		if !test.expectError && result != test.expected {
			t.Errorf("For header '%s', expected '%s', got '%s'", test.header, test.expected, result)
		}
	}
}

func TestWriteJSONErrorResponse(t *testing.T) {
	w := httptest.NewRecorder()

	writeJSONErrorResponse(w, http.StatusUnauthorized, "Unauthorized", "Test error message")

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status 401, got %d", w.Code)
	}

	contentType := w.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("Expected Content-Type 'application/json', got '%s'", contentType)
	}

	body := w.Body.String()
	if body == "" {
		t.Error("Expected JSON response body, got empty string")
	}
}
