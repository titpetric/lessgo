# lessgo CLI

The `lessgo` command-line tool provides formatting and compilation utilities for LESS files.

## Building

```bash
go build -o bin/lessgo ./cmd/lessgo
```

Or use the task:
```bash
task build
```

## Commands

### fmt - Format LESS Files

Formats LESS source files by:
- Adding missing semicolons after declarations
- Fixing indentation (2 spaces per level)
- Preserving structure and comments

**Usage:**
```bash
./bin/lessgo fmt <file_pattern>...
```

**Examples:**
```bash
# Format a single file
./bin/lessgo fmt styles.less

# Format all LESS files in a directory
./bin/lessgo fmt "*.less"

# Format specific files
./bin/lessgo fmt variables.less mixins.less theme.less
```

**Features:**
- Glob pattern support for multiple files
- Graceful handling of missing semicolons
- Detects new property declarations even without semicolons
- 2-space indentation (configurable in code)

**Limitations:**
- Does not evaluate variables in output (preserves `@variable` as-is)
- Nested rules may not maintain perfect indentation in complex cases
- See TODO.md for planned improvements

## Example

**Input (messy.less):**
```less
@primary: #3498db;

.card {
  background: white;
  border: 1px solid gray
  padding: 20px
}
```

**After `lessgo fmt messy.less`:**
```less
@primary: #3498db;

.card {
  background: white;
  border: 1px solid gray;
  padding: 20px;
}
```

## Notes

- The formatter is conservative - it won't remove comments or reformat code
- Missing semicolons are detected by looking for IDENT followed by COLON (new property pattern)
- The formatter rewrites files in-place; make backups if needed
