# Coding Practices

This document outlines the coding practices for the Effeta Santito (effeta-santito) project. The goal is to maintain a codebase that is easy to understand, maintain, and extend. We follow a functional programming style, with a focus on clean interfaces and simplicity.

## Tracking work
Always have a tasks.md file with [ ] markdown style checkboxes.
Keep your progress in there and when you complete tasks check the [ ] box.
Add new tasks as needed.

## Functional Programming

We favor a functional programming style, which means we strive to write pure functions whenever possible. Pure functions are functions that have the following properties:

- They always return the same output for the same input.
- They have no side effects, such as modifying global state or performing I/O.

Benefits of functional programming include:

- **Predictability:** Pure functions are easier to reason about and test.
- **Composability:** Pure functions can be easily combined to create more complex functionality.
- **Concurrency:** Pure functions are inherently thread-safe, making it easier to write concurrent code.

## Clean Interfaces

We believe in the importance of clean interfaces. Interfaces should be small, focused, and easy to understand. They should hide the implementation details of a component and expose only the necessary functionality.

## "Out of the Tar Pit" Practices

We follow the principles outlined in the paper "Out of the Tar Pit" to manage complexity in our codebase. These principles include:

- **Immutability:** We favor immutable data structures. Instead of modifying existing data, we create new data structures with the updated values.
- **Simplicity:** We strive to keep our code simple and easy to understand. We avoid unnecessary complexity and clever tricks.
- **Explicitness:** We make dependencies explicit by passing them as arguments to functions. This makes our code easier to test and reason about.

Also, avoid overly abstract or premature optimizations...
Abstractions are "discovered", follow the rule of three
    If there are 2 use cases, just duplicate code, add comments
    If there is a 3rd use case, then refactor and abstract

---
# Golang

## Tests
Do not create gomodcache or any other tmp directory when running tests
Run ``gotest ./...`` from the root of the golang module and keep debugging from there


## Error Handling

Errors are handled by returning them from functions, which is the standard Go way. We avoid using `panic` for error handling, as it can lead to unexpected crashes. Instead, we return errors to the caller, allowing them to handle the error in a way that is appropriate for their context.

When you encounter legacy code that still uses `panic`, treat it as technical debt. Prefer refactoring those call sites to surface errors explicitly.

## Configuration

Configuration is managed through environment variables. The `pkg.Config` struct provides a single source of truth for all configuration. This makes it easy to manage configuration for different environments (e.g., development, staging, production).

Use godotenv to load env vars


``golang
type Config struct {
	// Ticket Tailor
	TicketTailorAPIKey  string `env:"TICKETTAILOR_API_KEY,required"`
	TicketTailorEventId int    `env:"TT_EVENT_ID,required"`
	TicketTailorBaseUrl string `env:"TT_BASE_URL,required"`
}

....

err := godotenv.Load()
	if err != nil {
		logger.Logger.Fatal("Error loading .env file", zap.Any("err", err))
	}

	cfg := pkg.Config{}

``



## Logging

We use the `zap` library for structured logging. Structured logging allows us to easily search and analyze our logs. We log important events, such as incoming requests, outgoing requests, and errors.

## Testing

New behavior should come with focused tests whenever practical. Run `go test ./...` from the `src` directory before submitting changes to ensure the existing suite stays green. For logic that interacts with external services, prefer fakes or explicitly injected interfaces so that tests remain deterministic.

## Project Structure

The project follows the standard Go project layout, with code separated into `cmd` and `pkg` directories. This makes the code easy to navigate and understand.

- `cmd`: Contains the main application entry points.
- `pkg`: Contains reusable libraries and packages.

