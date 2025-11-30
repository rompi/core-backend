# Create GitHub Task

This document outlines the standard task types and process for creating well-structured GitHub issues in the AI Assistant Backend project.

## Task Types

### 1. Feature Development Tasks
**Label**: `type/feature`
- New functionality implementation
- API endpoint creation
- Database schema changes
- Integration with external systems

### 2. Bug Fix Tasks
**Label**: `type/bug`
- Production issues
- Regression bugs
- Security vulnerabilities
- Performance problems

### 3. Enhancement Tasks
**Label**: `type/enhancement`
- Code optimization
- Performance improvements
- Refactoring work
- Documentation updates

### 4. Testing Tasks
**Label**: `type/testing`
- Unit test creation/updates
- Integration test development
- Performance test implementation
- Security test cases

### 5. Infrastructure Tasks
**Label**: `type/infrastructure`
- CI/CD pipeline updates
- Development environment setup
- Deployment configuration
- Monitoring setup

### 6. Documentation Tasks
**Label**: `type/docs`
- API documentation
- Code documentation
- Architecture documentation
- User guides

### 7. Security Tasks
**Label**: `type/security`
- Security audit fixes
- Authentication improvements
- Authorization updates
- Input validation enhancements

### 8. Maintenance Tasks
**Label**: `type/maintenance`
- Dependency updates
- Code cleanup
- Technical debt reduction
- Database migrations

## Issue Creation Guidelines

You are helping to create a new GitHub issue (task) that clearly describes what needs to be built from a **user's perspective**.

### Your Goal

Create a well-structured GitHub issue with:
- **Clear, user-focused title** - What will users be able to do?
- **User story** - Why does the user need this?
- **User journey** - Step-by-step flow from the user's perspective
- **Acceptance criteria** - How users will know it works
- **What's in scope / out of scope** - Clear boundaries
- **Technical details** - Only when the task is inherently technical

## Core Principles

1. **User-Centric**: Start with user needs, not technical implementation
2. **Clear Boundaries**: Explicitly state what's included and what's not
3. **Measurable Success**: Define how to verify it works from user perspective
4. **Simple Language**: Write for product managers and junior engineers alike

## Task Template (User-Focused)

Every task you create should follow this structure:

```markdown
## User Story
As a [type of user], I want to [do something], so that [I get some benefit].

## Why This Matters
[Explain the user problem or need this solves - 1-2 sentences]

## User Journey

### Current Experience (Before)
1. User [does action A]
2. User [experiences problem B]
3. User [can't achieve goal C]

### Expected Experience (After)
1. User [does action A]
2. System [responds with B]
3. User [successfully achieves C]

## Acceptance Criteria (User Perspective)

**Must Have:**
- [ ] User can [specific action]
- [ ] User sees [specific feedback/result]
- [ ] User cannot [specific restriction]

**Should Have:**
- [ ] User receives [helpful message/feedback]
- [ ] User experience is [smooth/fast/intuitive]

**Won't Have (Out of Scope):**
- User does NOT need [feature X] in this task
- [Feature Y] will be handled in a separate task
- [Functionality Z] is planned for future iteration

## Scope

### In Scope ✅
- [Feature A]
- [Feature B]
- [Feature C]

### Out of Scope ❌
- [Feature X] - Will be addressed in issue #[NUMBER]
- [Feature Y] - Planned for future sprint
- [Feature Z] - Not required for MVP

## Success Metrics (How to Verify)

From user perspective:
- [ ] User can complete [action] in [X] steps
- [ ] User sees success message: "[exact message]"
- [ ] User gets result within [X] seconds

## Additional Context

[Any background information, screenshots, mockups, or user feedback]

## Technical Notes (Only if Task is Technical)

If this task is inherently technical (API, backend, infrastructure):

### Technical Requirements
- [Specific technical requirement]
- [Performance requirement]
- [Security requirement]

### Files to Modify
- `path/to/file.go` - [What needs to change]

### Testing Requirements
- [ ] Unit tests for [functionality]
- [ ] Integration test for [user flow]
```

## Instructions

When the user runs `/create-github-task [description]`:

### Step 1: Understand the User Need

Ask clarifying questions if needed:
- "Who is the user for this feature?" (end user, admin, developer, etc.)
- "What problem are they trying to solve?"
- "What should happen from the user's perspective?"
- "What should NOT be included in this task?" (out of scope)

