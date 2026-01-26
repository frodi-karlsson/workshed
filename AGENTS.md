# AGENTS.md

## Minimal Template

For new workspaces, use this minimal template:

```markdown
# Task

## Goal

## Commands

## Notes
```

See [github.com/frodi/workshed/AGENTS.md](https://github.com/frodi/workshed/blob/main/AGENTS.md) for a comprehensive reference.

---

## Running

Make religious use of the Makefile. It defines all common tasks.
Whenever you're verifying changes, run `go clean -testcache && make check`.

## Philosophy

**Simplicity births reliability.** Favor the simplest solution that clearly solves the problem. Cleverness, over-engineering, and premature abstraction reduce robustness.

No shortcuts. If something should be addressed, it **must be addressed**. Do not defer necessary work through TODOs, hacks, or partial solutions that knowingly leave the system in a worse state.

Aim for code that explains itself through structure and naming. Comments are a **rare necessity**, not a default. If the code needs many comments, the design likely needs improvement.

Duplication is usually a smell, but **abstracting too early is also a mistake**. Tolerate small, local duplication until the right abstraction becomes obvious. The balance matters.

---

## Code Guidelines

* Prefer clear, explicit code over dense or magical constructs.
* Choose descriptive names for variables, functions, and types; they are the primary form of documentation.
* Keep functions small and focused on a single responsibility.
* Avoid global state and hidden side effects.
* Make invalid states unrepresentable when possible.
* Handle errors explicitly and predictably.

---

## Testing Philosophy

Tests exist to increase confidence, not to inflate numbers.

* Each test should verify **one clear behavior**.
* High coverage is a goal, but **simplicity is the higher priority**.
* Do not test what is obvious or guaranteed by the language/runtime.
* Avoid excessive mocking. If you need dozens of mock files, something is wrong—either in the code design or the testing strategy.
* Prefer testing real behavior over implementation details.

Tests should be easy to read, easy to trust, and easy to delete when they no longer provide value.

---

## Design Smells to Watch For

* Large abstractions with only one implementation
* Complex test setups that dwarf the code under test
* Comments explaining *why* the code is confusing
* Duplication removed too early, resulting in indirection without clarity

When in doubt, step back and simplify.

---

## Final Note

Reliability comes from restraint. Write less code. Write clearer code. Let structure, not commentary, do the talking.
