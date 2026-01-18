# jsl - JSON and JSONL Query Tool

A powerful command-line tool written in Go for querying, filtering, and manipulating JSON and JSONL (JSON Lines) files. Designed for developers who work with JSON in the terminal daily.

## Features

- 🔍 **Query**: Extract specific fields using dot-notation paths
- 🔎 **Filter**: Filter records with simple expression syntax (e.g., `age>28`)
- � **Field Selection**: Choose specific fields for output with `--select` (e.g., `-s name,age`)
- 📦 **Extract Mode**: Flatten nested results or iterate through collections with `--extract`
- 🃏 **Wildcard Keys**: Filter keys within objects using `*` or shell-safe `%` (e.g., `.sensors.%~=temp`)
- �🎨 **Format**: Pretty-print JSON/JSONL files
- 🔄 **Convert**: Convert between JSON and JSONL formats
- 📊 **Stats**: Display file statistics and schema information
- ✅ **Validate**: Validate JSON/JSONL file syntax
- 📥 **Stdin Support**: Pipe JSON directly without file paths
- ⚡ **Inline JSON**: Pass JSON strings directly as arguments

## Installation

### Using Go Install

```bash
go install github.com/bisegni/jsl@latest
```

### From Source

```bash
git clone https://github.com/bisegni/jsl.git
cd jsl
go build -o jsl
sudo mv jsl /usr/local/bin/  # Optional: install globally
```

To install the binary to a standard macOS/Linux path (e.g., /usr/local/bin), set GOBIN globally: `go env -w GOBIN=/usr/local/bin`. Then run the install command above.

## Quick Start

jsl supports three input methods to maximize productivity:

```bash
# 1. File path
jsl users.json .name

# 2. Stdin (pipe from other commands)
cat users.json | jsl .name
curl -s api.example.com/users | jsl .name

# 3. Inline JSON (for quick testing)
jsl '{"name":"Alice","age":30}' .name
```

## Usage

### Basic Syntax

```bash
jsl [command] [file|JSON|-] [path|expression] [flags]
```

If no command is provided, `jsl` defaults to querying the specified file or stdin.

### Core Functionality

#### 1. Query - Extract Fields

Extract specific fields from JSON/JSONL files using dot-notation paths.

```bash
# Basic field access
jsl users.json .name

# Nested field access
jsl company.json .location.city

# Array access by index
jsl company.json .employees.0.name

# Wildcard array access (extracts from all elements)
jsl company.json .employees.*.name
```

**Flags:**

- `-p, --path`: Path expression to extract (alternative to positional argument)
- `--pretty`: Pretty print output (default: true)
- `-e, --extract`: **Extract Mode** - Flatten object keys or array elements into a list of records
- `-s, --select`: **Field Selection** - Comma-separated list of fields to include in the output

#### 2. Advanced Querying & Wildcards

Filter keys within objects using wildcards and operators.

> [!TIP]
> Use the `%` character as a shell-safe wildcard to avoid the need for quotes. If you use `*`, you must wrap the path in quotes to prevent shell expansion.

```bash
# Shell-safe wildcard: Match keys containing "temp"
jsl sensors.jsonl .sensors.%~=temp

# Match all keys in an object
jsl sensors.jsonl .sensors.%

# Deep wildcard filtering with conditions
jsl sensors.jsonl '.sensors.*.metadata.room=living'
```

#### 3. Filter - Filter Records

Filter records based on field conditions.

**New concise expression syntax:**

```bash
# Filter by numeric comparison
jsl users.json 'age>28'

# Filter by string match (contains)
jsl users.json 'name~=Alice'

# Filter by exact match
jsl users.json 'status=active'
```

**Unified Syntax Operators:**

- `=`: Equal to
- `!=`: Not equal to
- `>`: Greater than
- `>=`: Greater than or equal to
- `<`: Less than
- `<=`: Less than or equal to
- `~=`: Contains (for strings)

#### 4. Field Selection & Extract Mode

Combine query paths with selection and extraction for powerful data manipulation.

```bash
# Selection: Extract specific fields from the results
jsl sensors.jsonl . --select timestamp,id

# Extraction: Flatten nested objects into individual records
jsl -e sensors.jsonl '.sensors.*' --select value,metadata
```

#### 5. Format - Pretty Print

Format and pretty-print JSON/JSONL files.

```bash
jsl format data.json
jsl format data.jsonl --output jsonl
```

#### 6. Convert - Format Conversion

Convert between JSON and JSONL formats.

```bash
# Convert JSON to JSONL
jsl convert users.json --to jsonl > users.jsonl

# Convert JSONL to JSON
jsl convert users.jsonl --to json > users.json
```

#### 7. Stats & Validate

```bash
# Show file statistics and schema info
jsl stats users.json

# Validate syntax
jsl validate users.json
```

## Path Expression Syntax

- `.field` - Access object field
- `.field.nested` - Access nested field
- `.array.*` - Wildcard to access all array elements
- `.array.0` - Access specific array element by index
- `.object.*` - Access all values in an object
- `.object.%~=pattern` - Wildcard key filtering (shell-safe)
- `.` - Return entire structure

## Examples

### Complex Pipeline Example

Filter active users, extract their names, and convert the output to JSONL:

```bash
jsl users.json 'active=true' | jsl .name | jsl convert --to jsonl
```

### Working with APIs

```bash
curl -s api.example.com/sensors | jsl '.data.*' --select id,value
```

## File Format Detection

jsl automatically detects file format based on extension:

- `.json` - Treated as JSON
- `.jsonl` / `.ldjson` - Treated as JSONL (JSON Lines)

For files without standard extensions, the tool attempts to parse as JSON first, then falls back to JSONL.

## Exit Codes

- `0` - Success
- `1` - Error (invalid file, parse error, etc.)

## Development

### Building

```bash
go build -o jsl
```

### Running Tests

```bash
go test ./...
```

### Project Structure

```
jsl/
├── main.go              # Entry point
├── cmd/                 # CLI commands
│   ├── root.go         # Root command
│   ├── query.go        # Query command
│   ├── filter.go       # Filter command
│   ├── format.go       # Format command
│   ├── convert.go      # Convert command
│   ├── stats.go        # Stats command
│   └── validate.go     # Validate command
└── pkg/
    ├── parser/         # JSON/JSONL parser
    │   └── parser.go
    └── query/          # Query and filter logic
        └── query.go
```

## License

See [LICENSE](LICENSE) file for details.

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## Author

Created as a tool for working with JSON and JSONL files in command-line environments.
