# 🤝 Contributing to Go Coffee

<div align="center">

![Contributing](https://img.shields.io/badge/Contributing-Welcome-brightgreen?style=for-the-badge)
![Code of Conduct](https://img.shields.io/badge/Code%20of%20Conduct-Enforced-blue?style=for-the-badge)
![License](https://img.shields.io/badge/License-MIT-green?style=for-the-badge)

**We welcome contributions from developers of all skill levels!**

</div>

---

## 🌟 Ways to Contribute

<table>
<tr>
<td width="25%">

### 🐛 **Bug Reports**
- Report bugs and issues
- Provide reproduction steps
- Suggest fixes
- Test bug fixes

</td>
<td width="25%">

### ✨ **Feature Requests**
- Propose new features
- Discuss implementation
- Create feature specs
- Implement features

</td>
<td width="25%">

### 📖 **Documentation**
- Improve existing docs
- Write tutorials
- Create examples
- Fix typos

</td>
<td width="25%">

### 🧪 **Testing**
- Write unit tests
- Create integration tests
- Improve test coverage
- Performance testing

</td>
</tr>
</table>

---

## 🚀 Getting Started

### 📋 Prerequisites

- **Go 1.24+** - [Install Go](https://golang.org/doc/install)
- **Docker** - [Install Docker](https://docs.docker.com/get-docker/)
- **Git** - [Install Git](https://git-scm.com/downloads)
- **Make** - Usually pre-installed on Unix systems

### 🔧 Development Setup

1. **Fork the repository**
   ```bash
   # Click the "Fork" button on GitHub
   ```

2. **Clone your fork**
   ```bash
   git clone https://github.com/YOUR_USERNAME/go-coffee.git
   cd go-coffee
   ```

3. **Add upstream remote**
   ```bash
   git remote add upstream https://github.com/DimaJoyti/go-coffee.git
   ```

4. **Install dependencies**
   ```bash
   go mod download
   make install-tools
   ```

5. **Start development environment**
   ```bash
   # Start infrastructure services
   docker-compose -f docker-compose.dev.yml up -d
   
   # Verify setup
   make health-check
   ```

---

## 📝 Development Workflow

### 🌿 Creating a Feature Branch

```bash
# Update your main branch
git checkout main
git pull upstream main

# Create a feature branch
git checkout -b feature/your-feature-name

# Or for bug fixes
git checkout -b fix/bug-description
```

### 🔨 Making Changes

1. **Write your code**
   - Follow our [coding standards](#-coding-standards)
   - Add tests for new functionality
   - Update documentation as needed

2. **Test your changes**
   ```bash
   # Run all tests
   make test-all
   
   # Run specific service tests
   make -f Makefile.auth test
   
   # Check test coverage
   make test-coverage
   ```

3. **Lint your code**
   ```bash
   # Run linters
   make lint-all
   
   # Format code
   make format-all
   ```

4. **Commit your changes**
   ```bash
   # Stage your changes
   git add .
   
   # Commit with descriptive message
   git commit -m "feat: add user registration endpoint
   
   - Add POST /api/v1/auth/register endpoint
   - Implement password validation
   - Add comprehensive tests
   - Update API documentation
   
   Closes #123"
   ```

### 📤 Submitting Changes

1. **Push to your fork**
   ```bash
   git push origin feature/your-feature-name
   ```

2. **Create a Pull Request**
   - Go to your fork on GitHub
   - Click "New Pull Request"
   - Fill out the PR template
   - Link related issues

3. **Respond to feedback**
   - Address review comments
   - Update your branch as needed
   - Keep the conversation constructive

---

## 📏 Coding Standards

### 🎯 Go Code Style

```go
// ✅ Good: Clear function with proper error handling
func (s *AuthService) RegisterUser(ctx context.Context, req *RegisterRequest) (*RegisterResponse, error) {
    if err := s.validateRequest(req); err != nil {
        return nil, fmt.Errorf("validation failed: %w", err)
    }
    
    user, err := s.createUser(ctx, req)
    if err != nil {
        return nil, fmt.Errorf("failed to create user: %w", err)
    }
    
    return s.buildResponse(user), nil
}

// ❌ Bad: No error handling, unclear naming
func (s *AuthService) Reg(r *RegisterRequest) *RegisterResponse {
    u := s.create(r)
    return s.build(u)
}
```

### 📋 Code Guidelines

<table>
<tr>
<td width="50%">

**✅ Do:**
- Use descriptive variable names
- Handle all errors explicitly
- Write comprehensive tests
- Add godoc comments for public functions
- Follow Go naming conventions
- Use interfaces for dependencies
- Keep functions small and focused

</td>
<td width="50%">

**❌ Don't:**
- Ignore errors
- Use global variables
- Write functions longer than 50 lines
- Skip tests for new code
- Use magic numbers
- Panic in library code
- Mix business logic with transport

</td>
</tr>
</table>

### 🧪 Testing Standards

```go
// ✅ Good: Table-driven test with clear scenarios
func TestAuthService_RegisterUser(t *testing.T) {
    tests := []struct {
        name    string
        request *RegisterRequest
        setup   func(*MockUserRepo)
        want    *RegisterResponse
        wantErr bool
    }{
        {
            name: "successful registration",
            request: &RegisterRequest{
                Email:    "test@example.com",
                Password: "SecurePass123!",
            },
            setup: func(repo *MockUserRepo) {
                repo.EXPECT().CreateUser(gomock.Any(), gomock.Any()).Return(nil)
            },
            want: &RegisterResponse{
                User: &UserDTO{
                    Email: "test@example.com",
                    Role:  "user",
                },
            },
            wantErr: false,
        },
        // More test cases...
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Test implementation
        })
    }
}
```

---

## 📝 Commit Message Guidelines

### 🎯 Commit Format

```
<type>(<scope>): <subject>

<body>

<footer>
```

### 📋 Types

- **feat**: New feature
- **fix**: Bug fix
- **docs**: Documentation changes
- **style**: Code style changes (formatting, etc.)
- **refactor**: Code refactoring
- **test**: Adding or updating tests
- **chore**: Maintenance tasks

### 💡 Examples

```bash
# Feature
feat(auth): add JWT refresh token endpoint

Implement automatic token refresh functionality:
- Add POST /api/v1/auth/refresh endpoint
- Validate refresh token expiration
- Generate new token pair
- Update session with new tokens

Closes #45

# Bug fix
fix(auth): prevent account lockout bypass

Fix security vulnerability where users could bypass
account lockout by using different user agents.

- Add IP-based lockout tracking
- Implement progressive delay
- Add security event logging

Fixes #67

# Documentation
docs(api): update authentication flow examples

- Add cURL examples for all auth endpoints
- Update Postman collection
- Fix typos in API documentation
```

---

## 🔍 Pull Request Guidelines

### 📋 PR Checklist

Before submitting your PR, ensure:

- [ ] **Tests pass**: All existing and new tests pass
- [ ] **Linting passes**: Code follows style guidelines
- [ ] **Documentation updated**: Relevant docs are updated
- [ ] **Changelog updated**: Add entry to CHANGELOG.md
- [ ] **Breaking changes noted**: Document any breaking changes
- [ ] **Security reviewed**: No security vulnerabilities introduced

### 📝 PR Template

```markdown
## 🎯 Description

Brief description of what this PR does.

## 🔄 Type of Change

- [ ] Bug fix (non-breaking change which fixes an issue)
- [ ] New feature (non-breaking change which adds functionality)
- [ ] Breaking change (fix or feature that would cause existing functionality to not work as expected)
- [ ] Documentation update

## 🧪 Testing

- [ ] Unit tests added/updated
- [ ] Integration tests added/updated
- [ ] Manual testing performed

## 📋 Checklist

- [ ] My code follows the style guidelines
- [ ] I have performed a self-review of my code
- [ ] I have commented my code, particularly in hard-to-understand areas
- [ ] I have made corresponding changes to the documentation
- [ ] My changes generate no new warnings
- [ ] I have added tests that prove my fix is effective or that my feature works
- [ ] New and existing unit tests pass locally with my changes

## 🔗 Related Issues

Closes #(issue number)
```

---

## 🏗️ Project Structure

### 📁 Directory Layout

```
go-coffee/
├── cmd/                    # Application entry points
│   ├── auth-service/       # Auth service main
│   ├── order-service/      # Order service main
│   └── kitchen-service/    # Kitchen service main
├── internal/               # Private application code
│   ├── auth/              # Auth domain
│   │   ├── domain/        # Entities and business rules
│   │   ├── application/   # Use cases and DTOs
│   │   ├── infrastructure/# External concerns
│   │   └── transport/     # API handlers
│   ├── order/             # Order domain
│   └── kitchen/           # Kitchen domain
├── pkg/                   # Shared libraries
├── docs/                  # Documentation
├── scripts/               # Build and deployment scripts
└── tests/                 # Integration tests
```

### 🎯 Adding New Services

1. **Create service directory**
   ```bash
   mkdir -p cmd/new-service
   mkdir -p internal/new-service/{domain,application,infrastructure,transport}
   ```

2. **Follow Clean Architecture**
   - **Domain**: Business entities and rules
   - **Application**: Use cases and interfaces
   - **Infrastructure**: External dependencies
   - **Transport**: API handlers and middleware

3. **Add service documentation**
   ```bash
   mkdir -p docs/new-service
   # Create README.md, api-reference.md, etc.
   ```

---

## 🧪 Testing Guidelines

### 🎯 Test Categories

<table>
<tr>
<td width="33%">

**🔬 Unit Tests**
- Test individual functions
- Mock external dependencies
- Fast execution
- High coverage

</td>
<td width="33%">

**🔗 Integration Tests**
- Test service interactions
- Use real databases
- Test API endpoints
- Verify data flow

</td>
<td width="33%">

**🌐 E2E Tests**
- Test complete workflows
- Use real services
- Simulate user scenarios
- Validate business logic

</td>
</tr>
</table>

### 📋 Testing Commands

```bash
# Run all tests
make test-all

# Run specific test types
make test-unit
make test-integration
make test-e2e

# Run tests with coverage
make test-coverage

# Run benchmarks
make benchmark

# Run tests for specific service
make -f Makefile.auth test
```

---

## 📖 Documentation Guidelines

### 📝 Documentation Types

1. **API Documentation**
   - OpenAPI specifications
   - Request/response examples
   - Error codes and messages

2. **Architecture Documentation**
   - System design diagrams
   - Component interactions
   - Data flow diagrams

3. **User Guides**
   - Getting started tutorials
   - Configuration guides
   - Deployment instructions

4. **Developer Guides**
   - Code organization
   - Development setup
   - Contributing guidelines

### 🎯 Writing Guidelines

- **Be clear and concise**
- **Use examples liberally**
- **Keep documentation up-to-date**
- **Include diagrams when helpful**
- **Test all code examples**

---

## 🆘 Getting Help

### 💬 Communication Channels

- **GitHub Issues**: Bug reports and feature requests
- **GitHub Discussions**: General questions and ideas
- **Email**: aws.inspiration@gmail.com for private matters

### 🤔 Before Asking for Help

1. **Search existing issues** and discussions
2. **Check the documentation**
3. **Try to reproduce** the issue
4. **Prepare a minimal example**

### 📝 Asking Good Questions

- **Be specific** about the problem
- **Include relevant details** (OS, Go version, etc.)
- **Provide code examples** when applicable
- **Describe what you expected** vs. what happened

---

## 🏆 Recognition

### 🌟 Contributors

We recognize contributors in several ways:

- **Contributors list** in README
- **Release notes** mention significant contributions
- **Special badges** for major contributors
- **Maintainer status** for consistent contributors

### 🎯 Becoming a Maintainer

Regular contributors may be invited to become maintainers:

- **Consistent quality contributions**
- **Good understanding of the codebase**
- **Helpful in reviews and discussions**
- **Follows project guidelines**

---

## 📄 Code of Conduct

### 🤝 Our Pledge

We are committed to making participation in our project a harassment-free experience for everyone, regardless of:

- Age, body size, disability, ethnicity
- Gender identity and expression
- Level of experience, nationality
- Personal appearance, race, religion
- Sexual identity and orientation

### 📋 Our Standards

**Positive behavior includes:**
- Using welcoming and inclusive language
- Being respectful of differing viewpoints
- Gracefully accepting constructive criticism
- Focusing on what is best for the community

**Unacceptable behavior includes:**
- Harassment, trolling, or insulting comments
- Public or private harassment
- Publishing others' private information
- Other conduct inappropriate in a professional setting

### 🚨 Enforcement

Instances of abusive, harassing, or otherwise unacceptable behavior may be reported to the project team at aws.inspiration@gmail.com.

---

<div align="center">

**Thank you for contributing to Go Coffee! 🙏**

**Together, we're building something amazing ☕**

[🏠 Back to README](README.md) • [📖 Documentation](docs/) • [🐛 Report Issue](https://github.com/DimaJoyti/go-coffee/issues)

</div>