### Step 2: Write User Story

Create a user story in this format:
```
As a [user type],
I want to [action],
So that [benefit].
```

**Examples:**
- ✅ "As a customer, I want to reset my password, so that I can regain access to my account"
- ✅ "As an admin, I want to view all user registrations, so that I can monitor system usage"
- ❌ "As a system, I want to hash passwords" (not user-centric)

### Step 3: Define User Journey

Describe the step-by-step flow from user's perspective:

**Current Experience:**
- What does the user do now?
- What problem do they encounter?
- What's the pain point?

**Expected Experience:**
- What will the user do?
- What will they see?
- What will the outcome be?

### Step 4: Write Acceptance Criteria (User POV)

Focus on **observable user outcomes**, not technical implementation:

**Good (User POV):**
- ✅ "User sees error message 'Email is required' when submitting empty email"
- ✅ "User receives confirmation email within 1 minute of registration"
- ✅ "User's password reset link works for 1 hour"

**Bad (Technical POV):**
- ❌ "System validates email using regex"
- ❌ "Email service sends via SMTP"
- ❌ "Token expires after 3600 seconds"

### Step 5: Define Scope Boundaries

**CRITICAL**: Always include "Out of Scope" section.

If you're unsure what's out of scope, ask the user:
- "Should [related feature X] be included in this task?"
- "Is [feature Y] part of this or a separate task?"
- "What about [edge case Z] - in scope or future work?"

**Examples:**

**In Scope:**
- User can reset password
- User receives reset email
- User can set new password

**Out of Scope:**
- Social login (separate task)
- Email verification for new accounts (future work)
- Two-factor authentication (not in MVP)

### Step 6: Add Technical Details (Only When Necessary)

Only add technical section if:
- Task is inherently technical (API endpoint, database migration, etc.)
- Junior engineers need specific guidance
- There are performance or security requirements

Otherwise, keep it user-focused.

### Step 7: Create the Issue

Use GitHub MCP tools:
```
mcp__github__create_issue(
  owner: "samasvva",
  repo: "ai-assistant-backend",
  title: "[User-focused title]",
  body: "[Well-structured description using template]",
  labels: ["feature", "user-story"],
  assignees: []
)
```

### Step 8: Create Git Branch

After creating the issue, create a git branch for the work:

1. **Determine branch name** based on issue type and number:
   - Feature: `feature/issue-{NUMBER}-{short-description}`
   - Bug fix: `bugfix/issue-{NUMBER}-{short-description}`
   - Hotfix: `hotfix/issue-{NUMBER}-{short-description}`
   - Enhancement: `enhancement/issue-{NUMBER}-{short-description}`

2. **Create and push branch**:
   ```bash
   git checkout -b feature/issue-{NUMBER}-{short-description}
   git push -u origin feature/issue-{NUMBER}-{short-description}
   ```

**Branch naming conventions:**
- Use lowercase with hyphens
- Keep description short (3-5 words max)
- Example: `feature/issue-123-password-reset`
- Example: `bugfix/issue-45-fix-null-pointer`

## Examples

### Example 1: User-Focused Feature (Non-Technical)

**Title**: "User can reset forgotten password via email"

