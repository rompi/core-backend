---
description: Analyze items in a GitHub task based on the task description
---

# Analyze GitHub Task

You are given a GitHub issue/task `$ARGUMENTS`. Your job is to:

1. **Fetch the task details** - Use the GitHub MCP tool to retrieve the full issue description, labels, and metadata
2. **Extract actionable items** - Parse the description to identify specific work items, requirements, and acceptance criteria
3. **Analyze API contracts** - Define request/response schemas, endpoints, HTTP methods, and status codes for any API changes
4. **Assess security implications** - Identify authentication, authorization, data validation, and security risks
5. **Review data integration** - Analyze data flow, external integrations, database schema changes, and data migrations
6. **Analyze complexity** - Assess the technical complexity of each item (simple, moderate, complex)
7. **Identify dependencies** - Note any dependencies between items, external systems, APIs, or third-party services
8. **Suggest breakdown** - If the task is large, suggest how to break it into smaller sub-tasks
9. **Provide implementation hints** - Reference relevant files, functions, or patterns from the codebase that should be used

## Output Format

**IMPORTANT**: Create a dedicated directory for the issue analysis with the following structure:
```
docs/plans/issue-{NUMBER}/
├── analysis.md          # Main analysis document (use the template below)
├── api-contract.md      # API contract specifications (if applicable)
├── schema.md            # DB Schema changes (if applicable)
├── config.md            # Configuration changes (if applicable)
├── context.md           # Additional context, requirements, stakeholder notes
└── implementation-plan.md  # Step-by-step implementation code plan (optional)
```

Example:
```
docs/plans/issue-123/
├── analysis.md
├── api-contract.md
├── context.md
└── implementation-plan.md
```

**Primary file**: The main analysis should be saved as `docs/plans/issue-{NUMBER}/analysis.md` with the title `# Issue #{NUMBER}: {Issue Title}`.

The `analysis.md` file should contain:

- **Summary**: One-sentence overview of the task
- **Actionable Items**: Numbered list of specific work items

- **API Contract Analysis** (if applicable):
  - **Endpoints**: List of endpoints with HTTP methods (GET, POST, PUT, DELETE, PATCH)
  - **Request Schema**: Detailed request body/parameters with types and validation rules
  - **Response Schema**: Expected response structure with status codes (200, 201, 400, 401, 403, 404, 500)
  - **Headers**: Required headers (Authorization, Content-Type, etc.)
  - **Authentication**: Required authentication/authorization levels
  - **Rate Limiting**: Any rate limiting considerations
  - **Versioning**: API versioning strategy if applicable
  - **Swagger/OpenAPI**: Swagger annotations to be added

- **Security Analysis** (CRITICAL - Always include):
  - **Authentication Requirements**: How users/services will be authenticated (JWT, API keys, OAuth, etc.)
  - **Authorization Rules**: Who can access what (RBAC, permissions, ownership checks)
  - **Input Validation**: All input validation rules and sanitization requirements
  - **Data Protection**: Sensitive data handling (PII, passwords, tokens, secrets)
  - **SQL Injection Prevention**: Use of prepared statements and parameterized queries
  - **XSS Prevention**: Output encoding and sanitization
  - **CSRF Protection**: CSRF token requirements for state-changing operations
  - **Rate Limiting**: Throttling to prevent abuse
  - **Audit Logging**: Security events that need to be logged
  - **Encryption**: Data encryption requirements (at rest, in transit)
  - **Security Headers**: Required security headers (CORS, CSP, etc.)
  - **Threat Model**: Potential security threats and mitigations

- **Data Integration Analysis** (CRITICAL - Always include):
  - **Data Flow**: Detailed data flow diagram or description (source → processing → destination)
  - **External Systems**: List of external APIs, services, or databases to integrate with
  - **Data Mapping**: How data maps between systems (field mappings, transformations)
  - **Data Validation**: Validation rules for incoming and outgoing data
  - **Error Handling**: How to handle integration failures, retries, circuit breakers
  - **Data Consistency**: How to maintain data consistency across systems
  - **Idempotency**: Idempotency requirements for operations
  - **Transaction Management**: Transaction boundaries and rollback strategies
  - **Webhooks/Events**: Any webhook or event-driven integrations
  - **Data Sync**: Real-time vs batch synchronization requirements
  - **Fallback Strategy**: What happens when external systems are unavailable

