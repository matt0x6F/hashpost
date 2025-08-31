# Gomock Migration Guide

This document outlines the migration from `testify/mock` to `go.uber.org/mock/gomock` for better maintainability and type safety.

## Overview

We're migrating from manually maintained mocks using `testify/mock` to automatically generated mocks using `gomock`. This provides:

- **Zero manual maintenance** - Mocks are automatically generated from interfaces
- **Type safety** - Generated mocks always match interface signatures
- **Better error messages** - More detailed test failure information
- **Consistency** - All mocks follow the same pattern

## Current State

- **Old approach**: Manual mocks in `internal/database/dao/mocks/` and `internal/services/mocks/`
- **New approach**: Generated mocks using `go:generate` directives

## Migration Steps

### 1. Mock Generation

Mocks are automatically generated using:

```bash
make generate-mocks
```

This runs:
- `go generate ./internal/database/dao/` - Generates DAO mocks
- `go generate ./internal/services/` - Generates service mocks  
- `go generate ./internal/ibe/` - Generates IBE mocks

### 2. Test File Migration Pattern

#### Before (testify/mock):
```go
func TestHandler(t *testing.T) {
    mockUserDAO := &mocks.MockUserDAO{}
    
    mockUserDAO.On("GetUserByID", mock.Anything, int64(1)).
        Return(&models.User{UserID: 1}, nil)
    
    // Test logic...
}
```

#### After (gomock):
```go
func TestHandler(t *testing.T) {
    ctrl := gomock.NewController(t)
    defer ctrl.Finish()
    
    mockUserDAO := dao.NewMockUserDAOInterface(ctrl)
    
    mockUserDAO.EXPECT().
        GetUserByID(gomock.Any(), int64(1)).
        Return(&models.User{UserID: 1}, nil).
        Times(1)
    
    // Test logic...
}
```

### 3. Key Differences

#### Mock Creation
- **Old**: `&mocks.MockUserDAO{}`
- **New**: `dao.NewMockUserDAOInterface(ctrl)`

#### Expectation Setup
- **Old**: `mockUserDAO.On("method", args...).Return(result)`
- **New**: `mockUserDAO.EXPECT().method(args...).Return(result).Times(1)`

#### Controller Management
- **Old**: No controller needed
- **New**: Must create and finish controller: `ctrl := gomock.NewController(t); defer ctrl.Finish()`

#### Argument Matching
- **Old**: `mock.Anything`, `mock.AnythingOfType("string")`
- **New**: `gomock.Any()`, `gomock.AssignableToTypeOf("")`

### 4. Migration Checklist

For each test file:

1. **Add imports**: `gomock "go.uber.org/mock/gomock"`
2. **Create controller**: `ctrl := gomock.NewController(t); defer ctrl.Finish()`
3. **Replace mock creation**: Use generated mock constructors
4. **Update expectations**: Change from `.On()` to `.EXPECT()`
5. **Add Times()**: Specify how many times each call is expected
6. **Update argument matchers**: Use gomock equivalents
7. **Test**: Ensure tests still pass

### 5. Example Migration

#### Complete Example - Before:
```go
func TestAuthHandler_Login(t *testing.T) {
    handler, mockUserDAO, mockPseudonymDAO, _, mockRoleKeyDAO, _, _ := NewAuthHandlerWithMocks()
    
    testEmail := "test@example.com"
    testPassword := "TestPassword123!"
    testUserID := int64(1)
    
    hashedPassword := hashPasswordSHA256(testPassword)
    mockUser := &dbmodels.User{
        UserID:        testUserID,
        Email:         testEmail,
        PasswordHash:  hashedPassword,
        IsActive:      sql.Null[bool]{V: true, Valid: true},
        EmailVerified: sql.Null[bool]{V: true, Valid: true},
    }
    
    mockUserDAO.On("GetUserByEmail", mock.Anything, testEmail).Return(mockUser, nil)
    mockUserDAO.On("UpdateLastActive", mock.Anything, testUserID).Return(nil)
    
    // Test logic...
}
```

