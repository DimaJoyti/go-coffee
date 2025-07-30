# 🛡️ 6: Security & Authentication

## 📋 Overview

Master security patterns and authentication systems through Go Coffee's comprehensive security architecture. This covers authentication systems, authorization models, encryption, threat protection, and compliance frameworks.

## 🎯 Learning Objectives

By the end of this phase, you will:
- Design secure authentication and authorization systems
- Implement encryption and data protection strategies
- Build threat detection and prevention mechanisms
- Understand compliance requirements (GDPR, PCI-DSS)
- Master API security and rate limiting
- Analyze Go Coffee's security implementations

---

## 📖 6.1 Authentication Systems

### Core Concepts

#### Authentication Methods
- **JWT (JSON Web Tokens)**: Stateless token-based authentication
- **Session-Based**: Server-side session management
- **OAuth 2.0**: Third-party authentication delegation
- **Multi-Factor Authentication (MFA)**: Additional security layers
- **Biometric Authentication**: Fingerprint, face recognition
- **Certificate-Based**: X.509 certificates for service authentication

#### Token Management
- **Token Generation**: Secure random token creation
- **Token Validation**: Signature verification and expiration
- **Token Refresh**: Seamless token renewal
- **Token Revocation**: Blacklisting and invalidation
- **Token Storage**: Secure client-side storage

### 🔍 Go Coffee Analysis

#### Study JWT Authentication Implementation

<augment_code_snippet path="auth-service/internal/auth/jwt_service.go" mode="EXCERPT">
````go
type JWTService struct {
    secretKey     []byte
    issuer        string
    expiration    time.Duration
    refreshExpiry time.Duration
    logger        *slog.Logger
}

func (js *JWTService) GenerateTokenPair(userID string, roles []string) (*TokenPair, error) {
    // Generate access token
    accessClaims := &CustomClaims{
        UserID: userID,
        Roles:  roles,
        RegisteredClaims: jwt.RegisteredClaims{
            Issuer:    js.issuer,
            Subject:   userID,
            ExpiresAt: jwt.NewNumericDate(time.Now().Add(js.expiration)),
            IssuedAt:  jwt.NewNumericDate(time.Now()),
            NotBefore: jwt.NewNumericDate(time.Now()),
            ID:        uuid.New().String(),
        },
    }
    
    accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
    accessTokenString, err := accessToken.SignedString(js.secretKey)
    if err != nil {
        return nil, fmt.Errorf("failed to sign access token: %w", err)
    }
    
    // Generate refresh token
    refreshClaims := &RefreshClaims{
        UserID: userID,
        RegisteredClaims: jwt.RegisteredClaims{
            Issuer:    js.issuer,
            Subject:   userID,
            ExpiresAt: jwt.NewNumericDate(time.Now().Add(js.refreshExpiry)),
            IssuedAt:  jwt.NewNumericDate(time.Now()),
            ID:        uuid.New().String(),
        },
    }
    
    refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
    refreshTokenString, err := refreshToken.SignedString(js.secretKey)
    if err != nil {
        return nil, fmt.Errorf("failed to sign refresh token: %w", err)
    }
    
    return &TokenPair{
        AccessToken:  accessTokenString,
        RefreshToken: refreshTokenString,
        ExpiresIn:    int(js.expiration.Seconds()),
        TokenType:    "Bearer",
    }, nil
}

func (js *JWTService) ValidateToken(tokenString string) (*CustomClaims, error) {
    token, err := jwt.ParseWithClaims(tokenString, &CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
        if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
            return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
        }
        return js.secretKey, nil
    })
    
    if err != nil {
        return nil, fmt.Errorf("failed to parse token: %w", err)
    }
    
    if claims, ok := token.Claims.(*CustomClaims); ok && token.Valid {
        // Check if token is blacklisted
        if js.isTokenBlacklisted(claims.ID) {
            return nil, errors.New("token has been revoked")
        }
        
        return claims, nil
    }
    
    return nil, errors.New("invalid token")
}
````
</augment_code_snippet>

### 🛠️ Hands-on Exercise 6.1: Implement Advanced Authentication

