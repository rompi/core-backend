# Auth Package
The `auth` package provides authentication functionalities for your application. It includes methods for user login, logout, and session management.

## Installation
install as a golang package using:
```
go get github.com/rompi/core-backend/auth
```

## In Scope
The `auth` package includes the following features:
- User authentication
- Register
  - using email and password
- Login
- Logout
- Session management (Using JWT tokens)
- Password reset
- Token generation and validation
- Middleware for protecting routes
- Role-based access control
- Rate limiting for authentication endpoints
- Logging of authentication events
- Support for multiple authentication methods (e.g., email/password, API keys)
- password hashing and salting
- Account lockout after multiple failed login attempts
- Support for password complexity requirements
- Integration with any persistence for storing user credentials and session data
- Comprehensive unit tests for all functionalities
- Detailed documentation for developers
- well-defined error handling and reporting
- adherence to security best practices
- update and change password functionality
- Detailed logging and monitoring of authentication events
- Multi-language support for error messages and responses
- Comprehensive API documentation
- Support for environment-based configuration (e.g., development, staging, production)

## out of Scope
The `auth` package does not include:
- User profile management
- Social media authentication (e.g., OAuth)
- Email verification
- Two-factor authentication (2FA)
- Integration with third-party identity providers
- Frontend components for authentication
- User registration workflows beyond basic email/password
- login via biometric methods

## Data Models
The `auth` package uses the following data models:
- User
  - ID (uuid)
  - Email
  - PasswordHash
  - Roles
  - Metadata (JSONB)
  - Status (string)
  - CreatedAt
  - UpdatedAt
- Session
  - ID (uuid)
  - UserID (uuid)
  - Token (string)
  - ExpiresAt
  - CreatedAt
  - UpdatedAt
- PasswordResetToken
  - ID (uuid)
  - UserID (uuid)
  - Token (string)
  - ExpiresAt
  - CreatedAt
  - UpdatedAt
- Role
  - ID (uuid)
  - Name (string)
  - Permissions (JSONB)
  - CreatedAt
  - UpdatedAt
- AuditLog
  - ID (uuid)
  - UserID (uuid)
  - Action (string)
  - Timestamp
  - Metadata (JSONB)

## How to Use
Import the `auth` package in your Go application:
```go
import "github.com/rompi/core-backend/auth"
```
Initialize the authentication service:
```go
authService := auth.NewService(yourConfig)
```
Use the provided methods for authentication:
```go
// Register a new user
user, err := authService.Register(email, password)  
if err != nil {
    // handle error
}
```
```go
// Login a user
token, err := authService.Login(email, password)
if err != nil {
    // handle error
}
```
```go
// Logout a user
err := authService.Logout(token)
if err != nil {
    // handle error
}
```
```go
// Validate a token
isValid, err := authService.ValidateToken(token)
if err != nil {
    // handle error
}
```
Refer to the package documentation for more details on available methods and their usage.