# Project Guidelines

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Version Control: jj (Jujutsu)

This project uses **jj** (Jujutsu) for version control, colocated with git. **Always prefer jj commands over git commands.**

### Key jj Commands

```bash
# View status
jj status
jj log              # View commit history (better than git log)

# Creating commits
jj new              # Create a new empty change on top of current
jj commit -m "msg"  # Commit current working copy with message
jj describe -m "msg" # Set/update message for current change

# Stacked changes (the jj way)
jj new              # Start new change on top of current
jj new -A @-        # Insert new change AFTER parent (between parent and current)
jj new -B @         # Insert new change BEFORE current

# Moving between changes
jj edit <change>    # Edit a specific change
jj prev             # Move to parent change
jj next             # Move to child change

# Rebase and squash
jj rebase -d <dest> # Rebase current change onto destination
jj squash           # Squash current change into parent
jj squash --into <change> # Squash into specific change

# Working with branches
jj branch create <name>   # Create branch at current change
jj branch set <name>      # Move branch to current change
jj git push               # Push to git remote

# Sync with git
jj git fetch
jj git push
```

### Workflow: Stacked Atomic Commits

**Each logical change should be its own atomic commit.** When implementing features:

1. **Plan changes**: Break work into logical, atomic units before coding
2. **Create stacked changes**: Use `jj new` to create each change in sequence
3. **Describe each change**: Use `jj describe -m "..."` for clear commit messages
4. **Review the stack**: Use `jj log` to see your change stack
5. **Push when ready**: Use `jj git push` to push all changes


### Plan Design

When designing a plan to implement a feature / structuring a commit - these should be considered "vertical" slices that fully implement the feature - from the client (as necessary) all the way to the data layer.

For example: If a plan includes adding a "Edit" and "Delete" functionality, this would involve
- UI (to indicate in the client which behavior to take)
- Route handlers (on the server for accepting the request)
- Controllers to manage the request and isolate the business logic
- Data/storage layer to communicate with the database

In this case, we would have two commits:
1/ Edit: includes any necessary UI, route handler, controller, and data changes necessary to support the "edit" feature
2/ Delete: includes any necessary UI, route handler, controller, and data changes necessary to support the "delete" feature.

### Example: Creating a Stacked PR for Edit and Delete

```bash
# Start from main
jj new main -m "feat: edit: add user model fields"
# ... make changes ...

jj new -m "feat: edit: add repository method"
# ... make changes ...

jj new -m "feat: edit: add service layer"
# ... make changes ...

# View your stack
jj log

# Push all changes
jj branch create edit-feature
jj git push


# Re-Start from main
jj new main -m "feat: delete: add user model fields"
# ... make changes ...

jj new -m "feat: delete: add repository method"
# ... make changes ...

jj new -m "feat: delete: add service layer"
# ... make changes ...

# View your stack
jj log

# Push all changes
jj branch create delete-feature
jj git push
```

### Amending Changes in the Stack

With jj, you can easily edit any change in your stack:

```bash
# Edit a specific change
jj edit <change-id>
# ... make fixes ...
jj squash  # or just leave changes in place

# Go back to tip
jj edit <tip-change-id>
```

### Commit Message Format

```
<type>: <short description>

<optional longer description>

🤖 Generated with [Claude Code](https://claude.com/claude-code)

Co-Authored-By: Claude Opus 4.5 <noreply@anthropic.com>
```

Types: `feat`, `fix`, `refactor`, `test`, `docs`, `chore`

---

## Documentation Guidelines

When making changes to the codebase, keep documentation in sync.

### README.md Updates Required When:

- Adding new features or services
- Adding or modifying API endpoints
- Adding new configuration options
- Changing prerequisites or setup steps
- Modifying the project structure

### What to Update:

1. **Features list** - Add new features with ✅ checkmark
2. **Configuration** - Document new config options with defaults
3. **API Reference** - Add/update endpoint documentation with curl examples
4. **Testing** - Document new test scripts or testing approaches
5. **Architecture** - Update diagrams if data flow changes

### Example: Adding a New Feature

When adding a feature like "scheduled transfers":

1. Add to Features section: `- ✅ **Scheduled Transfers** - Weekly or monthly batched transfers`
2. Add config reference for `scheduler.*` options
3. Add any new API endpoints to API Reference
4. Add test commands/scripts to Testing section
5. Update architecture diagrams if needed


## Project Structure

TO BE ADDED


## Build & Test

TO BE ADDED