#### Step 1: Create Multi-Factor Authentication System
```go
// internal/auth/mfa_service.go
package auth

type MFAService struct {
    totpGenerator TOTPGenerator
    smsService    SMSService
    emailService  EmailService
    storage       MFAStorage
    logger        *slog.Logger
}

type MFAMethod string

const (
    MFAMethodTOTP  MFAMethod = "totp"
    MFAMethodSMS   MFAMethod = "sms"
    MFAMethodEmail MFAMethod = "email"
)

type MFAChallenge struct {
    ID          string    `json:"id"`
    UserID      string    `json:"user_id"`
    Method      MFAMethod `json:"method"`
    Challenge   string    `json:"challenge"`
    ExpiresAt   time.Time `json:"expires_at"`
    Attempts    int       `json:"attempts"`
    MaxAttempts int       `json:"max_attempts"`
    Verified    bool      `json:"verified"`
}

func (mfa *MFAService) InitiateMFA(userID string, preferredMethod MFAMethod) (*MFAChallenge, error) {
    // Get user's MFA settings
    settings, err := mfa.storage.GetMFASettings(userID)
    if err != nil {
        return nil, fmt.Errorf("failed to get MFA settings: %w", err)
    }
    
    // Validate preferred method is enabled
    if !settings.IsMethodEnabled(preferredMethod) {
        return nil, errors.New("MFA method not enabled for user")
    }
    
    challenge := &MFAChallenge{
        ID:          uuid.New().String(),
        UserID:      userID,
        Method:      preferredMethod,
        ExpiresAt:   time.Now().Add(5 * time.Minute),
        Attempts:    0,
        MaxAttempts: 3,
        Verified:    false,
    }
    
    switch preferredMethod {
    case MFAMethodTOTP:
        // TOTP doesn't need a challenge, user generates code
        challenge.Challenge = "totp_required"
        
    case MFAMethodSMS:
        code := mfa.generateRandomCode(6)
        challenge.Challenge = code
        
        if err := mfa.smsService.SendCode(settings.PhoneNumber, code); err != nil {
            return nil, fmt.Errorf("failed to send SMS code: %w", err)
        }
        
    case MFAMethodEmail:
        code := mfa.generateRandomCode(8)
        challenge.Challenge = code
        
        if err := mfa.emailService.SendCode(settings.Email, code); err != nil {
            return nil, fmt.Errorf("failed to send email code: %w", err)
        }
    }
    
    // Store challenge
    if err := mfa.storage.StoreMFAChallenge(challenge); err != nil {
        return nil, fmt.Errorf("failed to store MFA challenge: %w", err)
    }
    
    mfa.logger.Info("MFA challenge initiated", 
        "user_id", userID, 
        "method", preferredMethod,
        "challenge_id", challenge.ID)
    
    return challenge, nil
}

func (mfa *MFAService) VerifyMFA(challengeID, userCode string) (*MFAVerificationResult, error) {
    // Get challenge
    challenge, err := mfa.storage.GetMFAChallenge(challengeID)
    if err != nil {
        return nil, fmt.Errorf("failed to get MFA challenge: %w", err)
    }
    
    // Check if challenge is expired
    if time.Now().After(challenge.ExpiresAt) {
        return &MFAVerificationResult{
            Success: false,
            Error:   "MFA challenge expired",
        }, nil
    }
    
    // Check if too many attempts
    if challenge.Attempts >= challenge.MaxAttempts {
        return &MFAVerificationResult{
            Success: false,
            Error:   "Too many failed attempts",
        }, nil
    }
    
    // Increment attempts
    challenge.Attempts++
    
    var verified bool
    switch challenge.Method {
    case MFAMethodTOTP:
        verified, err = mfa.verifyTOTP(challenge.UserID, userCode)
    case MFAMethodSMS, MFAMethodEmail:
        verified = challenge.Challenge == userCode
    }
    
    if err != nil {
        return nil, fmt.Errorf("failed to verify MFA: %w", err)
    }
    
    challenge.Verified = verified
    
    // Update challenge
    if err := mfa.storage.UpdateMFAChallenge(challenge); err != nil {
        mfa.logger.Error("Failed to update MFA challenge", "error", err)
    }
    
    result := &MFAVerificationResult{
        Success:        verified,
        RemainingAttempts: challenge.MaxAttempts - challenge.Attempts,
    }
    
    if !verified {
        result.Error = "Invalid verification code"
    }
    
    mfa.logger.Info("MFA verification attempt", 
        "user_id", challenge.UserID,
        "method", challenge.Method,
        "success", verified,
        "attempts", challenge.Attempts)
    
    return result, nil
}

func (mfa *MFAService) verifyTOTP(userID, code string) (bool, error) {
    // Get user's TOTP secret
    secret, err := mfa.storage.GetTOTPSecret(userID)
    if err != nil {
        return false, fmt.Errorf("failed to get TOTP secret: %w", err)
    }
    
    return mfa.totpGenerator.Verify(code, secret), nil
}
```

