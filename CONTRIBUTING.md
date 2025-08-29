# Contributing to Go Logstash Logger

Thank you for your interest in contributing to this project! We welcome contributions from the community.

## How to Contribute

### Reporting Bugs

Before creating bug reports, please check the existing issues to avoid duplicates. When creating a bug report, please include:

- **Clear description** of what you expected to happen vs. what actually happened
- **Steps to reproduce** the issue
- **Go version** you're using
- **Operating system** and version
- **Code sample** that demonstrates the issue (if applicable)

### Suggesting Enhancements

Enhancement suggestions are welcome! Please provide:

- **Clear description** of the enhancement
- **Use case** that would benefit from this enhancement
- **Possible implementation** approach (if you have ideas)

### Pull Requests

1. **Fork** the repository
2. **Create a feature branch** from `main`
3. **Make your changes**
4. **Add tests** for new functionality
5. **Ensure all tests pass**: `go test -v`
6. **Update documentation** if necessary
7. **Submit a pull request**

### Development Setup

```bash
# Clone your fork
git clone https://github.com/SergeyDavidenko/logger.git
cd logger

# Install dependencies
go mod download

# Run tests
go test -v

# Run example
cd example
go run main.go
```

### Code Style

- Follow standard Go conventions
- Use `gofmt` to format your code
- Add comments for exported functions and types
- Keep functions focused and small
- Write clear, descriptive variable names

### Testing

- Add unit tests for new functionality
- Ensure all existing tests continue to pass
- Test both success and error cases
- Include performance tests for critical paths

### Commit Messages

Use clear and descriptive commit messages:

```
Add support for custom log formatters

- Implement LogFormatter interface
- Add JSON and text formatters
- Update tests and documentation
```

## Code of Conduct

This project follows the [Go Community Code of Conduct](https://golang.org/conduct). Please be respectful and constructive in all interactions.

## Questions?

Feel free to open an issue for questions or discussions about the project.
