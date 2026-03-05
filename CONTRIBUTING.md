# Contributing to urai

Thank you for considering a contribution. All forms of help are welcome —
bug reports, feature suggestions, documentation improvements, and code.

## Reporting bugs

Open an issue and include:

- Your OS and terminal emulator
- Go version (`go version`)
- Steps to reproduce
- What you expected vs. what happened

## Suggesting features

Open an issue with the `enhancement` label. Describe the use-case first —
what problem does it solve for a writer or screenwriter?

## Submitting code

1. Fork the repository and create a branch from `master`.
2. Make your changes inside `prose/` (the Go module root).
3. Run `go build ./...` and `go vet ./...` — both must pass cleanly.
4. Keep commits focused; one logical change per commit.
5. Open a pull request with a clear description of what changed and why.

## Code style

- Standard Go formatting (`gofmt`).
- No external linters required, but `go vet` must be clean.
- Prefer small, readable functions over clever one-liners.
- New packages go under `prose/internal/`.

## Commit messages

Use the imperative mood in the subject line (`Add X`, `Fix Y`, `Remove Z`).
Keep the subject under 72 characters. Add a body if the why is not obvious.

## License

By contributing you agree that your work will be released under the
[MIT License](LICENSE).
