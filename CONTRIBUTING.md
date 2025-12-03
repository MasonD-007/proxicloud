# Contributing to ProxiCloud

Thank you for your interest in contributing to ProxiCloud! This document provides guidelines and instructions for contributing.

---

## 📋 Table of Contents

- [Code of Conduct](#code-of-conduct)
- [How Can I Contribute?](#how-can-i-contribute)
- [Development Setup](#development-setup)
- [Coding Guidelines](#coding-guidelines)
- [Commit Guidelines](#commit-guidelines)
- [Pull Request Process](#pull-request-process)
- [Reporting Bugs](#reporting-bugs)
- [Suggesting Features](#suggesting-features)

---

## 📜 Code of Conduct

This project follows a Code of Conduct that all contributors are expected to adhere to:

- Be respectful and inclusive
- Welcome newcomers and help them get started
- Focus on constructive feedback
- Assume good intentions
- Accept and provide constructive criticism gracefully

---

## 🤝 How Can I Contribute?

### Types of Contributions

1. **Bug Reports** - Report issues you encounter
2. **Feature Requests** - Suggest new features or improvements
3. **Code Contributions** - Submit bug fixes or new features
4. **Documentation** - Improve or add documentation
5. **Testing** - Test new features or releases
6. **Design** - Contribute to UI/UX design

---

## 🛠️ Development Setup

See [DEVELOPMENT.md](docs/DEVELOPMENT.md) for detailed development setup instructions.

### Quick Start

```bash
# Clone repository
git clone https://github.com/MasonD-007/proxicloud.git
cd proxicloud

# Backend setup
cd backend
go mod download
export PROXICLOUD_CONFIG="/tmp/proxicloud/config.yaml"
go run cmd/api/main.go

# Frontend setup (new terminal)
cd frontend
npm install
npm run dev
```

---

## 💻 Coding Guidelines

### Go (Backend)

**Style Guide**: Follow [Effective Go](https://go.dev/doc/effective_go)

**Best Practices**:
```go
// ✅ Good: Clear, descriptive names
func GetContainerMetrics(vmid int, timeRange string) (*Metrics, error) {
    // ...
}

// ❌ Bad: Unclear, abbreviated names
func GetCM(v int, t string) (*M, error) {
    // ...
}

// ✅ Good: Handle errors explicitly
metrics, err := client.GetMetrics(vmid)
if err != nil {
    return nil, fmt.Errorf("failed to get metrics: %w", err)
}

// ❌ Bad: Ignore errors
metrics, _ := client.GetMetrics(vmid)

// ✅ Good: Use context for cancellation
func (s *Server) Start(ctx context.Context) error {
    // ...
}

// ✅ Good: Document exported functions
// GetContainerMetrics retrieves historical metrics for a container.
// It returns an error if the container doesn't exist or if the
// Proxmox API is unreachable.
func GetContainerMetrics(vmid int) (*Metrics, error) {
    // ...
}
```

**Formatting**:
```bash
# Format code
gofmt -w .

# Run vet
go vet ./...

# Run linter (optional but recommended)
golangci-lint run
```

### TypeScript/React (Frontend)

**Style Guide**: Follow [TypeScript guidelines](https://www.typescriptlang.org/docs/)

**Best Practices**:
```typescript
// ✅ Good: Type-safe props
interface ContainerCardProps {
  container: Container;
  onStart: (vmid: number) => void;
  onStop: (vmid: number) => void;
}

// ✅ Good: Functional components with hooks
export function ContainerCard({ container, onStart, onStop }: ContainerCardProps) {
  const [isLoading, setIsLoading] = useState(false);
  
  // ...
}

// ✅ Good: Named exports for components
export function Dashboard() { /* ... */ }

// ✅ Good: Async/await for API calls
async function createContainer(config: ContainerConfig) {
  try {
    const response = await api.createContainer(config);
    return response;
  } catch (error) {
    console.error('Failed to create container:', error);
    throw error;
  }
}

// ✅ Good: Use TypeScript types
type Status = 'running' | 'stopped' | 'paused';

interface Container {
  vmid: number;
  name: string;
  status: Status;
  cpu: number;
  memory: number;
}
```

**Formatting**:
```bash
# Format code (if Prettier is configured)
npm run format

# Lint code
npm run lint

# Type check
npm run type-check
```

---

## 📝 Commit Guidelines

We use [Conventional Commits](https://www.conventionalcommits.org/) for clear and structured commit messages.

### Commit Message Format

```
<type>(<scope>): <description>

[optional body]

[optional footer]
```

### Types

- `feat` - New feature
- `fix` - Bug fix
- `docs` - Documentation changes
- `style` - Code style changes (formatting, no logic change)
- `refactor` - Code refactoring
- `perf` - Performance improvements
- `test` - Adding or updating tests
- `chore` - Maintenance tasks (deps, build, etc.)
- `ci` - CI/CD changes

### Examples

```bash
# Feature
git commit -m "feat(backend): add VM support endpoints"

# Bug fix
git commit -m "fix(frontend): resolve container list pagination issue"

# Documentation
git commit -m "docs(api): add WebSocket event documentation"

# Refactor
git commit -m "refactor(analytics): optimize metrics collection query"

# Multiple lines
git commit -m "feat(containers): add bulk operation support

- Add bulk start/stop functionality
- Update UI with multi-select checkboxes
- Add confirmation dialog for bulk operations"
```

---

## 🔄 Pull Request Process

### 1. Create a Branch

```bash
git checkout -b feature/your-feature-name
```

**Branch Naming**:
- `feature/add-vm-support` - New features
- `fix/container-list-crash` - Bug fixes
- `docs/update-api-guide` - Documentation
- `refactor/improve-metrics` - Refactoring

### 2. Make Your Changes

- Write clear, focused commits
- Add tests for new features
- Update documentation
- Follow coding guidelines

### 3. Test Your Changes

```bash
# Backend tests
cd backend
go test ./...

# Frontend lint
cd frontend
npm run lint

# Manual testing
# - Test affected functionality
# - Check for regressions
```

### 4. Push and Create PR

```bash
git push origin feature/your-feature-name
```

Then create a Pull Request on GitHub.

### 5. PR Checklist

- [ ] Branch is up to date with `main`
- [ ] All tests pass
- [ ] Code follows style guidelines
- [ ] Documentation is updated
- [ ] Commit messages are clear
- [ ] No merge conflicts
- [ ] Screenshots included (for UI changes)

### 6. Review Process

- Maintainers will review your PR
- Address feedback and requested changes
- Be responsive to comments
- Once approved, your PR will be merged

---

## 🐛 Reporting Bugs

### Before Submitting

1. Check [existing issues](https://github.com/MasonD-007/proxicloud/issues)
2. Ensure you're using the latest version
3. Try to reproduce the bug

### Bug Report Template

```markdown
**Describe the bug**
A clear and concise description of what the bug is.

**To Reproduce**
Steps to reproduce the behavior:
1. Go to '...'
2. Click on '...'
3. Scroll down to '...'
4. See error

**Expected behavior**
What you expected to happen.

**Screenshots**
If applicable, add screenshots.

**Environment:**
- ProxiCloud version: [e.g., v1.0.0]
- Proxmox VE version: [e.g., 8.1.3]
- OS: [e.g., Debian 12]
- Browser: [e.g., Chrome 120]

**Additional context**
Any other context about the problem.

**Logs**
```
Paste relevant logs here
```
```

---

## 💡 Suggesting Features

### Before Submitting

1. Check [existing feature requests](https://github.com/MasonD-007/proxicloud/issues?q=is%3Aissue+is%3Aopen+label%3Aenhancement)
2. Ensure it aligns with project goals
3. Consider if it benefits most users

### Feature Request Template

```markdown
**Is your feature request related to a problem?**
A clear description of the problem.

**Describe the solution you'd like**
A clear description of what you want to happen.

**Describe alternatives you've considered**
Alternative solutions or features you've considered.

**Additional context**
Any other context, screenshots, or mockups.

**Would you be willing to contribute this feature?**
- [ ] Yes, I can implement this
- [ ] Yes, with guidance
- [ ] No, but I'd like to see it
```

---

## 🎯 Good First Issues

New contributors should look for issues labeled:
- `good first issue` - Beginner-friendly
- `help wanted` - Maintainers need help
- `documentation` - Docs improvements

---

## 📞 Getting Help

- **Documentation**: Check [docs/](docs/)
- **Discussions**: [GitHub Discussions](https://github.com/MasonD-007/proxicloud/discussions)
- **Issues**: [GitHub Issues](https://github.com/MasonD-007/proxicloud/issues)

---

## 🏆 Recognition

Contributors will be:
- Listed in release notes
- Mentioned in the README (major contributions)
- Credited in commit history

---

## 📄 License

By contributing, you agree that your contributions will be licensed under the MIT License.

---

**Thank you for contributing to ProxiCloud!** 🚀