#### Complete Example - After:
```go
func TestAuthHandler_Login(t *testing.T) {
    ctrl := gomock.NewController(t)
    defer ctrl.Finish()
    
    handler, mockUserDAO, mockPseudonymDAO, _, mockRoleKeyDAO, _, _ := NewAuthHandlerWithGomocks(t)
    
    testEmail := "test@example.com"
    testPassword := "TestPassword123!"
    testUserID := int64(1)
    
    hashedPassword := hashPasswordSHA256(testPassword)
    mockUser := &models.User{
        UserID:        testUserID,
        Email:         testEmail,
        PasswordHash:  hashedPassword,
        IsActive:      sql.Null[bool]{V: true, Valid: true},
        EmailVerified: sql.Null[bool]{V: true, Valid: true},
    }
    
    mockUserDAO.EXPECT().
        GetUserByEmail(gomock.Any(), testEmail).
        Return(mockUser, nil).
        Times(1)
    
    mockUserDAO.EXPECT().
        UpdateLastActive(gomock.Any(), testUserID).
        Return(nil).
        Times(1)
    
    // Test logic...
}
```

### 6. Helper Function Updates

Update test helper functions to use gomock:

```go
// Before
func NewAuthHandlerWithMocks() (*handlers.AuthHandler, *mocks.MockUserDAO, ...) {
    mockUserDAO := &mocks.MockUserDAO{}
    // ...
}

// After  
func NewAuthHandlerWithGomocks(t *testing.T) (*handlers.AuthHandler, *dao.MockUserDAOInterface, ...) {
    ctrl := gomock.NewController(t)
    mockUserDAO := dao.NewMockUserDAOInterface(ctrl)
    // ...
}
```

### 7. Common Patterns

#### Multiple Calls
```go
// Old
mockDAO.On("GetUser", mock.Anything, 1).Return(user1, nil)
mockDAO.On("GetUser", mock.Anything, 2).Return(user2, nil)

// New
mockDAO.EXPECT().GetUser(gomock.Any(), 1).Return(user1, nil).Times(1)
mockDAO.EXPECT().GetUser(gomock.Any(), 2).Return(user2, nil).Times(1)
```

#### Any Times
```go
// Old
mockDAO.On("GetUser", mock.Anything, mock.AnythingOfType("int64")).Return(user, nil)

// New  
mockDAO.EXPECT().GetUser(gomock.Any(), gomock.Any()).Return(user, nil).AnyTimes()
```

#### Argument Matching
```go
// Old
mockDAO.On("GetUser", mock.Anything, mock.MatchedBy(func(id int64) bool { return id > 0 }))

// New
mockDAO.EXPECT().GetUser(gomock.Any(), gomock.AssignableToTypeOf(int64(0))).Return(user, nil)
```

### 8. Benefits After Migration

1. **No more manual mock maintenance**
2. **Interface changes automatically update mocks**
3. **Better test failure messages**
4. **Type safety guarantees**
5. **Consistent mock behavior**

### 9. Troubleshooting

#### Common Issues

1. **"no required module provides package go.uber.org/mock/gomock"**
   - Run: `go get go.uber.org/mock/gomock`

2. **Mock methods not found**
   - Regenerate mocks: `make generate-mocks`

3. **Import conflicts**
   - Remove old mock imports
   - Use generated mock imports from dao package

4. **Test failures after migration**
   - Check that all expectations are set
   - Verify argument matchers are correct
   - Ensure Times() is specified

### 10. Rollback Plan

If issues arise:

1. Keep old mock files as backup
2. Migrate one test file at a time
3. Run tests after each migration
4. Use git to revert specific files if needed

## Next Steps

1. **Phase 1**: Migrate simple test files first
2. **Phase 2**: Update test helper functions
3. **Phase 3**: Migrate complex test scenarios
4. **Phase 4**: Remove old mock files
5. **Phase 5**: Update CI/CD to use new mocks

## Questions?

Contact the development team or refer to the gomock documentation: https://github.com/uber-go/mock
