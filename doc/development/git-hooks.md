# Git Hooks for the Globus Go SDK

This document explains the Git hooks available in the repository and how to install and use them.

## Available Hooks

### Pre-Commit Hook

The pre-commit hook runs the following checks before each commit:

1. License header verification
2. Code formatting with `go fmt`
3. Code linting with `staticcheck` (if installed)
4. Static analysis with `go vet`
5. Unit tests in short mode

If any of these checks fail, the commit will be blocked.

### Pre-Push Hook

The pre-push hook runs more comprehensive checks before pushing code:

1. All tests (including integration tests)
2. Documentation checks (if applicable)
3. Security scan (if available)

If these checks fail, you'll be prompted whether to continue with the push or abort.

## Installing the Hooks

To install all hooks at once, run:

```bash
./scripts/install-all-hooks.sh
```

To install only specific hooks:

- Pre-commit hook only: `./scripts/install-hooks.sh`
- Pre-push hook only: `./scripts/install-pre-push-hook.sh`

## Bypassing Hooks

In certain situations, you may need to bypass the hooks, for example, when committing work-in-progress code that doesn't pass all checks yet. To bypass hooks, add the `--no-verify` flag to your git command:

```bash
git commit --no-verify -m "WIP: ..."
git push --no-verify
```

**Note:** Bypassing hooks should be used sparingly. The hooks are designed to maintain code quality and prevent issues from reaching the repository.

## Hook Maintenance

The hook scripts are stored in the `.git/hooks` directory of your local repository. If hook behaviors need to be modified, update the installation scripts and run them again to reinstall the hooks.

When updating the hooks, the old hooks are backed up with timestamps for easy restoration if needed.