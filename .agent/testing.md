# Testing Requirements

## Backend Testing

**Tests are MANDATORY for all new features.**

| Layer | Coverage Target |
|-------|-----------------|
| Service layer | 100% |
| Handler layer | >90% |

### Conventions

- Test file naming: `*_test.go`
- Test function naming: `TestFunctionName_Scenario`
- Use mock repositories/services for isolation

### Example

```go
func TestCreateUser_Success(t *testing.T) {
    // Arrange
    mockRepo := mocks.NewMockUserRepository()
    svc := service.NewUserService(mockRepo)
    
    // Act
    user, err := svc.CreateUser(ctx, "test@example.com")
    
    // Assert
    assert.NoError(t, err)
    assert.Equal(t, "test@example.com", user.Email)
}

func TestCreateUser_EmptyEmail(t *testing.T) {
    // ...
}
```

## Frontend Testing

| Type | Tool | Purpose |
|------|------|---------|
| Component tests | React Testing Library | Test UI behavior |
| Unit tests | Vitest | Test utility functions |
| Integration tests | Vitest | Test user flows |

### Commands

```bash
just fe-test          # Run all tests
```
