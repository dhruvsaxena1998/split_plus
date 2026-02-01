# Backend Conventions (Go)

## Project Structure

```
be/internal/
├── http/handlers/    # HTTP layer only
├── service/          # Business logic
├── repository/       # Data access
├── db/               # SQL queries via sqlc
└── app/              # Dependency wiring
```

## Naming Conventions

| Type | Pattern | Example |
|------|---------|---------|
| Interface | PascalCase | `UserService`, `UserRepository` |
| Implementation | camelCase | `userService`, `userRepository` |
| Constructor | `New` + Interface | `NewUserService()` |
| Error variable | `Err` + Description | `ErrUserNotFound` |
| Test file | `*_test.go` | `user_test.go` |
| Test function | `Test` + Func + `_` + Scenario | `TestCreateUser_InvalidEmail` |

## Handler Pattern

```go
type CreateUserRequest struct {
    Email string `json:"email" validate:"required,email"`
}

type UserResponse struct {
    ID    int64  `json:"id"`
    Email string `json:"email"`
}

func CreateUserHandler(userService service.UserService) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        req, ok := middleware.GetBody[CreateUserRequest](r)
        if !ok {
            response.SendError(w, http.StatusInternalServerError, "invalid request context")
            return
        }
        
        user, err := userService.CreateUser(r.Context(), req.Email)
        if err != nil {
            response.SendError(w, statusCode, err.Error())
            return
        }
        
        resp := UserResponse{ID: user.ID, Email: user.Email}
        response.SendSuccess(w, http.StatusCreated, resp)
    }
}
```

## Service Pattern

```go
var (
    ErrUserNotFound = errors.New("user not found")
)

type UserService interface {
    CreateUser(ctx context.Context, email string) (sqlc.User, error)
}

type userService struct {
    repo repository.UserRepository
}

func NewUserService(repo repository.UserRepository) UserService {
    return &userService{repo: repo}
}

func (s *userService) CreateUser(ctx context.Context, email string) (sqlc.User, error) {
    email = strings.TrimSpace(email)
    if email == "" {
        return sqlc.User{}, errors.New("email is required")
    }
    return s.repo.CreateUser(ctx, email)
}
```

## Error Handling

```go
// Define errors at package level
var (
    ErrUserNotFound = errors.New("user not found")
    ErrUserAlreadyExists = errors.New("user already exists")
)

// Map service errors to HTTP status codes
switch err {
case service.ErrUserNotFound:
    statusCode = http.StatusNotFound
case service.ErrUserAlreadyExists:
    statusCode = http.StatusConflict
default:
    statusCode = http.StatusBadRequest
}
```

## Imports Order

```go
import (
    // Standard library
    "context"
    "errors"
    "net/http"
    
    // Third-party
    "github.com/go-chi/chi/v5"
    
    // Internal packages
    "github.com/dhruvsaxena1998/split_plus/internal/service"
)
```

## HTTP Status Codes

| Code | Usage |
|------|-------|
| `200 OK` | Successful GET/PUT/PATCH |
| `201 Created` | Successful POST |
| `400 Bad Request` | Invalid request format |
| `404 Not Found` | Resource not found |
| `409 Conflict` | Resource already exists |
| `422 Unprocessable Entity` | Validation errors |
| `500 Internal Server Error` | Server errors |

## Anti-Patterns

- Business logic in handlers
- Business logic in repositories
- Direct database access from handlers
- Ignoring errors
- Exposing internal errors to clients
- Global state
- God objects
