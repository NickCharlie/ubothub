# Contributing to UBotHub

Thank you for your interest in contributing to UBotHub! This document provides guidelines and instructions for contributing.

## Code of Conduct

By participating in this project, you agree to maintain a respectful and inclusive environment for everyone.

## How to Contribute

### Reporting Bugs

1. Check existing [Issues](https://github.com/NickCharlie/ubothub/issues) to avoid duplicates.
2. Create a new issue with a clear title and description.
3. Include steps to reproduce, expected behavior, and actual behavior.
4. Attach relevant logs, screenshots, or error messages.

### Suggesting Features

1. Open an issue with the `feature` label.
2. Describe the use case and expected behavior.
3. Explain why this feature would be useful to the project.

### Submitting Pull Requests

1. Fork the repository and create a feature branch from `main`.
2. Follow the coding standards described below.
3. Write tests for new functionality.
4. Ensure all existing tests pass.
5. Submit a pull request with a clear description of changes.

## Development Setup

```bash
# Clone your fork
git clone git@github.com:YOUR_USERNAME/ubothub.git
cd ubothub

# Start infrastructure
make docker-up

# Backend development
cd backend && air

# Frontend development
cd frontend && pnpm dev
```

## Coding Standards

### Go Backend

- Follow [Effective Go](https://go.dev/doc/effective_go) and [Google Go Style Guide](https://google.github.io/styleguide/go/).
- Use `gofmt` for formatting (enforced by CI).
- Write table-driven tests where applicable.
- Use structured logging via `zap` with module-scoped loggers.
- All public functions and types must have documentation comments.

### React Frontend

- Follow [Google TypeScript Style Guide](https://google.github.io/styleguide/tsguide.html).
- Use functional components with hooks.
- Use TypeScript strict mode.
- Format code with Prettier and lint with ESLint.

### Commit Messages

Follow [Conventional Commits](https://www.conventionalcommits.org/):

```
feat(scope): add new feature
fix(scope): resolve specific bug
docs: update documentation
refactor(scope): restructure without behavior change
test(scope): add or update tests
chore: maintenance tasks
```

## Branch Strategy

- `main` — stable release branch
- `develop` — integration branch
- `feat/*` — feature branches
- `fix/*` — bug fix branches

## License

By contributing, you agree that your contributions will be licensed under the GPL-3.0 License.