#### Step 2: Implement OAuth 2.0 Integration
```go
// internal/auth/oauth_service.go
package auth

type OAuthService struct {
    providers map[string]OAuthProvider
    storage   OAuthStorage
    logger    *slog.Logger
}

type OAuthProvider interface {
    GetAuthURL(state string) string
    ExchangeCodeForToken(code string) (*OAuthToken, error)
    GetUserInfo(token *OAuthToken) (*OAuthUserInfo, error)
}

type GoogleOAuthProvider struct {
    clientID     string
    clientSecret string
    redirectURL  string
    httpClient   *http.Client
}

func (g *GoogleOAuthProvider) GetAuthURL(state string) string {
    params := url.Values{}
    params.Add("client_id", g.clientID)
    params.Add("redirect_uri", g.redirectURL)
    params.Add("scope", "openid email profile")
    params.Add("response_type", "code")
    params.Add("state", state)
    
    return "https://accounts.google.com/o/oauth2/v2/auth?" + params.Encode()
}

func (g *GoogleOAuthProvider) ExchangeCodeForToken(code string) (*OAuthToken, error) {
    data := url.Values{}
    data.Set("client_id", g.clientID)
    data.Set("client_secret", g.clientSecret)
    data.Set("code", code)
    data.Set("grant_type", "authorization_code")
    data.Set("redirect_uri", g.redirectURL)
    
    resp, err := g.httpClient.PostForm("https://oauth2.googleapis.com/token", data)
    if err != nil {
        return nil, fmt.Errorf("failed to exchange code: %w", err)
    }
    defer resp.Body.Close()
    
    if resp.StatusCode != http.StatusOK {
        return nil, fmt.Errorf("token exchange failed with status: %d", resp.StatusCode)
    }
    
    var token OAuthToken
    if err := json.NewDecoder(resp.Body).Decode(&token); err != nil {
        return nil, fmt.Errorf("failed to decode token response: %w", err)
    }
    
    return &token, nil
}

func (oauth *OAuthService) HandleCallback(provider, code, state string) (*AuthResult, error) {
    // Validate state parameter
    if !oauth.validateState(state) {
        return nil, errors.New("invalid state parameter")
    }
    
    // Get provider
    oauthProvider, exists := oauth.providers[provider]
    if !exists {
        return nil, fmt.Errorf("unsupported OAuth provider: %s", provider)
    }
    
    // Exchange code for token
    token, err := oauthProvider.ExchangeCodeForToken(code)
    if err != nil {
        return nil, fmt.Errorf("failed to exchange code for token: %w", err)
    }
    
    // Get user info
    userInfo, err := oauthProvider.GetUserInfo(token)
    if err != nil {
        return nil, fmt.Errorf("failed to get user info: %w", err)
    }
    
    // Find or create user
    user, err := oauth.findOrCreateUser(userInfo, provider)
    if err != nil {
        return nil, fmt.Errorf("failed to find or create user: %w", err)
    }
    
    // Store OAuth token
    if err := oauth.storage.StoreOAuthToken(user.ID, provider, token); err != nil {
        oauth.logger.Error("Failed to store OAuth token", "error", err)
    }
    
    oauth.logger.Info("OAuth authentication successful", 
        "user_id", user.ID,
        "provider", provider,
        "email", userInfo.Email)
    
    return &AuthResult{
        User:    user,
        Success: true,
    }, nil
}
```

### 💡 Practice Question 6.1
**"Design an authentication system for Go Coffee that supports multiple authentication methods and can scale to 10M users globally."**

**Solution Framework:**
1. **Multi-Method Support**
   - JWT for API authentication
   - OAuth 2.0 for social login
   - MFA for enhanced security
   - Certificate-based for service-to-service

2. **Scalability Considerations**
   - Stateless JWT tokens
   - Distributed session storage
   - Token caching and validation
   - Geographic distribution

3. **Security Features**
   - Token rotation and revocation
   - Rate limiting and brute force protection
   - Secure token storage
   - Audit logging and monitoring

---

## 📖 6.2 Authorization & Access Control

### Core Concepts

