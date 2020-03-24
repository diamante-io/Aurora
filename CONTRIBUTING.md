# How to contribute

👍🎉 First off, thanks for taking the time to contribute! 🎉👍

Check out the [HcNet Contribution Guide](https://github.com/hcnet/docs/blob/master/CONTRIBUTING.md) for details.

## Style guides

### Git Commit Messages

* Use the present tense ("Add feature" not "Added feature")
* Use the imperative mood ("Move cursor to..." not "Moves cursor to...")

### Pull Requests

* PR titles start with package name, ex. “aurora/ingest: fix...”
* PRs must update the [CHANGELOG](CHANGELOG.md) with a small description of the change
* PRs are merged into master or release branch using squash merge
* Carefully think about where your PR fits according to [semver](https://semver.org). Target it at master if it’s only a patch change, otherwise if it contains breaking change or significant feature additions, set the base branch to the next major or minor release.
* Keep PR scope narrow. Expectation: 20 minutes to review max
* Explicitly differentiate refactoring PRs and feature PRs. Refactoring PRs don’t change functionality. They usually touch a lot more code, and are reviewed in less detail. Avoid refactoring in feature PRs.

### Go Style Guide

* Use `gofmt` or preferably `goimports` to format code
* Follow [Effective Go](https://golang.org/doc/effective_go.html) and [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
