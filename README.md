# jsl - JSON and JSONL Query Tool

A powerful command-line tool written in Go for querying, filtering, and manipulating JSON and JSONL (JSON Lines) files. Designed for developers who work with JSON in the terminal daily.

## Features

- 🗣️ **SQL-like Syntax**: Query using familiar syntax: `SELECT ... WHERE ...`
- 🎨 **Format**: Pretty-print JSON/JSONL files
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
jsl users.json "SELECT name"

# 2. Stdin (pipe from other commands)
cat users.json | jsl "SELECT name"
curl -s api.example.com/users | jsl "SELECT name"

# 3. Inline JSON (for quick testing)
jsl '{"name":"Alice","age":30}' "SELECT name"
```

## Usage

### Basic Syntax

```bash
jsl [command] [file|JSON|-] [path|expression] [flags]
```

If no command is provided, `jsl` defaults to querying the specified file or stdin.

### Core Functionality

#### 1. SQL-like Query Syntax

Perform queries using a familiar SQL-style syntax.

```bash
# Select specific fields
jsl users.json "SELECT name, age"

# Select with condition
jsl users.json "SELECT name, city WHERE age > 25"

# Select all fields with condition
jsl users.json "SELECT * WHERE active = true"
```

#### 2. Format - Pretty Print

Format and pretty-print JSON/JSONL files.

```bash
jsl format data.json
jsl format data.jsonl --output jsonl
```

#### 3. Convert - Format Conversion

Convert between JSON and JSONL formats.

```bash
# Convert JSON to JSONL
jsl convert users.json --to jsonl > users.jsonl

# Convert JSONL to JSON
jsl convert users.jsonl --to json > users.json
```

#### 4. Stats & Validate

```bash
# Show file statistics and schema info
jsl stats users.json

# Validate syntax
jsl validate users.json
```


## Examples

### Complex Pipeline Example

Filter active users, extract their names, and convert the output to JSONL:

```bash
jsl users.json "SELECT name WHERE active=true" | jsl convert --to jsonl
```

### Working with APIs

```bash
curl -s api.example.com/sensors | jsl "SELECT id, value"
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
