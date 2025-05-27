# Contributing to Go Coffee Web3 Platform

Thank you for your interest in contributing to Go Coffee! We welcome contributions from the community and are excited to work with you.

## 🌟 Ways to Contribute

- **🐛 Bug Reports** - Help us identify and fix issues
- **✨ Feature Requests** - Suggest new features and improvements
- **💻 Code Contributions** - Submit pull requests with fixes and features
- **📚 Documentation** - Improve our documentation and guides
- **🧪 Testing** - Help us test new features and find edge cases
- **🎨 Design** - Contribute to UI/UX improvements

## 🚀 Getting Started

### Prerequisites

- Go 1.21 or higher
- Docker and Docker Compose
- Git
- Basic knowledge of blockchain and DeFi concepts

### Development Setup

1. **Fork the repository**
   ```bash
   git clone https://github.com/YOUR_USERNAME/go-coffee.git
   cd go-coffee/web3-wallet-backend
   ```

2. **Install dependencies**
   ```bash
   make deps
   ```

3. **Set up environment**
   ```bash
   cp .env.example .env
   # Edit .env with your configuration
   ```

4. **Start development environment**
   ```bash
   docker-compose up -d postgres redis kafka
   make db-migrate
   ```

5. **Run tests**
   ```bash
   make test
   ```

## 📝 Development Guidelines

### Code Style

- Follow Go best practices and conventions
- Use `gofmt` and `goimports` for formatting
- Write clear, self-documenting code
- Add comments for complex logic
- Use meaningful variable and function names

### Testing

- Write unit tests for all new functionality
- Maintain test coverage above 80%
- Include integration tests for critical paths
- Test edge cases and error conditions
- Use table-driven tests where appropriate

### Git Workflow

1. **Create a feature branch**
   ```bash
   git checkout -b feature/your-feature-name
   ```

2. **Make your changes**
   - Write clean, well-tested code
   - Follow our coding standards
   - Update documentation as needed

3. **Commit your changes**
   ```bash
   git add .
   git commit -m "feat: add new coffee payment feature"
   ```

4. **Push and create PR**
   ```bash
   git push origin feature/your-feature-name
   ```

### Commit Message Format

We use [Conventional Commits](https://www.conventionalcommits.org/) format:

```
<type>[optional scope]: <description>

[optional body]

[optional footer(s)]
```

**Types:**
- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation changes
- `style`: Code style changes (formatting, etc.)
- `refactor`: Code refactoring
- `test`: Adding or updating tests
- `chore`: Maintenance tasks

**Examples:**
```
feat(wallet): add Solana wallet support
fix(defi): resolve Jupiter swap calculation error
docs(api): update trading endpoints documentation
test(solana): add integration tests for Raydium
```

## 🔍 Pull Request Process

### Before Submitting

- [ ] Code follows our style guidelines
- [ ] All tests pass (`make test`)
- [ ] Code coverage is maintained or improved
- [ ] Documentation is updated
- [ ] Self-review completed
- [ ] No merge conflicts

### PR Description Template

```markdown
## Description
Brief description of changes

## Type of Change
- [ ] Bug fix
- [ ] New feature
- [ ] Breaking change
- [ ] Documentation update

## Testing
- [ ] Unit tests added/updated
- [ ] Integration tests added/updated
- [ ] Manual testing completed

## Checklist
- [ ] Code follows style guidelines
- [ ] Self-review completed
- [ ] Documentation updated
- [ ] Tests pass
```

### Review Process

1. **Automated Checks** - CI/CD pipeline runs tests and checks
2. **Code Review** - Team members review your code
3. **Testing** - Manual testing if needed
4. **Approval** - At least one maintainer approval required
5. **Merge** - Squash and merge to main branch

## 🐛 Bug Reports

### Before Reporting

- Check existing issues to avoid duplicates
- Try to reproduce the bug
- Gather relevant information

### Bug Report Template

```markdown
**Bug Description**
Clear description of the bug

**Steps to Reproduce**
1. Go to '...'
2. Click on '....'
3. Scroll down to '....'
4. See error

**Expected Behavior**
What you expected to happen

**Actual Behavior**
What actually happened

**Environment**
- OS: [e.g. macOS, Linux, Windows]
- Go version: [e.g. 1.21.0]
- Docker version: [e.g. 20.10.0]

**Additional Context**
Screenshots, logs, or other relevant information
```

## ✨ Feature Requests

### Feature Request Template

```markdown
**Feature Description**
Clear description of the proposed feature

**Problem Statement**
What problem does this solve?

**Proposed Solution**
How should this feature work?

**Alternatives Considered**
Other solutions you've considered

**Additional Context**
Mockups, examples, or other relevant information
```

## 🔒 Security

### Reporting Security Issues

**DO NOT** create public issues for security vulnerabilities.

Instead, email us at: security@gocoffee.io

Include:
- Description of the vulnerability
- Steps to reproduce
- Potential impact
- Suggested fix (if any)

### Security Guidelines

- Never commit secrets or private keys
- Use environment variables for sensitive data
- Follow secure coding practices
- Validate all inputs
- Use HTTPS for all external communications

## 📚 Documentation

### Documentation Standards

- Write clear, concise documentation
- Include code examples
- Update relevant documentation with code changes
- Use proper markdown formatting
- Add diagrams for complex concepts

### Types of Documentation

- **API Documentation** - Endpoint descriptions and examples
- **User Guides** - Step-by-step instructions
- **Developer Guides** - Technical implementation details
- **Architecture Docs** - System design and architecture

## 🧪 Testing Guidelines

### Test Categories

1. **Unit Tests** - Test individual functions and methods
2. **Integration Tests** - Test component interactions
3. **End-to-End Tests** - Test complete user workflows
4. **Load Tests** - Test system performance under load

### Testing Best Practices

- Write tests before or alongside code (TDD)
- Test both happy path and error cases
- Use descriptive test names
- Keep tests simple and focused
- Mock external dependencies
- Use test fixtures for consistent data

### Running Tests

```bash
# All tests
make test

# Unit tests only
make unit-test

# Integration tests
make integration-test

# Solana tests
make solana-test

# With coverage
make coverage
```

## 🏷️ Issue Labels

We use labels to categorize issues:

- `bug` - Something isn't working
- `enhancement` - New feature or request
- `documentation` - Documentation improvements
- `good first issue` - Good for newcomers
- `help wanted` - Extra attention needed
- `priority:high` - High priority
- `priority:medium` - Medium priority
- `priority:low` - Low priority
- `area:wallet` - Wallet-related
- `area:defi` - DeFi-related
- `area:solana` - Solana-specific
- `area:coffee` - Coffee payment system

## 🎉 Recognition

Contributors will be recognized in:

- README.md contributors section
- Release notes for significant contributions
- Annual contributor appreciation posts
- Special Discord roles for active contributors

## 📞 Getting Help

- **Discord** - [Join our Discord](https://discord.gg/gocoffee)
- **GitHub Discussions** - [Ask questions](https://github.com/DimaJoyti/go-coffee/discussions)
- **Email** - development@gocoffee.io

## 📄 License

By contributing to Go Coffee, you agree that your contributions will be licensed under the MIT License.

---

Thank you for contributing to Go Coffee! Together, we're revolutionizing coffee payments with Web3 technology. ☕🚀