```markdown
## User Story
As a registered user who forgot my password,
I want to reset my password using my email,
So that I can regain access to my account without contacting support.

## Why This Matters
Currently, users who forget their password have no way to recover their account. They must contact support, causing frustration and support ticket overhead.

## User Journey

### Current Experience (Before)
1. User forgets password
2. User tries to login but fails
3. User has no option to reset password
4. User must contact support and wait for help

### Expected Experience (After)
1. User clicks "Forgot Password?" on login page
2. User enters their email address
3. User receives email with reset link
4. User clicks link and sets new password
5. User can login with new password

## Acceptance Criteria (User Perspective)

**Must Have:**
- [ ] User sees "Forgot Password?" link on login page
- [ ] User can enter email address on reset page
- [ ] User receives email with reset link within 2 minutes
- [ ] User can click link and set new password
- [ ] User sees success message: "Password successfully reset"
- [ ] User can login with new password
- [ ] Reset link stops working after 1 hour
- [ ] Reset link stops working after being used once

**Should Have:**
- [ ] User sees helpful message if email doesn't exist (without revealing this)
- [ ] User sees password requirements (min 8 characters, etc.)
- [ ] User receives confirmation that password was changed

**Won't Have (Out of Scope):**
- User does NOT need to verify their email first (handled separately)
- SMS-based password reset (future feature)
- Security questions (not in MVP)

## Scope

### In Scope ✅
- "Forgot Password?" link on login page
- Email with password reset link
- Password reset form
- Password validation
- Success/error messages

### Out of Scope ❌
- Email verification for new accounts - Issue #45
- Two-factor authentication - Future sprint
- SMS/phone-based reset - Not planned for MVP
- Password strength meter - Nice-to-have, separate task

## Success Metrics (How to Verify)

From user perspective:
- [ ] User completes password reset in 4 steps or less
- [ ] User receives email within 2 minutes
- [ ] User sees clear success message after reset
- [ ] Old password stops working immediately
- [ ] New password works immediately

## Additional Context

**User Feedback:**
"I forgot my password and couldn't find any way to reset it. Had to email support and wait 2 days." - User feedback from survey

**Security Note:**
Don't reveal whether an email exists in the system (prevent user enumeration).
```

### Example 2: Technical Task (Bug Fix)

**Title**: "Fix API returning 500 error when user submits empty email"

```markdown
## User Story
As an API consumer,
I want to receive proper error messages when I send invalid data,
So that I can fix my request and understand what went wrong.

## Why This Matters
Currently, the API crashes (500 error) when email is empty. Users see a generic error and don't know what's wrong. This creates a poor developer experience.

## User Journey

### Current Experience (Before)
1. Developer sends POST /api/auth/login with empty email
2. API returns 500 Internal Server Error
3. Developer sees generic error, doesn't know what's wrong
4. Developer must guess what field is invalid

### Expected Experience (After)
1. Developer sends POST /api/auth/login with empty email
2. API returns 400 Bad Request
3. Developer sees error message: "Email is required"
4. Developer fixes the request and succeeds

## Acceptance Criteria (User Perspective)

**Must Have:**
- [ ] API returns 400 Bad Request (not 500) when email is empty
- [ ] Error message clearly states: "Email is required"
- [ ] API doesn't crash (no null pointer errors)
- [ ] Response format matches standard error format

**Should Have:**
- [ ] Error response includes field name: `{"error": "validation_error", "field": "email"}`
- [ ] Multiple validation errors shown together (if both email and password missing)

**Won't Have (Out of Scope):**
- Email format validation (only check if empty in this task)
- Rate limiting (separate security task)
- Detailed password validation (separate task)

## Scope

### In Scope ✅
- Fix null pointer error when email is empty
- Return proper 400 error code
- Return clear error message
- Validation for required fields

### Out of Scope ❌
- Email format validation (e.g., must contain @) - Issue #47
- Password strength validation - Issue #48
- Rate limiting for invalid attempts - Issue #49

## Success Metrics (How to Verify)

From API consumer perspective:
- [ ] Sending empty email returns 400 (not 500)
- [ ] Error message is clear and actionable
- [ ] API response time < 100ms
- [ ] No crashes in logs

## Technical Notes

This is a bug fix, so technical details are relevant:

### Technical Requirements
- Add validation before processing login request
- Return structured error response
- Log validation failures at INFO level (not ERROR)

### Files to Modify
- `internal/adapters/http/handler/auth.go` - Add validation
- `internal/adapters/http/handler/auth_test.go` - Add test cases

### Testing Requirements
- [ ] Unit test: empty email returns 400
- [ ] Unit test: empty password returns 400
- [ ] Unit test: both empty returns 400 with both errors
- [ ] Integration test: full request flow
```

### Example 3: Admin Feature

**Title**: "Admin can view list of all registered users"