#### Authorization Models
- **Role-Based Access Control (RBAC)**: Users assigned to roles with permissions
- **Attribute-Based Access Control (ABAC)**: Context-aware access decisions
- **Access Control Lists (ACL)**: Direct user-resource permissions
- **Policy-Based Access Control**: Rule-based access decisions

#### Permission Systems
- **Hierarchical Permissions**: Nested permission structures
- **Resource-Based Permissions**: Permissions tied to specific resources
- **Action-Based Permissions**: Permissions for specific operations
- **Context-Aware Permissions**: Time, location, device-based access

### 🔍 Go Coffee Analysis

#### Study RBAC Implementation

<augment_code_snippet path="auth-service/internal/authorization/rbac_service.go" mode="EXCERPT">
````go
type RBACService struct {
    roleRepo       RoleRepository
    permissionRepo PermissionRepository
    userRoleRepo   UserRoleRepository
    cache          Cache
    logger         *slog.Logger
}

type Role struct {
    ID          string       `json:"id"`
    Name        string       `json:"name"`
    Description string       `json:"description"`
    Permissions []Permission `json:"permissions"`
    CreatedAt   time.Time    `json:"created_at"`
    UpdatedAt   time.Time    `json:"updated_at"`
}

type Permission struct {
    ID       string `json:"id"`
    Resource string `json:"resource"`
    Action   string `json:"action"`
    Scope    string `json:"scope,omitempty"`
}

func (rbac *RBACService) CheckPermission(userID, resource, action string, context *AuthContext) (bool, error) {
    // Get user roles from cache or database
    cacheKey := fmt.Sprintf("user_roles:%s", userID)
    var userRoles []Role
    
    if cached := rbac.cache.Get(cacheKey); cached != nil {
        userRoles = cached.([]Role)
    } else {
        roles, err := rbac.userRoleRepo.GetUserRoles(userID)
        if err != nil {
            return false, fmt.Errorf("failed to get user roles: %w", err)
        }
        userRoles = roles
        rbac.cache.Set(cacheKey, userRoles, 15*time.Minute)
    }
    
    // Check permissions for each role
    for _, role := range userRoles {
        for _, permission := range role.Permissions {
            if rbac.matchesPermission(permission, resource, action, context) {
                rbac.logger.Debug("Permission granted", 
                    "user_id", userID,
                    "role", role.Name,
                    "resource", resource,
                    "action", action)
                return true, nil
            }
        }
    }
    
    rbac.logger.Warn("Permission denied", 
        "user_id", userID,
        "resource", resource,
        "action", action)
    
    return false, nil
}

func (rbac *RBACService) matchesPermission(permission Permission, resource, action string, context *AuthContext) bool {
    // Check resource match
    if !rbac.matchesResource(permission.Resource, resource) {
        return false
    }
    
    // Check action match
    if !rbac.matchesAction(permission.Action, action) {
        return false
    }
    
    // Check scope constraints
    if permission.Scope != "" {
        return rbac.matchesScope(permission.Scope, context)
    }
    
    return true
}

func (rbac *RBACService) matchesResource(permissionResource, requestedResource string) bool {
    // Support wildcard matching
    if permissionResource == "*" {
        return true
    }
    
    // Support hierarchical resources (e.g., "coffee:*" matches "coffee:orders")
    if strings.HasSuffix(permissionResource, ":*") {
        prefix := strings.TrimSuffix(permissionResource, ":*")
        return strings.HasPrefix(requestedResource, prefix+":")
    }
    
    return permissionResource == requestedResource
}
````
</augment_code_snippet>

### 🛠️ Hands-on Exercise 6.2: Implement Advanced Authorization