- **Database Changes**:
  - **Migration Scripts**: Migration file names and execution order
  - **Schema Changes**: Tables, columns, indexes, constraints to add/modify/remove
  - **SQL Queries**: Exact SQL migration queries with rollback scripts
  - **Data Migration**: Any data transformation or migration logic needed
  - **Backward Compatibility**: How to maintain compatibility during deployment
  - **Performance Impact**: Impact on existing queries and indexes

- **Code Changes Needed**:
  - **New Files**: Files to be created with their purpose
  - **Modified Files**: Existing files that need changes
  - **Package Structure**: New packages or reorganization needed
  - **Reusable Components**: Existing functions/packages to reuse (DRY principle)

- **Test Cases** (Comprehensive coverage required):
  - **Unit Tests**: Function-level tests with table-driven approach
  - **Integration Tests**: Database and API integration tests
  - **Security Tests**: Authentication, authorization, input validation tests
  - **Positive Cases**: Expected successful scenarios to test
  - **Edge Cases**: Boundary conditions, empty inputs, null values, unusual inputs
  - **Error Cases**: Expected failure scenarios and error handling
  - **Performance Tests**: Load testing for high-traffic endpoints (if applicable)
  - **Test Data**: Required test fixtures and mock data

- **Complexity Assessment**: Overall complexity rating and per-item ratings (Simple/Moderate/Complex)
- **Dependencies**: List of dependencies (code, APIs, other tasks, external systems, libraries)
- **Suggested Breakdown**: If applicable, how to split into sub-tasks with clear boundaries
- **Implementation Notes**:
  - Specific files, patterns, or approaches to use
  - SOLID principles application
  - DRY methodology - existing code to reuse
  - Go idioms to follow
- **Security Considerations**: Detailed security requirements and compliance needs
- **Performance Considerations**:
  - Expected load and scalability requirements
  - Caching strategy
  - Database query optimization
  - Connection pooling
- **Monitoring & Observability**:
  - Logging requirements (structured logs, log levels)
  - Metrics to track (counters, gauges, histograms)
  - Alerts to configure
  - Tracing requirements
- **Estimated Effort**: T-shirt sizing (XS, S, M, L, XL) with justification

### File Organization Guidelines

**When to create additional files:**

1. **api-contract.md** - Create when:
   - Issue involves API endpoints (REST, GraphQL, gRPC)
   - Multiple endpoints with complex request/response schemas
   - API versioning or contract evolution is involved

2. **context.md** - Create when:
   - Issue has extensive background information
   - Multiple stakeholder requirements or meeting notes
   - Business context and domain knowledge needed
   - Links to related issues, PRs, or external documents

3. **implementation-plan.md** - Create when:
   - Issue is complex and requires step-by-step breakdown
   - Multiple phases or iterations planned
   - Specific order of operations is critical
   - Need to track implementation progress separately

**Minimum requirement**: Always create `analysis.md` as the main analysis document.

---

## Example File Structure

