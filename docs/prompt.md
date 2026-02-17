# Build forg — Smart File Organizer CLI

You are building a **portfolio project** for a Senior AI Engineer's public GitHub. It must be impressive, clean, and production-grade. Read these docs before writing any code:

1. **`docs/G02-go-cli-tool.md`** — Complete project spec: architecture, phases, design decisions, commit plan. This is your primary blueprint. Follow it phase by phase.
2. **`docs/github-portfolio.md`** — Portfolio goals and Definition of Done (Level 1 + Level 2). Understand the quality bar.
3. **`docs/github-portfolio-checklist.md`** — Pre-publish checklist. Every item must pass before you're done.

---

## Instructions

### Read first, build second
Read all three docs completely before writing a single line of code. Understand the architecture, the phases, the quality expectations.

### Follow the phases in order
The project spec has 5 phases. Do them in order:
1. **Scaffold & Core Logic** — project setup, config parsing, scanner, rules engine, organizer
2. **Commands & UX** — Cobra commands, output formatting, undo log
3. **Comprehensive Tests** — unit tests (table-driven), integration test, edge cases
4. **Refactor for Elegance** — DRY, interfaces, doc comments, lint clean
5. **Documentation & Polish** — README, LICENSE, final checklist pass

### Use subagents
This is a substantial project. Use subagents to parallelize where it makes sense:
- One subagent for the rules engine + tests while another does the scanner + tests
- One subagent for the Cobra commands while another writes the integration test
- A dedicated subagent for the refactoring pass (review all code for DRY, naming, clarity)
- A dedicated subagent for README + documentation

### Commit frequently
Follow the commit plan in the spec. Use **conventional commits** (`feat:`, `test:`, `refactor:`, `docs:`, `chore:`). Each commit should be a logical unit. Do NOT accumulate a massive uncommitted diff.

### Quality non-negotiables
- **Tests must be real.** Table-driven tests. Test behavior, not implementation. Happy path + error cases + edge cases. Tests must actually run and pass.
- **No fake anything.** No placeholder tests. No "TODO" comments. No stubbed-out functions. Everything works.
- **Lint clean.** Run `golangci-lint run` and fix everything. Run `gofmt` on all files.
- **Error handling.** Every error is handled. User-friendly messages. Wrapped with `fmt.Errorf("context: %w", err)`.
- **DRY.** The refactoring phase is mandatory. Extract shared patterns. Use interfaces where they help testability (e.g., `Matcher` interface, `FileSystem` interface for testing).

### Final verification
Before you consider the project done:
1. `go build ./...` — compiles clean
2. `go test ./... -race` — all tests pass, no race conditions
3. `golangci-lint run` — no issues
4. `go vet ./...` — no issues
5. Walk through `docs/github-portfolio-checklist.md` item by item
6. Read the README as if you're a hiring manager seeing it for the first time — does it make sense in 30 seconds?
7. Review git log — does the commit history tell a coherent story?

### What NOT to do
- Don't skip the refactoring phase. It's where good code becomes great code.
- Don't write tests after everything else. Write them alongside or immediately after each component.
- Don't leave `// TODO` or `// FIXME` comments anywhere.
- Don't hardcode any personal paths, usernames, or data.
- Don't use `any` types or skip error handling to save time.
- Don't commit the binary, `.DS_Store`, or any generated files.
- Don't use Docker. No Dockerfile, no docker-compose. Just a clean Go binary built with `go build`.

---

## GitHub Username

The GitHub username is **devaloi**. For Go module paths, use `github.com/devaloi/forg`. All internal imports must use this module path. Do not guess or use any other username.

## Fix Before Starting

The current `go.mod` uses the wrong module path. Before doing anything else:
1. Update `go.mod` module to `github.com/devaloi/forg`
2. Find/replace `github.com/jasonaloi/forg` → `github.com/devaloi/forg` across ALL `.go` files
3. Run `go build ./...` to verify
4. Commit: `fix: correct module path to github.com/devaloi/forg`

## Start

Read the three docs. Fix the module path above. Then review the existing code against the spec and continue from wherever the previous agent left off.