#### Step 1: Create Policy-Based Access Control
```go
// internal/authorization/policy_engine.go
package authorization

type PolicyEngine struct {
    policies []Policy
    evaluator PolicyEvaluator
    logger   *slog.Logger
}

type Policy struct {
    ID          string            `json:"id"`
    Name        string            `json:"name"`
    Description string            `json:"description"`
    Rules       []PolicyRule      `json:"rules"`
    Effect      PolicyEffect      `json:"effect"`
    Conditions  map[string]string `json:"conditions"`
    Priority    int               `json:"priority"`
}

type PolicyRule struct {
    Resource   string   `json:"resource"`
    Actions    []string `json:"actions"`
    Principals []string `json:"principals"`
    Conditions []string `json:"conditions"`
}

type PolicyEffect string

const (
    PolicyEffectAllow PolicyEffect = "allow"
    PolicyEffectDeny  PolicyEffect = "deny"
)

func (pe *PolicyEngine) EvaluateAccess(request *AccessRequest) (*AccessDecision, error) {
    var applicablePolicies []Policy
    
    // Find applicable policies
    for _, policy := range pe.policies {
        if pe.isPolicyApplicable(policy, request) {
            applicablePolicies = append(applicablePolicies, policy)
        }
    }
    
    // Sort by priority (higher priority first)
    sort.Slice(applicablePolicies, func(i, j int) bool {
        return applicablePolicies[i].Priority > applicablePolicies[j].Priority
    })
    
    // Evaluate policies in priority order
    for _, policy := range applicablePolicies {
        decision, err := pe.evaluatePolicy(policy, request)
        if err != nil {
            pe.logger.Error("Failed to evaluate policy", 
                "policy_id", policy.ID, "error", err)
            continue
        }
        
        if decision.Effect == PolicyEffectDeny {
            // Explicit deny takes precedence
            return &AccessDecision{
                Effect:   PolicyEffectDeny,
                PolicyID: policy.ID,
                Reason:   decision.Reason,
            }, nil
        }
        
        if decision.Effect == PolicyEffectAllow {
            return &AccessDecision{
                Effect:   PolicyEffectAllow,
                PolicyID: policy.ID,
                Reason:   decision.Reason,
            }, nil
        }
    }
    
    // Default deny if no explicit allow
    return &AccessDecision{
        Effect: PolicyEffectDeny,
        Reason: "No applicable policy found",
    }, nil
}

func (pe *PolicyEngine) evaluatePolicy(policy Policy, request *AccessRequest) (*AccessDecision, error) {
    // Check all rules in the policy
    for _, rule := range policy.Rules {
        if pe.matchesRule(rule, request) {
            // Evaluate conditions
            if pe.evaluateConditions(policy.Conditions, request) {
                return &AccessDecision{
                    Effect: policy.Effect,
                    Reason: fmt.Sprintf("Matched policy: %s", policy.Name),
                }, nil
            }
        }
    }
    
    return &AccessDecision{
        Effect: PolicyEffectDeny,
        Reason: "Policy rules not satisfied",
    }, nil
}

func (pe *PolicyEngine) evaluateConditions(conditions map[string]string, request *AccessRequest) bool {
    for key, expectedValue := range conditions {
        actualValue := pe.getContextValue(key, request)
        if !pe.matchesCondition(expectedValue, actualValue) {
            return false
        }
    }
    return true
}

func (pe *PolicyEngine) matchesCondition(expected, actual string) bool {
    // Support various condition operators
    if strings.HasPrefix(expected, ">=") {
        return pe.compareNumeric(actual, strings.TrimPrefix(expected, ">="), ">=")
    }
    if strings.HasPrefix(expected, "<=") {
        return pe.compareNumeric(actual, strings.TrimPrefix(expected, "<="), "<=")
    }
    if strings.HasPrefix(expected, ">") {
        return pe.compareNumeric(actual, strings.TrimPrefix(expected, ">"), ">")
    }
    if strings.HasPrefix(expected, "<") {
        return pe.compareNumeric(actual, strings.TrimPrefix(expected, "<"), "<")
    }
    if strings.HasPrefix(expected, "in:") {
        values := strings.Split(strings.TrimPrefix(expected, "in:"), ",")
        return pe.containsValue(values, actual)
    }
    if strings.HasPrefix(expected, "regex:") {
        pattern := strings.TrimPrefix(expected, "regex:")
        matched, _ := regexp.MatchString(pattern, actual)
        return matched
    }
    
    // Default: exact match
    return expected == actual
}
```

