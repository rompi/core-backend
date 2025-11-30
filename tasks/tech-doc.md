# Command: /tech-doc

# Instructions

1. **Parse the arguments**: Extract the issue number or topic from the command arguments (e.g., `/tech-doc #123` or `/tech-doc issue-5`)

2. **Gather context**:
   - If an issue number is provided, fetch the issue details using `gh issue view <issue-number>`
   - Review git changes related to the issue using `git log`, `git diff`, and `git show`
   - Search for related files, particularly in `docs/plans/` if they exist
   - Identify modified files using `git status` and `git diff` commands

3. **Analyze the changes**:
   - Examine modified files to understand the scope and impact
   - Identify API changes (new endpoints, modified routes, request/response changes)
   - Identify database changes (new tables, schema modifications, migrations)
   - Identify configuration changes (new env vars, config parameters)
   - Determine which domains/features are affected

4. **Generate the technical documentation** with the following structure:

   ## Summary
   Provide a high-level overview of the changes that is accessible to non-technical stakeholders. Include:
   - What was changed and why
   - Which domains/features are impacted
   - The business value or problem being solved

   ## Changes Objective
   Describe the specific goals and objectives of these changes:
   - What problem is being solved
   - What new functionality is being added
   - What is being improved or refactored

   ## Domain Impacted
   Detail the specific areas affected by the changes:

   ### Features Affected
   - List each feature or module that is impacted
   - Describe what changed in each feature

   ### API Changes
   **Before:**
   - Document the previous API state (if applicable)
   - Include endpoints, request/response formats

   **After:**
   - Document the new API state
   - Include new/modified endpoints, request/response formats
   - Highlight breaking changes

   ### Database Changes
   **Before:**
   - Document the previous schema (if applicable)

   **After:**
   - Document new/modified tables, columns, indexes
   - Include migration details if applicable

   ### Configuration Changes
   **Before:**
   - List previous configuration parameters

   **After:**
   - List new/modified environment variables
   - List new/modified config parameters
   - Include default values and descriptions

   ## Additional Configuration
   Provide step-by-step instructions for any required configuration:
   - Environment variables to set
   - Config files to update
   - Dependencies to install
   - Services to configure

   ## Technical Details
   - Implementation approach
   - Key architectural decisions
   - Dependencies added or updated
   - Testing strategy

   ## Author
   - Identify the primary author(s) from git commits
   - Include commit references

5. **Save the documentation**:
   - Create the `docs/tech-summary` directory if it doesn't exist
   - Save the file as `docs/tech-summary/<issue-or-topic>.md`
   - Use the issue number or topic as the filename (e.g., `issue-5.md` or `feature-name.md`)

6. **Important guidelines**:
   - Be specific and detailed, not generic
   - Use actual code examples where relevant
   - Include before/after comparisons for changed functionality
   - Highlight breaking changes prominently
   - Keep the language clear and technical but accessible
   - Use proper markdown formatting
   - Include code blocks with proper syntax highlighting

# Example Usage

```
/tech-doc #123
/tech-doc issue-5
/tech-doc parser-service
```

# Output

After generating the documentation, inform the user:
- Where the file was saved
- A brief summary of what was documented
- Any notable findings or concerns
- Generate the markdown in `docs/tech-summary` folder