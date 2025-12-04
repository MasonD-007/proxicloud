# ProxiCloud Release Guide

This guide covers how to create releases and manage your GitHub project like a pro.

## Table of Contents
- [Creating a Release](#creating-a-release)
- [GitHub Project Features](#github-project-features)
- [Automated Workflows](#automated-workflows)
- [Best Practices](#best-practices)

---

## Creating a Release

Your project is already set up with automated releases! Here's how to use it:

### 1. Prepare for Release

**Update the CHANGELOG.md:**
```bash
# Edit CHANGELOG.md and move items from [Unreleased] to a new version section
# Follow the format:
## [X.Y.Z] - YYYY-MM-DD

### Added
- New feature 1
- New feature 2

### Changed
- Updated feature X

### Fixed
- Bug fix Y
```

**Commit your changes:**
```bash
git add CHANGELOG.md
git commit -m "docs: Update changelog for vX.Y.Z"
git push origin main
```

### 2. Create and Push a Tag

**For a new release (e.g., v1.0.0):**
```bash
# Create an annotated tag
git tag -a v1.0.0 -m "Release v1.0.0 - Initial stable release"

# Push the tag to GitHub
git push origin v1.0.0
```

**That's it!** The GitHub Actions workflow will automatically:
- Build backend binaries for linux-amd64 and linux-arm64
- Build and package the frontend
- Generate SHA256 checksums
- Create a GitHub release with all artifacts
- Generate release notes
- Update the `latest` tag for the installer

### 3. Version Numbering (Semantic Versioning)

Follow [Semantic Versioning](https://semver.org/):

- **MAJOR** version (v2.0.0): Breaking changes
  - API changes that aren't backward compatible
  - Configuration format changes
  - Database schema changes

- **MINOR** version (v1.1.0): New features (backward compatible)
  - New API endpoints
  - New UI features
  - New configuration options

- **PATCH** version (v1.0.1): Bug fixes (backward compatible)
  - Bug fixes
  - Security patches
  - Performance improvements

**Examples:**
```bash
# First stable release
git tag -a v1.0.0 -m "Release v1.0.0 - Initial stable release"

# Bug fix release
git tag -a v1.0.1 -m "Release v1.0.1 - Fix container deletion bug"

# Feature release
git tag -a v1.1.0 -m "Release v1.1.0 - Add VM support"

# Breaking change
git tag -a v2.0.0 -m "Release v2.0.0 - New API authentication system"
```

### 4. Pre-releases and Beta Versions

For testing before stable release:

```bash
# Beta release
git tag -a v1.0.0-beta.1 -m "Release v1.0.0-beta.1 - Beta testing"
git push origin v1.0.0-beta.1

# Release candidate
git tag -a v1.0.0-rc.1 -m "Release v1.0.0-rc.1 - Release candidate"
git push origin v1.0.0-rc.1
```

The workflow will automatically mark these as "pre-release" on GitHub.

---

## GitHub Project Features

### 1. Repository Settings

**Enable useful features:**

```bash
# Go to: Settings → General

✅ Issues (for bug reports and feature requests)
✅ Projects (for project management)
✅ Discussions (for community Q&A)
✅ Wiki (for additional documentation)
```

**Branch Protection (Settings → Branches):**
- Protect `main` branch
- Require pull request reviews
- Require status checks to pass before merging
- Require branches to be up to date

### 2. Issue Templates

Create issue templates for better bug reports and feature requests.

**Create `.github/ISSUE_TEMPLATE/bug_report.md`:**
```markdown
---
name: Bug Report
about: Report a bug to help us improve
title: '[BUG] '
labels: bug
assignees: ''
---

**Describe the bug**
A clear description of what the bug is.

**To Reproduce**
Steps to reproduce the behavior:
1. Go to '...'
2. Click on '...'
3. See error

**Expected behavior**
What you expected to happen.

**Environment:**
- ProxiCloud version: [e.g., v1.0.0]
- Proxmox version: [e.g., 8.1]
- OS: [e.g., Debian 12]
- Browser: [e.g., Chrome 120]

**Additional context**
Add any other context about the problem here.
```

### 3. Pull Request Template

Create `.github/PULL_REQUEST_TEMPLATE.md`:
```markdown
## Description
Brief description of changes

## Type of Change
- [ ] Bug fix (non-breaking change)
- [ ] New feature (non-breaking change)
- [ ] Breaking change (fix or feature that would cause existing functionality to change)
- [ ] Documentation update

## Testing
- [ ] Tests pass locally
- [ ] Added new tests for new features
- [ ] Manual testing performed

## Checklist
- [ ] Code follows project style guidelines
- [ ] Self-review completed
- [ ] Comments added for complex code
- [ ] Documentation updated
- [ ] CHANGELOG.md updated
```

### 4. GitHub Projects (Project Management)

**Create a project board:**
1. Go to Projects → New Project
2. Choose "Board" template
3. Create columns: Backlog, To Do, In Progress, Done
4. Link issues to the project

**Automate with labels:**
```bash
# Go to Issues → Labels and create:
- enhancement (new feature)
- bug (something isn't working)
- documentation (docs improvements)
- good first issue (easy for newcomers)
- help wanted (extra attention needed)
- wontfix (will not be worked on)
- duplicate (already exists)
- priority:high
- priority:medium
- priority:low
```

### 5. GitHub Discussions

Enable Discussions for:
- Q&A section
- Feature requests discussion
- Show and tell (user deployments)
- Announcements

### 6. Repository Topics

Add topics to make your repo discoverable:
```
proxmox, lxc, containers, dashboard, golang, nextjs, 
typescript, devops, infrastructure, self-hosted
```

Go to: About (top right) → Settings icon → Topics

---

## Automated Workflows

Your project already has these workflows:

### Build Workflow (`.github/workflows/build.yml`)
**Triggers on:**
- Push to `main` or `develop` branches
- Pull requests to `main` or `develop`

**What it does:**
- Runs Go tests for backend
- Runs linter for frontend
- Builds backend binaries
- Builds frontend
- Uploads artifacts

### Release Workflow (`.github/workflows/release.yml`)
**Triggers on:**
- Push of tags matching `v*`

**What it does:**
- Builds release binaries
- Creates GitHub release
- Uploads artifacts with checksums
- Updates `latest` tag

### Additional Workflow Ideas

**Create `.github/workflows/auto-assign.yml` for auto-assigning issues:**
```yaml
name: Auto Assign
on:
  issues:
    types: [opened]
  pull_request:
    types: [opened]

jobs:
  assign:
    runs-on: ubuntu-latest
    steps:
      - uses: kentaro-m/auto-assign-action@v1.2.5
        with:
          configuration-path: '.github/auto-assign.yml'
```

**Create `.github/workflows/stale.yml` for closing stale issues:**
```yaml
name: Close Stale Issues
on:
  schedule:
    - cron: '0 0 * * *'

jobs:
  stale:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/stale@v9
        with:
          stale-issue-message: 'This issue is stale and will be closed in 7 days if there is no activity.'
          days-before-stale: 60
          days-before-close: 7
```

---

## Best Practices

### Release Checklist

Before creating a release:

- [ ] All tests passing in CI
- [ ] CHANGELOG.md updated
- [ ] Documentation updated (if needed)
- [ ] Version number follows semantic versioning
- [ ] Breaking changes clearly documented
- [ ] Migration guide provided (if breaking changes)
- [ ] Tested in a staging environment

### Communication

**Write good release notes:**
```markdown
## What's New in v1.1.0

### 🎉 Features
- **VM Support**: You can now manage KVM virtual machines
- **Bulk Operations**: Select multiple containers for batch actions
- **Dark/Light Theme Toggle**: Choose your preferred theme

### 🐛 Bug Fixes
- Fixed container deletion leaving orphaned data
- Resolved memory leak in metrics collector
- Fixed timezone display issues

### 📚 Documentation
- Added VM management guide
- Updated API documentation

### ⚠️ Breaking Changes
None

### 🔄 Upgrade Notes
Simply replace the binaries and restart services. No configuration changes needed.
```

### Maintenance

**Keep dependencies updated:**
```bash
# Backend (Go)
cd backend
go get -u ./...
go mod tidy

# Frontend (Node)
cd frontend
npm update
npm audit fix
```

**Regular release schedule:**
- Patch releases: As needed for critical bugs
- Minor releases: Monthly or when features are ready
- Major releases: When breaking changes accumulate

---

## Quick Reference

### Create Your First Release

```bash
# 1. Update changelog
vim CHANGELOG.md

# 2. Commit changes
git add CHANGELOG.md
git commit -m "docs: Update changelog for v1.0.0"
git push origin main

# 3. Create and push tag
git tag -a v1.0.0 -m "Release v1.0.0 - Initial stable release"
git push origin v1.0.0

# 4. Wait for GitHub Actions to complete
# 5. Check the Releases page on GitHub
```

### View Your Releases

```bash
# List all tags
git tag -l

# View release on GitHub
# Go to: https://github.com/YOUR_USERNAME/proxicloud/releases

# Download release artifacts
curl -L https://github.com/YOUR_USERNAME/proxicloud/releases/download/v1.0.0/proxicloud-api-linux-amd64
```

### Delete a Tag (if you made a mistake)

```bash
# Delete local tag
git tag -d v1.0.0

# Delete remote tag
git push origin :refs/tags/v1.0.0

# Delete release on GitHub manually
# Go to Releases → Click on release → Delete release
```

---

## Resources

- [GitHub Releases Documentation](https://docs.github.com/en/repositories/releasing-projects-on-github)
- [Semantic Versioning](https://semver.org/)
- [Keep a Changelog](https://keepachangelog.com/)
- [GitHub Actions Documentation](https://docs.github.com/en/actions)
- [GitHub Projects](https://docs.github.com/en/issues/planning-and-tracking-with-projects)

---

**Happy Releasing! 🚀**
