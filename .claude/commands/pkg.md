# /pkg Command

You are in **Package Planning Mode**. Your job is to craft a complete implementation plan for introducing a new, standalone Go package inside the repository’s `/pkg` directory that any of the user’s own GitHub repositories can import and use with minimal effort.

## Mission

1. Understand the package’s purpose, responsibilities, and target consumers. If the user hasn’t already provided:
   - Ask for the package name, high-level feature it solves, and the primary use cases.
   - Ask what assumptions we can make about the data or environment (e.g., blocking/async, persistence, required inputs/outputs, configuration).
   - Ask if there are non-goal constraints such as zero external dependencies, backwards compatibility, or release/versioning expectations.
2. Produce a structured plan (see “Plan template” below) that covers design, implementation, validation, delivery, and documentation. Do **not** implement the package yet; this command only creates the plan.

## Standalone & Reusable Guidance

The plan must explicitly explain how the package will be:

- **Self-contained**: describe the minimal dependencies, interfaces to the rest of this repo, and how internal state is isolated inside `/pkg`.
- **Easy to consume from other repos**: include the go module path/namespace, usage examples for external clients, documentation checklist, and any release steps (tagging, versioning, `go install` instructions).
- **Well documented and tested**: cover README/GoDoc, sample usage snippets, and the test strategy a consumer would expect.

## Plan Template

When responding, fill out this structure:

1. **Context & Goals** – summarize the requested capability, business/technical justification, and success criteria.
2. **In-scope / Out-of-scope** – clarify boundaries so downstream implementers know what this plan covers.
3. **API surface & contracts** – describe exported types/functions, configuration options, error handling, and concurrency expectations.
4. **Internal architecture** – outline packages or files under `/pkg`, dependency graph, interfaces, and helpers.
5. **Integration with other repos** – detail module path, release cadence, versioning strategy, and example import statements or helper scripts for consumers.
6. **Testing & validation** – list unit, integration, and (if needed) benchmark or fuzz tests, along with tooling (e.g., `go test`, `golangci-lint`).
7. **Documentation & examples** – specify README sections, GoDoc comments, quick-start usage, and any docs updates.
8. **Risks & mitigations** – call out major uncertainties (thread safety, platform support, dependency upgrades) and how to address them.
9. **Delivery plan** – break the work into concrete steps/tasks (including doc creation, coding, tests, QA), with rough sequencing.

If any required input is missing, ask the user before producing the plan. Once you have all necessary information, deliver the plan via this template so the user can turn it into `docs/plans/...` later.
