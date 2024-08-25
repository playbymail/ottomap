# Ubiquitous Language Documentation

---

## Authentication Domain

This domain defines the way we identify and authorize users.

## Introduction

As part of our design-driven development approach, it is crucial to establish a shared understanding of the key terms we use in our codebase.
This section will define the concepts of a **Repository**, a **Store**, and a **Service** within the context of our authentication domain.
Understanding these concepts will help ensure consistency across the project and facilitate smoother collaboration among team members.

### Repository

#### Definition
A **Repository** is an abstraction layer that provides access to domain entities.
It acts as a mediator between the domain layer and the data layer, encapsulating the logic required to retrieve and store data.
The repository pattern allows for decoupling the domain model from the database or any other persistence mechanism.

#### Purpose
The main purpose of a repository is to hide the details of data access and persistence.
This ensures that changes in the data storage mechanism do not affect the rest of the application.
For example, if we decide to switch from SQLite to PostgreSQL, only the repository layer would need to change, leaving the domain and service layers unaffected.

#### Example in Authentication
In our authentication package, a `UserRepository` would be responsible for providing access to user entities.
It would have methods like `FindByID`, `FindByUsername`, `Save`, and `Delete`.

```go
type UserRepository interface {
    FindByID(id int64) (*User, error)
    FindByUsername(username string) (*User, error)
    Save(user *User) error
    Delete(id int64) error
}
```

### Store

#### Definition
A **Store** is a more concrete implementation of a repository.
While the repository defines the interface, the store provides the actual implementation that interacts with the database or other persistence layers.
The store is where you would typically find SQL queries or calls to other databases, such as Redis or MongoDB.

#### Purpose
The store serves as the bridge between the high-level business logic and the low-level data access code.
It is responsible for executing the specific data access operations defined by the repository interface.

#### Example in Authentication
In our package, a `UserStore` might be the SQLite-backed implementation of the `UserRepository`.
It would contain the actual SQL queries used to retrieve or store user data.

```go
type SQLiteUserStore struct {
    db *sql.DB
}

func (s *SQLiteUserStore) FindByID(id int64) (*User, error) {
    var user User
    err := s.db.QueryRow("SELECT id, username, email, password_hash FROM users WHERE id = ?", id).Scan(&user.ID, &user.Username, &user.Email, &user.PasswordHash)
    if err != nil {
        return nil, err
    }
    return &user, nil
}
```

### Service

#### Definition
A **Service** encapsulates the business logic of an application.
It coordinates the interactions between repositories and other parts of the application.
A service typically includes operations that require multiple steps or involve multiple repositories.

#### Purpose
The service layer is responsible for handling the core use cases of the application.
It ensures that the business rules are applied consistently and acts as the orchestrator of the different components involved in a specific operation.

#### Example in Authentication
In our authentication package, an `AuthService` would be responsible for handling the core authentication logic, such as user registration, login, and password management.
The service would interact with the `UserRepository` (via the store) to persist data and manage the overall process.

```go
type AuthService struct {
    userRepo UserRepository
}

func (s *AuthService) RegisterUser(username, email, password string) (*User, error) {
    hashedPassword, err := hashPassword(password)
    if err != nil {
        return nil, err
    }

    user := &User{
        Username:     username,
        Email:        email,
        PasswordHash: hashedPassword,
    }

    err = s.userRepo.Save(user)
    if err != nil {
        return nil, err
    }

    return user, nil
}

func (s *AuthService) Login(username, password string) (*User, error) {
    user, err := s.userRepo.FindByUsername(username)
    if err != nil {
        return nil, err
    }

    if !checkPasswordHash(password, user.PasswordHash) {
        return nil, errors.New("invalid credentials")
    }

    return user, nil
}
```

## Conclusion

In summary:
- **Repository**: Defines the interface for accessing domain entities, abstracting the data layer.
- **Store**: Implements the repository interface, containing the actual data access logic.
- **Service**: Handles business logic, orchestrating the interactions between repositories and other parts of the application.

Understanding these distinctions will help you navigate our codebase more effectively and contribute to the project with greater confidence.
If you have any further questions, don’t hesitate to reach out.
