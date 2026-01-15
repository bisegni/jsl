# jsl - JSON and JSONL Query Tool

A powerful command-line tool written in Go for querying, filtering, and manipulating JSON and JSONL (JSON Lines) files. Designed for developers who work with JSON in the terminal daily.

## Features

- 🔍 **Query**: Extract specific fields using dot-notation paths
- 🔎 **Filter**: Filter records with simple expression syntax (e.g., `age>28`)
- 🎨 **Format**: Pretty-print JSON/JSONL files
- 🔄 **Convert**: Convert between JSON and JSONL formats
- 📊 **Stats**: Display file statistics and schema information
- ✅ **Validate**: Validate JSON/JSONL file syntax
- 📥 **Stdin Support**: Pipe JSON directly without file paths
- ⚡ **Inline JSON**: Pass JSON strings directly as arguments

## Installation

### From Source

```bash
git clone https://github.com/bisegni/jsl.git
cd jsl
go build -o jsl
sudo mv jsl /usr/local/bin/  # Optional: install globally
```

### Using Go Install

```bash
go install github.com/bisegni/jsl@latest
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

### Basic Commands

```bash
jsl [command] [file|JSON|-] [arguments]
```

### Commands

#### 1. Query - Extract Fields

Extract specific fields from JSON/JSONL files using dot-notation paths.

**New concise syntax:**
```bash
# From file
jsl users.json .name
jsl company.json .location.city

# From stdin
cat users.json | jsl .name
echo '{"name":"Alice"}' | jsl .name

# Inline JSON
jsl '{"user":{"name":"Alice"}}' .user.name

# Using the query subcommand (alternative)
jsl query users.json .name
```

**Traditional syntax (still supported):**
```bash
jsl query users.json --path name
jsl query company.json --path location.city
```

**Examples:**
```bash
# Extract nested fields
jsl company.json .employees.*.name

# Extract specific array element
jsl company.json .employees.0.salary

# Get all records (default)
jsl users.json .
```

**Flags:**
- `-p, --path`: Path expression to extract (default: ".")
- `--pretty`: Pretty print output (default: true)

#### 2. Filter - Filter Records

Filter records based on field conditions with a simple expression syntax.

**New concise expression syntax:**
```bash
# From file
jsl filter users.json 'age>28'
jsl filter users.json 'name~=Alice'  # contains
jsl filter users.json 'status=active'

# From stdin
cat users.json | jsl filter 'age>=30'
```

**Traditional flag syntax (still supported):**
```bash
jsl filter users.json --field age --op ">" --value 28
jsl filter users.json --field name --op contains --value Alice
```

**Expression Operators:**
- `=`: Equal to
- `!=`: Not equal to
- `>`: Greater than
- `>=`: Greater than or equal to
- `<`: Less than
- `<=`: Less than or equal to
- `~=`: Contains (for strings)

**Note:** Shell special characters like `>`, `<` need to be quoted: `'age>28'`

**Flags:**
- `-f, --field`: Field path to filter on
- `-o, --op`: Comparison operator (default: "=")
- `-v, --value`: Value to compare against
- `--format`: Output format (json or jsonl, default: "json")
- `--pretty`: Pretty print output (default: true)

#### 3. Format - Pretty Print

Format and pretty-print JSON/JSONL files.

```bash
# Format JSON file
jsl format data.json

# Format from stdin
cat data.json | jsl format
echo '{"name":"Alice"}' | jsl format

# Format JSONL file
jsl format data.jsonl

# Output as JSONL
jsl format data.json --output jsonl
```

**Flags:**
- `--pretty`: Pretty print output (default: true)
- `-o, --output`: Output format (json or jsonl, auto-detect if not specified)

#### 4. Convert - Format Conversion

Convert between JSON and JSONL formats.

```bash
# Convert JSON to JSONL
jsl convert users.json --to jsonl > users.jsonl

# Convert JSONL to JSON
jsl convert users.jsonl --to json > users.json

# From stdin
cat users.json | jsl convert --to jsonl
```

**Flags:**
- `-t, --to`: Target format (json or jsonl, required)
- `--pretty`: Pretty print output (default: true)

#### 5. Stats - Show Statistics

Display statistics about JSON/JSONL files.

```bash
# From file
jsl stats users.json

# From stdin
cat users.json | jsl stats
```

**Output includes:**
- File format (JSON or JSONL)
- Total record count
- Field names and types
- Type distribution per field

#### 6. Validate - Syntax Validation

Validate JSON/JSONL file syntax.

```bash
# From file
jsl validate users.json
jsl validate users.jsonl

# From stdin
cat users.json | jsl validate
```

## Examples

### Working with User Data

```bash
# Sample users.json
[
  {"id": 1, "name": "Alice", "age": 30, "city": "New York"},
  {"id": 2, "name": "Bob", "age": 25, "city": "San Francisco"},
  {"id": 3, "name": "Charlie", "age": 35, "city": "Boston"}
]

# Get all names (new concise syntax)
jsl users.json .name
# Output: ["Alice", "Bob", "Charlie"]

# Filter users over 28 (new concise syntax)
jsl filter users.json 'age>28'

# Get statistics
jsl stats users.json

# Convert to JSONL
jsl convert users.json --to jsonl > users.jsonl
```

### Working with Nested Data

```bash
# Sample company.json
{
  "company": "TechCorp",
  "employees": [
    {"name": "John", "role": "Engineer", "salary": 80000},
    {"name": "Jane", "role": "Manager", "salary": 95000}
  ],
  "location": {"city": "Austin", "state": "TX"}
}

# Extract all employee names (new concise syntax)
jsl company.json .employees.*.name
# Output: ["John", "Jane"]

# Get company location (new concise syntax)
jsl company.json .location.city
# Output: "Austin"
```

### Chaining Commands with Stdin

```bash
# Filter and convert in one pipeline (new concise syntax)
jsl filter users.json 'age>25' | jsl convert --to jsonl > filtered_users.jsonl

# Extract names from filtered results (new concise syntax)
jsl filter users.json 'active=true' | jsl .name

# Quick inline JSON testing
echo '{"users":[{"name":"Alice","age":30}]}' | jsl .users.*.name

# Fetch from API and query
curl -s api.example.com/users | jsl .data.*.email
```

## Input Methods

jsl supports three flexible input methods:

### 1. File Paths
```bash
jsl users.json .name
jsl filter data.json 'age>28'
```

### 2. Standard Input (stdin)
```bash
# Pipe from other commands
cat users.json | jsl .name
curl -s api.example.com/data | jsl .items.*.id

# Explicit stdin marker (optional)
cat users.json | jsl - .name
```

### 3. Inline JSON
```bash
# Quick testing without files
jsl '{"name":"Alice","age":30}' .name
jsl '[{"id":1},{"id":2}]' .*.id
```

## Path Expression Syntax

Path expressions use dot notation to navigate JSON structures:

- `.field` - Access object field
- `.field.nested` - Access nested field
- `.array.*` - Wildcard to access all array elements
- `.array.0` - Access specific array element by index
- `.` - Return entire structure

## File Format Detection

jsl automatically detects file format based on extension:
- `.json` - Treated as JSON
- `.jsonl` - Treated as JSONL (JSON Lines)

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