#### Step 2: Implement Resource-Based Authorization
```go
// internal/authorization/resource_authz.go
package authorization

type ResourceAuthzService struct {
    aclRepo    ACLRepository
    policyRepo PolicyRepository
    cache      Cache
    logger     *slog.Logger
}

type ResourceACL struct {
    ResourceID   string            `json:"resource_id"`
    ResourceType string            `json:"resource_type"`
    Owner        string            `json:"owner"`
    Permissions  map[string][]string `json:"permissions"` // principal -> actions
    Inherited    []string          `json:"inherited"`    // parent resource IDs
    CreatedAt    time.Time         `json:"created_at"`
    UpdatedAt    time.Time         `json:"updated_at"`
}

func (ras *ResourceAuthzService) CheckResourceAccess(userID, resourceID, action string) (bool, error) {
    // Get resource ACL
    acl, err := ras.getResourceACL(resourceID)
    if err != nil {
        return false, fmt.Errorf("failed to get resource ACL: %w", err)
    }
    
    // Check direct permissions
    if ras.hasDirectPermission(acl, userID, action) {
        return true, nil
    }
    
    // Check role-based permissions
    userRoles, err := ras.getUserRoles(userID)
    if err != nil {
        return false, fmt.Errorf("failed to get user roles: %w", err)
    }
    
    for _, role := range userRoles {
        if ras.hasDirectPermission(acl, "role:"+role, action) {
            return true, nil
        }
    }
    
    // Check inherited permissions
    for _, parentResourceID := range acl.Inherited {
        hasAccess, err := ras.CheckResourceAccess(userID, parentResourceID, action)
        if err != nil {
            ras.logger.Error("Failed to check inherited permission", 
                "parent_resource", parentResourceID, "error", err)
            continue
        }
        if hasAccess {
            return true, nil
        }
    }
    
    // Check if user is owner
    if acl.Owner == userID {
        return ras.ownerHasPermission(action), nil
    }
    
    return false, nil
}

func (ras *ResourceAuthzService) GrantResourcePermission(resourceID, principal, action string, grantedBy string) error {
    // Verify granter has permission to grant
    canGrant, err := ras.CheckResourceAccess(grantedBy, resourceID, "grant")
    if err != nil {
        return fmt.Errorf("failed to check grant permission: %w", err)
    }
    if !canGrant {
        return errors.New("insufficient permissions to grant access")
    }
    
    // Get current ACL
    acl, err := ras.getResourceACL(resourceID)
    if err != nil {
        return fmt.Errorf("failed to get resource ACL: %w", err)
    }
    
    // Add permission
    if acl.Permissions == nil {
        acl.Permissions = make(map[string][]string)
    }
    
    actions := acl.Permissions[principal]
    if !contains(actions, action) {
        acl.Permissions[principal] = append(actions, action)
    }
    
    acl.UpdatedAt = time.Now()
    
    // Save ACL
    if err := ras.aclRepo.UpdateResourceACL(acl); err != nil {
        return fmt.Errorf("failed to update resource ACL: %w", err)
    }
    
    // Invalidate cache
    ras.cache.Delete(fmt.Sprintf("resource_acl:%s", resourceID))
    
    ras.logger.Info("Resource permission granted", 
        "resource_id", resourceID,
        "principal", principal,
        "action", action,
        "granted_by", grantedBy)
    
    return nil
}
```

### 💡 Practice Question 6.2
**"Design an authorization system for Go Coffee that supports fine-grained permissions for coffee shops, orders, and user data with multi-tenant isolation."**

**Solution Framework:**
1. **Multi-Tenant Architecture**
   - Tenant-based resource isolation
   - Hierarchical permission inheritance
   - Cross-tenant access controls
   - Tenant-specific policies

2. **Fine-Grained Permissions**
   - Resource-based access control
   - Action-specific permissions
   - Context-aware authorization
   - Dynamic permission evaluation

3. **Scalability Considerations**
   - Permission caching strategies
   - Distributed policy evaluation
   - Efficient ACL storage
   - Real-time permission updates

---

## 📖 6.3 Encryption & Data Protection

### Core Concepts

#### Encryption Types
- **Symmetric Encryption**: Same key for encryption/decryption (AES)
- **Asymmetric Encryption**: Public/private key pairs (RSA, ECC)
- **Hashing**: One-way functions for data integrity (SHA-256)
- **Digital Signatures**: Authentication and non-repudiation

#### Data Protection Strategies
- **Encryption at Rest**: Database and file encryption
- **Encryption in Transit**: TLS/SSL for network communication
- **End-to-End Encryption**: Client-to-client encryption
- **Key Management**: Secure key generation, storage, rotation

### 🔍 Go Coffee Analysis

#### Study Encryption Implementation

<augment_code_snippet path="pkg/crypto/encryption_service.go" mode="EXCERPT">
````go
type EncryptionService struct {
    keyManager KeyManager
    logger     *slog.Logger
}