### docs/plans/issue-123/analysis.md
```markdown
# Issue #123: Add User Authentication

## Summary
Implement JWT-based authentication for user login and registration with secure password hashing and token management.

## Actionable Items
1. Create user model and repository following SOLID principles
2. Implement password hashing using bcrypt (existing crypto package)
3. Implement JWT token generation and validation
4. Create authentication middleware for protected routes
5. Add rate limiting to prevent brute force attacks
6. Implement audit logging for authentication events

## API Contract Analysis

### Endpoints
- `POST /api/v1/auth/register` - User registration
- `POST /api/v1/auth/login` - User login
- `POST /api/v1/auth/refresh` - Refresh access token
- `POST /api/v1/auth/logout` - User logout

### Request Schema (Register)
```json
{
  "email": "string (required, valid email format, max 255 chars)",
  "password": "string (required, min 8 chars, must contain uppercase, lowercase, number, special char)",
  "name": "string (required, max 100 chars)"
}
```

### Response Schema (Login Success)
```json
{
  "access_token": "string (JWT)",
  "refresh_token": "string (JWT)",
  "expires_in": "number (seconds)",
  "token_type": "Bearer"
}
```
Status Codes: 200 (success), 400 (validation error), 401 (invalid credentials), 429 (rate limit)

### Headers
- `Content-Type: application/json`
- `Authorization: Bearer <token>` (for protected endpoints)

### Swagger/OpenAPI
Add Swagger annotations to all auth endpoints following project conventions.

## Security Analysis

### Authentication Requirements
- JWT tokens with HS256 algorithm
- Access token expiry: 15 minutes
- Refresh token expiry: 7 days
- Store refresh tokens in database with revocation capability

### Authorization Rules
- Public endpoints: register, login
- Protected endpoints require valid JWT in Authorization header
- Role-based access control for admin endpoints

### Input Validation
- Email: valid format, max 255 chars, lowercase normalization
- Password: min 8 chars, complexity requirements (uppercase, lowercase, number, special char)
- Sanitize all string inputs to prevent XSS

### Data Protection
- Hash passwords using bcrypt (cost factor: 12)
- Never log passwords or tokens
- Encrypt refresh tokens in database
- Use secure random for token generation

### SQL Injection Prevention
- Use prepared statements for all queries
- Parameterize user inputs

### Rate Limiting
- Login: 5 attempts per 15 minutes per IP
- Register: 3 attempts per hour per IP
- Use Redis or in-memory cache for rate limit tracking

### Audit Logging
- Log all authentication events: login success/failure, registration, logout
- Include: timestamp, IP address, user agent, user ID (if authenticated)
- No sensitive data in logs

### Security Headers
- Add CORS headers with allowed origins
- Set secure cookie flags (HttpOnly, Secure, SameSite)

### Threat Model
- Brute force attacks → Rate limiting
- Token theft → Short expiry, secure storage
- SQL injection → Prepared statements
- XSS → Input sanitization, output encoding

## Data Integration Analysis

### Data Flow
1. User submits credentials → API endpoint
2. Validate input → Check format, length, complexity
3. Query database for user → Use prepared statement
4. Verify password → bcrypt comparison
5. Generate JWT tokens → Sign with secret key
6. Store refresh token → Database with expiry
7. Return tokens → JSON response

### External Systems
- None (authentication is internal)

### Data Mapping
- Request email → User.email (lowercase)
- Request password → User.password_hash (bcrypt hashed)
- Generated tokens → JWT payload with user claims

### Data Validation
- Pre-persistence: email format, password complexity
- Post-retrieval: token signature verification, expiry check

### Error Handling
- Database errors → 500 Internal Server Error
- Validation errors → 400 Bad Request with error details
- Invalid credentials → 401 Unauthorized (generic message)
- Use circuit breaker pattern if external services added later

### Idempotency
- Register: Check for existing email before insert
- Login: Stateless operation, inherently idempotent

### Transaction Management
- User registration: Single transaction (insert user)
- Token refresh: Transaction to invalidate old token and create new one

## Database Changes

### Migration Scripts
- `20250107_001_create_users_table.sql`
- `20250107_002_create_refresh_tokens_table.sql`

### Schema Changes
```sql
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    email VARCHAR(255) NOT NULL UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    name VARCHAR(100) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_users_email ON users(email);