```markdown
## User Story
As an admin,
I want to view a list of all registered users,
So that I can monitor system usage and user growth.

## Why This Matters
Admins currently have no visibility into who is using the system. They need basic user information to monitor growth and identify issues.

## User Journey

### Current Experience (Before)
1. Admin logs into admin panel
2. Admin has no way to see user list
3. Admin must query database directly (requires technical knowledge)

### Expected Experience (After)
1. Admin logs into admin panel
2. Admin clicks "Users" in navigation menu
3. Admin sees list of all users with basic info
4. Admin can search/filter users by email or name

## Acceptance Criteria (User Perspective)

**Must Have:**
- [ ] Admin sees "Users" menu item in admin panel
- [ ] Admin sees list of users with: name, email, registration date
- [ ] Admin sees total user count
- [ ] List shows 20 users per page with pagination
- [ ] Admin can search users by email or name

**Should Have:**
- [ ] Admin sees user status (active/inactive)
- [ ] Admin can sort by registration date
- [ ] List loads in under 2 seconds

**Won't Have (Out of Scope):**
- Admin does NOT edit user details in this task
- Admin does NOT delete users in this task
- Admin does NOT see user activity logs in this task
- Export to CSV (separate feature)

## Scope

### In Scope ✅
- View user list (read-only)
- Basic user information (name, email, date)
- Search by email or name
- Pagination (20 per page)
- User count

### Out of Scope ❌
- Edit user details - Issue #52
- Delete users - Issue #53
- User activity logs - Future feature
- Export to CSV - Future feature
- Advanced filtering (by date range, status) - Future enhancement

## Success Metrics (How to Verify)

From admin perspective:
- [ ] Admin can find any user by email in under 3 clicks
- [ ] List loads in under 2 seconds even with 10,000 users
- [ ] Admin can see newest users first (sorted by date)

## Additional Context

**Admin Feedback:**
"I need to see who's signing up. Right now I have to ask engineering to run SQL queries." - Admin team lead

**Note:**
This is read-only view. User editing/deletion will come in future tasks.
```

## Common Pitfalls to Avoid

### 1. Too Technical, Not User-Focused
❌ **Bad**: "Implement JWT authentication middleware"
✅ **Good**: "User stays logged in for 7 days without re-entering password"

### 2. Missing Out of Scope
❌ **Bad**: Only lists what's included
✅ **Good**: Explicitly states what's NOT included and why

### 3. Vague Acceptance Criteria
❌ **Bad**: "User can reset password"
✅ **Good**: "User receives reset email within 2 minutes and can set new password"

### 4. No User Journey
❌ **Bad**: Lists features without context
✅ **Good**: Shows step-by-step user flow before and after

### 5. Unclear Scope
❌ **Bad**: "Build user management"
✅ **Good**: "Admin can view user list (read-only). Editing and deletion are separate tasks."

## When to Ask About Scope

If you're unsure what should be in/out of scope, ask the user:

**Questions to ask:**
- "Should [feature A] be included, or is that a separate task?"
- "Is [edge case B] in scope for this task?"
- "What about [related feature C] - part of this or future work?"
- "Are there any features that should NOT be in this task?"

**Example:**
```
I'm creating a task for password reset. I need to clarify the scope:

Questions:
1. Should email verification be required before password reset?
2. Should we support SMS-based reset, or only email?
3. What about two-factor authentication - in scope or separate?
4. Should users be logged out from all devices after password reset?

Please clarify so I can define clear boundaries.
```

## Quality Checklist

Before creating the issue, verify:
- [ ] Title describes user benefit (not technical implementation)
- [ ] User story follows "As a [user], I want [action], so that [benefit]" format
- [ ] User journey shows before/after experience
- [ ] Acceptance criteria are observable user outcomes
- [ ] **"Out of Scope" section is included**
- [ ] Scope boundaries are clear
- [ ] Success metrics are user-focused
- [ ] Technical details only included if task is inherently technical
- [ ] Junior engineer AND product manager can understand it

## Repository Information

**Default Repository**: `samasvva/ai-assistant-backend`

Always use:
- Owner: `samasvva`
- Repo: `ai-assistant-backend`

## Output Format

After successful creation:
```
✓ Created issue #123: User can reset forgotten password via email
  URL: https://github.com/samasvva/ai-assistant-backend/issues/123
  Labels: feature, user-story

✓ Created git branch: feature/issue-123-password-reset
  Branch pushed to origin

The issue includes:
- Clear user story and journey
- Observable acceptance criteria
- Defined scope (in/out)
- Success metrics from user perspective

Next steps:
1. Review and confirm scope is correct
2. Add any missing context or screenshots
3. Run /analyze-github-task 123 to create analysis documents
4. Assign when ready for implementation
```

---

**Remember**: Start with the user's perspective. What do they want to do? What will they experience? What's in scope and what's not? Keep it simple and clear.