func (es *EncryptionService) EncryptSensitiveData(data []byte, keyID string) (*EncryptedData, error) {
    // Get encryption key
    key, err := es.keyManager.GetKey(keyID)
    if err != nil {
        return nil, fmt.Errorf("failed to get encryption key: %w", err)
    }
    
    // Generate random IV
    iv := make([]byte, aes.BlockSize)
    if _, err := rand.Read(iv); err != nil {
        return nil, fmt.Errorf("failed to generate IV: %w", err)
    }
    
    // Create cipher
    block, err := aes.NewCipher(key.Data)
    if err != nil {
        return nil, fmt.Errorf("failed to create cipher: %w", err)
    }
    
    // Encrypt data using AES-GCM
    gcm, err := cipher.NewGCM(block)
    if err != nil {
        return nil, fmt.Errorf("failed to create GCM: %w", err)
    }
    
    ciphertext := gcm.Seal(nil, iv, data, nil)
    
    return &EncryptedData{
        KeyID:      keyID,
        IV:         iv,
        Ciphertext: ciphertext,
        Algorithm:  "AES-256-GCM",
        CreatedAt:  time.Now(),
    }, nil
}

func (es *EncryptionService) DecryptSensitiveData(encData *EncryptedData) ([]byte, error) {
    // Get decryption key
    key, err := es.keyManager.GetKey(encData.KeyID)
    if err != nil {
        return nil, fmt.Errorf("failed to get decryption key: %w", err)
    }
    
    // Create cipher
    block, err := aes.NewCipher(key.Data)
    if err != nil {
        return nil, fmt.Errorf("failed to create cipher: %w", err)
    }
    
    // Decrypt data
    gcm, err := cipher.NewGCM(block)
    if err != nil {
        return nil, fmt.Errorf("failed to create GCM: %w", err)
    }
    
    plaintext, err := gcm.Open(nil, encData.IV, encData.Ciphertext, nil)
    if err != nil {
        return nil, fmt.Errorf("failed to decrypt data: %w", err)
    }
    
    return plaintext, nil
}
````
</augment_code_snippet>

### 🛠️ Hands-on Exercise 6.3: Implement Comprehensive Data Protection

#### Step 1: Create Key Management System
```go
// internal/crypto/key_manager.go
package crypto

type KeyManager struct {
    storage    KeyStorage
    hsm        HSMClient
    rotationPolicy RotationPolicy
    logger     *slog.Logger
}

type EncryptionKey struct {
    ID        string    `json:"id"`
    Version   int       `json:"version"`
    Algorithm string    `json:"algorithm"`
    KeySize   int       `json:"key_size"`
    Data      []byte    `json:"-"` // Never serialize key data
    CreatedAt time.Time `json:"created_at"`
    ExpiresAt time.Time `json:"expires_at"`
    Status    KeyStatus `json:"status"`
}

type KeyStatus string

const (
    KeyStatusActive     KeyStatus = "active"
    KeyStatusRotating   KeyStatus = "rotating"
    KeyStatusDeprecated KeyStatus = "deprecated"
    KeyStatusRevoked    KeyStatus = "revoked"
)

func (km *KeyManager) GenerateKey(algorithm string, keySize int) (*EncryptionKey, error) {
    keyID := uuid.New().String()
    
    var keyData []byte
    var err error
    
    switch algorithm {
    case "AES-256":
        keyData = make([]byte, 32) // 256 bits
        if _, err := rand.Read(keyData); err != nil {
            return nil, fmt.Errorf("failed to generate AES key: %w", err)
        }
        
    case "RSA-2048":
        privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
        if err != nil {
            return nil, fmt.Errorf("failed to generate RSA key: %w", err)
        }
        keyData = x509.MarshalPKCS1PrivateKey(privateKey)
        
    default:
        return nil, fmt.Errorf("unsupported algorithm: %s", algorithm)
    }
    
    key := &EncryptionKey{
        ID:        keyID,
        Version:   1,
        Algorithm: algorithm,
        KeySize:   keySize,
        Data:      keyData,
        CreatedAt: time.Now(),
        ExpiresAt: time.Now().Add(km.rotationPolicy.KeyLifetime),
        Status:    KeyStatusActive,
    }
    
    // Store key securely
    if err := km.storage.StoreKey(key); err != nil {
        return nil, fmt.Errorf("failed to store key: %w", err)
    }
    
    // Store in HSM if available
    if km.hsm != nil {
        if err := km.hsm.StoreKey(key); err != nil {
            km.logger.Error("Failed to store key in HSM", "key_id", keyID, "error", err)
        }
    }
    
    km.logger.Info("Generated new encryption key", 
        "key_id", keyID,
        "algorithm", algorithm,
        "key_size", keySize)
    
    return key, nil
}

func (km *KeyManager) RotateKey(keyID string) (*EncryptionKey, error) {
    // Get current key
    currentKey, err := km.storage.GetKey(keyID)
    if err != nil {
        return nil, fmt.Errorf("failed to get current key: %w", err)
    }
    
    // Mark current key as rotating
    currentKey.Status = KeyStatusRotating
    if err := km.storage.UpdateKey(currentKey); err != nil {
        return nil, fmt.Errorf("failed to update current key status: %w", err)
    }
    
    // Generate new key version
    newKey := &EncryptionKey{
        ID:        keyID,
        Version:   currentKey.Version + 1,
        Algorithm: currentKey.Algorithm,
        KeySize:   currentKey.KeySize,
        CreatedAt: time.Now(),
        ExpiresAt: time.Now().Add(km.rotationPolicy.KeyLifetime),
        Status:    KeyStatusActive,
    }
    
    // Generate new key data
    switch newKey.Algorithm {
    case "AES-256":
        newKey.Data = make([]byte, 32)
        if _, err := rand.Read(newKey.Data); err != nil {
            return nil, fmt.Errorf("failed to generate new key data: %w", err)
        }
    }
    
    // Store new key version
    if err := km.storage.StoreKey(newKey); err != nil {
        return nil, fmt.Errorf("failed to store new key version: %w", err)
    }
    
    // Schedule deprecation of old key
    go km.scheduleKeyDeprecation(currentKey, km.rotationPolicy.DeprecationDelay)
    
    km.logger.Info("Key rotated successfully", 
        "key_id", keyID,
        "old_version", currentKey.Version,
        "new_version", newKey.Version)
    
    return newKey, nil
}

func (km *KeyManager) scheduleKeyDeprecation(key *EncryptionKey, delay time.Duration) {
    time.Sleep(delay)
    
    key.Status = KeyStatusDeprecated
    if err := km.storage.UpdateKey(key); err != nil {
        km.logger.Error("Failed to deprecate key", "key_id", key.ID, "version", key.Version, "error", err)
    } else {
        km.logger.Info("Key deprecated", "key_id", key.ID, "version", key.Version)
    }
}
```