CREATE TABLE refresh_tokens (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash VARCHAR(255) NOT NULL UNIQUE,
    expires_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_refresh_tokens_user_id ON refresh_tokens(user_id);
CREATE INDEX idx_refresh_tokens_expires_at ON refresh_tokens(expires_at);
```

### Rollback Script
```sql
DROP TABLE IF EXISTS refresh_tokens;
DROP TABLE IF EXISTS users;
```

## Code Changes Needed

### New Files
- `internal/auth/service.go` - Authentication service (reuse existing validation patterns)
- `internal/auth/service_test.go` - Comprehensive unit tests
- `internal/auth/handler.go` - HTTP handlers with Swagger annotations
- `internal/auth/handler_test.go` - Handler tests
- `internal/auth/middleware.go` - JWT validation middleware
- `internal/auth/middleware_test.go` - Middleware tests
- `internal/user/repository.go` - User data access (reuse existing DB patterns)
- `internal/user/repository_test.go` - Repository tests
- `internal/user/model.go` - User model definition
- `migrations/20250107_001_create_users_table.sql`
- `migrations/20250107_002_create_refresh_tokens_table.sql`

### Modified Files
- `cmd/assistantd/main.go` - Add auth routes
- `docs/swagger.yaml` - Add auth endpoint documentation

### Reusable Components
- Use existing database connection pool
- Reuse existing error handling middleware
- Reuse existing validation utilities (if available, otherwise create reusable validator)
- Follow existing handler pattern for consistent API responses

## Test Cases

### Unit Tests
- `TestHashPassword` - Password hashing and verification
- `TestGenerateJWT` - Token generation with claims
- `TestValidateJWT` - Token validation and expiry
- `TestUserRepository_Create` - User creation with duplicate email check
- `TestUserRepository_FindByEmail` - User retrieval

### Integration Tests
- Full registration flow with database
- Full login flow with token generation
- Token refresh flow
- Rate limiting enforcement

### Security Tests
- SQL injection attempts
- XSS payload in inputs
- Invalid token signatures
- Expired tokens
- Brute force simulation

### Positive Cases
- Valid registration → 201 Created
- Valid login → 200 OK with tokens
- Valid token refresh → 200 OK with new tokens

### Edge Cases
- Empty email/password
- Maximum length inputs
- Special characters in name
- Concurrent registrations with same email

### Error Cases
- Duplicate email registration → 400 Bad Request
- Invalid email format → 400 Bad Request
- Weak password → 400 Bad Request
- Invalid credentials → 401 Unauthorized
- Expired token → 401 Unauthorized
- Rate limit exceeded → 429 Too Many Requests

## Complexity Assessment
Overall: **Moderate**
- User model & repository: Simple
- Password hashing: Simple (use existing bcrypt library)
- JWT generation: Moderate
- Authentication middleware: Moderate
- Rate limiting: Moderate

## Dependencies
- bcrypt library (for password hashing)
- JWT library (e.g., golang-jwt/jwt)
- Existing database connection package
- Existing validation utilities (or create reusable validator)

## Implementation Notes
- Follow SOLID principles: separate authentication logic, user repository, and HTTP handling
- Apply DRY: reuse existing database patterns, validation utilities, and error handling
- Use Go idioms: accept interfaces for repositories, return concrete types
- Implement comprehensive table-driven tests for all components
- Add structured logging for all authentication events
- Consider adding metrics for login success/failure rates

## Security Considerations
- OWASP Top 10 compliance
- Secure password storage with bcrypt
- JWT token security best practices
- Rate limiting to prevent abuse
- Audit logging for compliance
- Input validation and sanitization
- Secure headers and CORS configuration

## Performance Considerations
- Database indexes on email and token lookup fields
- Connection pooling for database access
- In-memory cache for rate limiting (Redis recommended for production)
- JWT validation should be fast (stateless)

## Monitoring & Observability
- Log metrics: login attempts, success rate, failures
- Track: registration rate, active sessions, token refresh rate
- Alerts: high failure rate, unusual registration patterns
- Structured logging with correlation IDs

## Estimated Effort
**M (Medium)** - Approximately 3-5 days for a senior engineer
- Authentication logic: 1 day
- Database schema and migrations: 0.5 day
- API endpoints and middleware: 1 day
- Comprehensive testing: 1-1.5 days
- Documentation and security review: 1 day
```

### docs/plans/issue-123/api-contract.md
```markdown
# API Contract - Issue #123: User Authentication

## Endpoints Overview
| Method | Endpoint | Description | Auth Required |
|--------|----------|-------------|---------------|
| POST | /api/v1/auth/register | User registration | No |
| POST | /api/v1/auth/login | User login | No |
| POST | /api/v1/auth/refresh | Refresh access token | Yes (Refresh Token) |
| POST | /api/v1/auth/logout | User logout | Yes (Access Token) |

## POST /api/v1/auth/register

### Request
**Headers:**
- `Content-Type: application/json`

**Body:**
```json
{
  "email": "user@example.com",
  "password": "SecurePass123!",
  "name": "John Doe"
}
```

**Validation Rules:**
- `email`: Required, valid email format, max 255 chars, unique in database
- `password`: Required, min 8 chars, must contain uppercase, lowercase, number, special char
- `name`: Required, max 100 chars

### Response

**Success (201 Created):**
```json
{
  "user_id": "uuid",
  "email": "user@example.com",
  "name": "John Doe",
  "created_at": "2025-01-07T10:30:00Z"
}
```

**Error (400 Bad Request):**
```json
{
  "error": "validation_error",
  "message": "Invalid input",
  "details": [
    {
      "field": "password",
      "message": "Password must contain at least one uppercase letter"
    }
  ]
}
```

**Error (409 Conflict):**
```json
{
  "error": "duplicate_email",
  "message": "Email already registered"
}
```

### Swagger Annotation
```go
// @Summary      Register a new user
// @Description  Creates a new user account with email and password
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request  body      RegisterRequest  true  "User registration data"
// @Success      201      {object}  RegisterResponse
// @Failure      400      {object}  ErrorResponse
// @Failure      409      {object}  ErrorResponse
// @Failure      500      {object}  ErrorResponse
// @Router       /api/v1/auth/register [post]
```

## POST /api/v1/auth/login

### Request
**Headers:**
- `Content-Type: application/json`

**Body:**
```json
{
  "email": "user@example.com",
  "password": "SecurePass123!"
}
```

### Response

**Success (200 OK):**
```json
{
  "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "token_type": "Bearer",
  "expires_in": 900,
  "user": {
    "id": "uuid",
    "email": "user@example.com",
    "name": "John Doe"
  }
}
```

**Error (401 Unauthorized):**
```json
{
  "error": "invalid_credentials",
  "message": "Invalid email or password"
}
```

**Error (429 Too Many Requests):**
```json
{
  "error": "rate_limit_exceeded",
  "message": "Too many login attempts. Please try again in 15 minutes.",
  "retry_after": 900
}
```

### Rate Limiting
- 5 attempts per 15 minutes per IP address
- Returns 429 status with `Retry-After` header

### Swagger Annotation
```go
// @Summary      User login
// @Description  Authenticates user and returns JWT tokens
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request  body      LoginRequest  true  "User credentials"
// @Success      200      {object}  LoginResponse
// @Failure      401      {object}  ErrorResponse
// @Failure      429      {object}  ErrorResponse
// @Failure      500      {object}  ErrorResponse
// @Router       /api/v1/auth/login [post]
```

## JWT Token Structure

### Access Token Claims
```json
{
  "sub": "user-uuid",
  "email": "user@example.com",
  "name": "John Doe",
  "type": "access",
  "iat": 1704621000,
  "exp": 1704621900
}
```

### Refresh Token Claims
```json
{
  "sub": "user-uuid",
  "type": "refresh",
  "iat": 1704621000,
  "exp": 1705225800
}
```

## Error Response Schema

All error responses follow this structure:
```json
{
  "error": "error_code",
  "message": "Human-readable error message",
  "details": []  // Optional array of field-specific errors
}
```

## Security Headers

All responses include:
```
X-Content-Type-Options: nosniff
X-Frame-Options: DENY
X-XSS-Protection: 1; mode=block
Strict-Transport-Security: max-age=31536000; includeSubDomains
```
```

### docs/plans/issue-123/context.md
```markdown
# Context - Issue #123: User Authentication

## Background
The application currently has no user authentication mechanism. Users cannot create accounts or log in, which limits the application to public read-only access. This feature is a prerequisite for implementing user-specific features like favorites, personalized recommendations, and user management.

## Business Requirements
- Users should be able to register with email and password
- Password must meet security complexity requirements
- System must prevent brute force attacks
- Token-based authentication for stateless API access
- Tokens should be short-lived for security

## Stakeholder Notes

### From Product Manager (2025-01-05)
- Registration should be simple - just email, password, and name
- Consider adding social login (Google/GitHub) in future iteration
- Email verification can be added later, not in scope for this issue

### From Security Team (2025-01-06)
- Must use bcrypt for password hashing (minimum cost factor 12)
- Implement rate limiting on authentication endpoints
- JWT tokens should have short expiry (recommend 15 minutes)
- Refresh tokens should be revocable
- All authentication events must be logged for audit

### From Backend Lead (2025-01-06)
- Follow existing project structure: internal/auth for auth logic
- Reuse existing database connection patterns
- Write comprehensive tests including security tests
- Add Swagger documentation for all endpoints

## Related Issues
- #45 - User Profile Management (depends on this issue)
- #67 - Role-Based Access Control (depends on this issue)
- #89 - Email Verification (future enhancement)

## External References
- [OWASP Authentication Cheat Sheet](https://cheatsheetseries.owasp.org/cheatsheets/Authentication_Cheat_Sheet.html)
- [JWT Best Practices](https://tools.ietf.org/html/rfc8725)
- Project coding guidelines: `coding-guidelines.md`

## Success Criteria
- [ ] Users can register with email and password
- [ ] Users can log in and receive JWT tokens
- [ ] Tokens can be used to access protected endpoints
- [ ] Rate limiting prevents brute force attacks
- [ ] All security requirements are met
- [ ] Test coverage ≥80%
- [ ] Swagger documentation complete
```

### docs/plans/issue-123/implementation-plan.md
```markdown
# Implementation Plan - Issue #123: User Authentication

## Phase 1: Database Schema (Day 1, Morning)
**Goal**: Set up database tables for users and refresh tokens

### Tasks
1. ✅ Create migration script `20250107_001_create_users_table.sql`
   - Users table with email, password_hash, name
   - Index on email for fast lookups
2. ✅ Create migration script `20250107_002_create_refresh_tokens_table.sql`
   - Refresh tokens table with user_id, token_hash, expires_at
   - Indexes on user_id and expires_at
3. ✅ Test migrations on local database
4. ✅ Create rollback scripts

### Validation
- Run migrations successfully
- Verify indexes are created
- Test rollback

---

## Phase 2: User Model & Repository (Day 1, Afternoon)
**Goal**: Create data access layer for users

### Tasks
1. ✅ Create `internal/user/model.go`
   - User struct with validation tags
   - Custom error types (ErrUserNotFound, ErrDuplicateEmail)
2. ✅ Create `internal/user/repository.go`
   - UserRepository interface (following ISP principle)
   - PostgreSQL implementation
   - Methods: Create, FindByEmail, FindByID
3. ✅ Create `internal/user/repository_test.go`
   - Table-driven tests for all repository methods
   - Test duplicate email handling
   - Test not found scenarios

### Validation
- All repository tests pass
- Test coverage ≥80%
- golangci-lint passes

---

## Phase 3: Authentication Service (Day 2, Morning)
**Goal**: Implement core authentication logic

### Tasks
1. ✅ Create `internal/auth/service.go`
   - AuthService struct with dependencies (UserRepository)
   - HashPassword function (bcrypt, cost 12)
   - VerifyPassword function
   - GenerateTokens function (JWT access + refresh)
   - ValidateAccessToken function
   - RefreshAccessToken function
2. ✅ Create `internal/auth/service_test.go`
   - Test password hashing and verification
   - Test JWT token generation and validation
   - Test token expiry
   - Test invalid token scenarios

### Validation
- All service tests pass
- Passwords properly hashed
- JWT tokens generated correctly

---

## Phase 4: Authentication Middleware (Day 2, Afternoon)
**Goal**: Create middleware for protecting endpoints

### Tasks
1. ✅ Create `internal/auth/middleware.go`
   - RequireAuth middleware
   - Extract user from JWT token
   - Inject user context into request
2. ✅ Create `internal/auth/middleware_test.go`
   - Test valid token access
   - Test missing token rejection
   - Test invalid token rejection
   - Test expired token rejection

### Validation
- Middleware correctly validates tokens
- User context properly injected
- Error responses match API contract

---

## Phase 5: HTTP Handlers (Day 3, Morning)
**Goal**: Create API endpoints

### Tasks
1. ✅ Create `internal/auth/handler.go`
   - RegisterHandler
   - LoginHandler
   - RefreshHandler
   - LogoutHandler
   - Add Swagger annotations to all handlers
2. ✅ Create `internal/auth/handler_test.go`
   - Test all endpoints with table-driven tests
   - Test validation errors
   - Test authentication errors
   - Test success scenarios

### Validation
- All handler tests pass
- Proper HTTP status codes returned
- Response format matches API contract

---

## Phase 6: Rate Limiting (Day 3, Afternoon)
**Goal**: Implement rate limiting to prevent brute force

### Tasks
1. ✅ Create `internal/ratelimit/middleware.go`
   - In-memory rate limiter (can upgrade to Redis later)
   - Configurable limits per endpoint
   - Clean up expired entries periodically
2. ✅ Apply rate limiting to login and register endpoints
   - Login: 5 attempts per 15 minutes per IP
   - Register: 3 attempts per hour per IP
3. ✅ Test rate limiting behavior

### Validation
- Rate limits enforced correctly
- 429 status returned when exceeded
- Retry-After header included

---

## Phase 7: Integration & Testing (Day 4)
**Goal**: End-to-end testing and security validation

### Tasks
1. ✅ Wire up routes in `cmd/assistantd/main.go`
2. ✅ Write integration tests
   - Full registration flow
   - Full login flow
   - Token refresh flow
   - Protected endpoint access
3. ✅ Security testing
   - SQL injection attempts
   - XSS payload testing
   - Brute force simulation
   - Token tampering tests
4. ✅ Run full test suite
   - `go test ./...`
   - Verify coverage ≥80%
5. ✅ Run linter
   - `golangci-lint run`
   - Fix all warnings

### Validation
- All tests pass (unit + integration + security)
- Coverage ≥80%
- No linter warnings

---

## Phase 8: Documentation (Day 5, Morning)
**Goal**: Complete all documentation

### Tasks
1. ✅ Update Swagger documentation
   - Run `swag init` to regenerate docs
   - Test Swagger UI
2. ✅ Update `docs/architecture.md`
   - Add authentication system diagram
   - Document authentication flow
3. ✅ Update `CHANGELOG.md`
   - Add feature description
   - List all endpoints added
4. ✅ Create API usage examples
   - Example registration request
   - Example login request
   - Example authenticated request

### Validation
- Swagger UI displays all endpoints
- Documentation is clear and complete
- Examples work correctly

---

## Phase 9: Security Review & Deployment Prep (Day 5, Afternoon)
**Goal**: Final security review and production readiness

### Tasks
1. ✅ Security checklist review
   - [ ] Passwords hashed with bcrypt (cost 12)
   - [ ] JWT tokens properly signed
   - [ ] Rate limiting active
   - [ ] Input validation comprehensive
   - [ ] SQL injection prevented (prepared statements)
   - [ ] XSS prevented (input sanitization)
   - [ ] Audit logging implemented
   - [ ] No sensitive data in logs
2. ✅ Environment configuration
   - JWT secret from environment variable
   - Token expiry configurable
   - Database connection secure
3. ✅ Performance validation
   - Test with 100 concurrent requests
   - Verify response times acceptable
   - Check database query performance

### Validation
- Security checklist 100% complete
- Performance meets requirements
- Ready for deployment

---

## Rollback Plan

If issues are discovered after deployment:

1. **Immediate rollback**: Revert code deployment
2. **Database rollback**: Run rollback migration scripts
   ```sql
   DROP TABLE IF EXISTS refresh_tokens;
   DROP TABLE IF EXISTS users;
   ```
3. **Verify**: Ensure application works without auth system

## Post-Deployment Monitoring

**Week 1 after deployment:**
- Monitor registration success rate
- Monitor login success/failure rates
- Check for unusual authentication patterns
- Review security logs for attempted attacks
- Monitor API response times

**Alerts to configure:**
- High authentication failure rate (>20%)
- Unusual spike in registrations
- Rate limit frequently exceeded
- Database connection errors
```

---

## Instructions

When invoked with an issue number:

1. **Fetch the issue details** from GitHub
2. **Ensure the issue has an associated branch on GitHub**:
   - Use the GitHub MCP to check for an existing branch named `feature/issue-{NUMBER}-{description}`
   - If it does not exist, create a new branch on GitHub from the default branch using that naming convention
   - Link the branch to the GitHub issue in the Development section
3. **Tag the issue appropriately**:
   - Add the label `analysis` (create it first if missing)
   - Assign the issue to the `Assistant` project (e.g., under “Projects” → “Add to project”)
4. **Checkout the branch locally** so analysis artifacts are committed in the correct context:
   ```bash
   git fetch origin
   git checkout feature/issue-{NUMBER}-{description}
   ```
5. **Create the analysis documents** in `docs/plans/issue-{NUMBER}/`
6. **Provide the comprehensive analysis** following the format above

### Git Branch Workflow

Before starting analysis:
- Use the GitHub MCP to confirm the remote branch exists (`feature/issue-{NUMBER}-{description}`); create it from the default branch if missing
- Link the branch to the GitHub issue so progress is tracked under Development
- Apply the `analysis` label and ensure the issue lives in the `Assistant` project
- Fetch and checkout the branch locally once it exists on GitHub

**Directory Creation Steps:**
1. Create the directory: `docs/plans/issue-{NUMBER}/`
2. Always create: `analysis.md` (main analysis document)
3. Optionally create: `api-contract.md`, `context.md`, `implementation-plan.md` (based on issue complexity)
4. Inform the user which files were created and their purposes
