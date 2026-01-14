# jsl - JSON and JSONL Query Tool

A powerful command-line tool written in Go for querying, filtering, and manipulating JSON and JSONL (JSON Lines) files. Designed for easy integration into bash scripts and command-line workflows.

## Features

- 🔍 **Query**: Extract specific fields using dot-notation paths
- 🔎 **Filter**: Filter records based on field conditions
- 🎨 **Format**: Pretty-print JSON/JSONL files
- 🔄 **Convert**: Convert between JSON and JSONL formats
- 📊 **Stats**: Display file statistics and schema information
- ✅ **Validate**: Validate JSON/JSONL file syntax

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

## Usage

### Basic Commands

```bash
jsl [command] [file] [flags]
```

### Commands

#### 1. Query - Extract Fields

Extract specific fields from JSON/JSONL files using dot-notation paths.

```bash
# Extract a single field from all records
jsl query users.json --path name

# Extract nested fields
jsl query company.json --path location.city

# Extract array elements with wildcard
jsl query company.json --path employees.*.name

# Extract specific array element
jsl query company.json --path employees.0.salary

# Get all records (default)
jsl query users.json --path .
```

**Flags:**
- `-p, --path`: Path expression to extract (default: ".")
- `--pretty`: Pretty print output (default: true)

#### 2. Filter - Filter Records

Filter records based on field conditions.

```bash
# Filter by numeric comparison
jsl filter users.json --field age --op ">" --value 28

# Filter by equality
jsl filter users.json --field active --op "=" --value true

# Filter by string contains
jsl filter users.json --field name --op contains --value Alice

# Output as JSONL
jsl filter users.json --field age --op ">=" --value 30 --format jsonl
```

**Operators:**
- `=` or `==`: Equal to
- `!=`: Not equal to
- `>`: Greater than
- `>=`: Greater than or equal to
- `<`: Less than
- `<=`: Less than or equal to
- `contains`: String contains

**Flags:**
- `-f, --field`: Field path to filter on (required)
- `-o, --op`: Comparison operator (default: "=")
- `-v, --value`: Value to compare against (required)
- `--format`: Output format (json or jsonl, default: "json")
- `--pretty`: Pretty print output (default: true)

#### 3. Format - Pretty Print

Format and pretty-print JSON/JSONL files.

```bash
# Format JSON file
jsl format data.json

# Format JSONL file
jsl format data.jsonl

# Output as JSONL
jsl format data.json --output jsonl
```

**Flags:**
- `-p, --pretty`: Pretty print output (default: true)
- `-o, --output`: Output format (json or jsonl, auto-detect if not specified)

#### 4. Convert - Format Conversion

Convert between JSON and JSONL formats.

```bash
# Convert JSON to JSONL
jsl convert users.json --to jsonl > users.jsonl

# Convert JSONL to JSON
jsl convert users.jsonl --to json > users.json
```

**Flags:**
- `-t, --to`: Target format (json or jsonl, required)
- `--pretty`: Pretty print output (default: true)

#### 5. Stats - Show Statistics

Display statistics about JSON/JSONL files.

```bash
jsl stats users.json
```

**Output includes:**
- File format (JSON or JSONL)
- Total record count
- Field names and types
- Type distribution per field

#### 6. Validate - Syntax Validation

Validate JSON/JSONL file syntax.

```bash
jsl validate users.json
jsl validate users.jsonl
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

# Get all names
jsl query users.json --path name
# Output: ["Alice", "Bob", "Charlie"]

# Filter users over 28
jsl filter users.json --field age --op ">" --value 28

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

# Extract all employee names
jsl query company.json --path employees.*.name
# Output: ["John", "Jane"]

# Get company location
jsl query company.json --path location.city
# Output: "Austin"
```

### Chaining Commands

```bash
# Filter and convert in one pipeline
jsl filter users.json --field age --op ">" --value 25 | \
  jsl convert /dev/stdin --to jsonl > filtered_users.jsonl

# Extract names from filtered results
jsl filter users.json --field active --op "=" --value true | \
  jsl query /dev/stdin --path name
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