### 💡 Practice Question 6.3
**"Design an encryption strategy for Go Coffee that protects customer payment data, personal information, and business secrets while maintaining performance and compliance."**

**Solution Framework:**
1. **Data Classification**
   - Public data: No encryption needed
   - Internal data: Encryption at rest
   - Sensitive data: End-to-end encryption
   - Payment data: PCI-DSS compliant encryption

2. **Encryption Strategy**
   - AES-256 for symmetric encryption
   - RSA-2048 for key exchange
   - TLS 1.3 for data in transit
   - Hardware Security Modules (HSM) for key storage

3. **Key Management**
   - Automated key rotation
   - Secure key distribution
   - Key escrow and recovery
   - Audit logging and compliance

---

## 🎯 6 Completion Checklist

### Knowledge Mastery
- [ ] Understand authentication methods and token management
- [ ] Can design authorization systems with RBAC and ABAC
- [ ] Know encryption techniques and key management
- [ ] Understand threat modeling and security patterns
- [ ] Can implement compliance requirements

### Practical Skills
- [ ] Can implement JWT-based authentication with MFA
- [ ] Can build fine-grained authorization systems
- [ ] Can design encryption and key management solutions
- [ ] Can implement API security and rate limiting
- [ ] Can handle security monitoring and incident response

### Go Coffee Analysis
- [ ] Analyzed authentication and authorization implementations
- [ ] Studied encryption and data protection strategies
- [ ] Examined security monitoring and threat detection
- [ ] Understood compliance and regulatory requirements
- [ ] Identified security optimization opportunities

###  Readiness
- [ ] Can design secure authentication systems
- [ ] Can explain authorization models and trade-offs
- [ ] Can implement encryption and data protection
- [ ] Can handle security threat scenarios
- [ ] Can discuss compliance and regulatory requirements

---

## 🚀 Next Steps

Ready for **7: Monitoring & Observability**:
- Logging strategies and centralized log management
- Metrics collection and monitoring systems
- Distributed tracing and performance monitoring
- Alerting and incident response
- Business metrics and analytics

**Excellent progress on mastering security and authentication! 🎉**
