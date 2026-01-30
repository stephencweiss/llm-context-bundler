# Product Spec: LLM Context Bundler (`lcb`)

### 1. Objective

Create a CLI tool that recursively scans a directory for Markdown files and bundles them into a single `.md` file optimized for LLM context windows (e.g., Gemini Apps).

### 2. Core Features

- **Recursive Traversal:** Walk the directory tree starting from the specified root.
- **Markdown Only:** Process only `.md` files; ignore all other file types.
- **Smart Filtering:**
  - Exclude by default: `.git`, `node_modules`, `vendor`, hidden directories.
  - Support `.lcbignore` file for custom exclusions.
- **Table of Contents:** Generate a navigable TOC at the top of each output file.
- **Document Separation:** Use HTML comments (`<!-- SOURCE: path -->`) and horizontal rules for clean separation between bundled documents.
- **Auto-Splitting:** Automatically split output into multiple parts if total size exceeds 100 MB.

### 3. Target Platform Constraints (Gemini Apps)

| Constraint | Limit |
| --- | --- |
| Max file size | 100 MB per file |
| File type | Single `.md` output |
| GitHub integration | Out of scope |
| ZIP output | Not needed (single-file approach) |

### 4. User Workflow

1. User navigates to their documentation/story directory: `cd ~/docs/my-story`
2. User runs the tool: `./lcb --output=context.md`
3. The tool scans for `.md` files, respecting `.lcbignore` rules.
4. Output file(s) created: `context.md` (or `context_part1.md`, `context_part2.md`, etc. if over 100 MB).
5. User uploads the bundled file to Gemini Apps.

### 5. Output Format

```markdown
# Bundled Context

## Table of Contents
- [intro.md](#intromd)
- [chapters/chapter-1.md](#chapterschapter-1md)
- [chapters/chapter-2.md](#chapterschapter-2md)

---

<!-- SOURCE: intro.md -->
# Introduction

[original content of intro.md]

---

<!-- SOURCE: chapters/chapter-1.md -->
# Chapter 1

[original content of chapter-1.md]

---

<!-- SOURCE: chapters/chapter-2.md -->
# Chapter 2

[original content of chapter-2.md]
```

### 6. Implementation Plan

1. **Directory Walking:** Use `filepath.WalkDir` to recursively find `.md` files.
2. **Ignore Logic:** Parse `.lcbignore` (gitignore-style patterns) and skip matching paths.
3. **Size Tracking:** Track cumulative output size; split into new part file when approaching 100 MB.
4. **TOC Generation:** Build TOC from collected file paths before writing content.
5. **Content Assembly:** Write TOC, then iterate files with `<!-- SOURCE: path -->` headers and `---` separators.

### 7. Technical Requirements

| Feature | Requirement |
| --- | --- |
| Language | Go 1.21+ |
| Output Type | Markdown (`.md`) |
| Max Output Size | 100 MB per file (auto-split if exceeded) |
| Input Files | `.md` only |

### 8. Out of Scope (v1)

- GitHub repository integration
- Non-Markdown file types
- ZIP file output
- Cross-platform binary distribution
- Concurrent file processing
