# JSL Usage Examples

This file demonstrates common usage patterns and examples for the jsl command-line tool.

## Quick Start Examples

### Most Common Patterns (New Concise Syntax)

```bash
# Query fields from a file
jsl examples/users.json .name

# Filter with expressions
jsl filter examples/users.json 'age>28'

# Query from stdin
cat examples/users.json | jsl .name

# Filter from stdin
cat examples/users.json | jsl filter 'age>=30'

# Inline JSON testing
jsl '{"name":"Alice","age":30}' .name
```

## Basic Query Examples

### Query all records
```bash
# New concise syntax
jsl examples/users.json .

# Traditional syntax
jsl query examples/users.json --path .
```

### Extract a single field from all records
```bash
# New concise syntax (recommended)
jsl examples/users.json .name

# Traditional syntax
jsl query examples/users.json --path name
# Output: ["Alice", "Bob", "Charlie", "Diana"]
```

### Extract nested fields
```bash
# New concise syntax
jsl examples/company.json .location.city

# Traditional syntax
jsl query examples/company.json --path location.city
# Output: "Austin"
```

### Extract from arrays with wildcard
```bash
# New concise syntax
jsl examples/company.json .employees.*.name

# Traditional syntax
jsl query examples/company.json --path employees.*.name
# Output: ["John", "Jane", "Mike"]
```

### Extract specific array element
```bash
jsl examples/company.json .employees.0.salary
# Output: 80000
```

## Filter Examples

### Numeric comparisons with new expression syntax
```bash
# Greater than (note: use quotes to protect shell special characters)
jsl filter examples/users.json 'age>28'

# Greater than or equal
jsl filter examples/users.json 'age>=30'

# Less than or equal
jsl filter examples/users.json 'age<=30'

# Equal to
jsl filter examples/users.json 'id=2'

# Not equal
jsl filter examples/users.json 'age!=25'
```

### Traditional flag syntax (still supported)
```bash
jsl filter examples/users.json --field age --op ">" --value 28
jsl filter examples/users.json --field age --op "<=" --value 30
```

### Boolean filters
```bash
jsl filter examples/users.json 'active=true'
```

### String operations
```bash
# Contains substring (using ~= operator)
jsl filter examples/users.json 'name~=li'

# Exact match
jsl filter examples/users.json 'city=Boston'
```

### Output as JSONL
```bash
jsl filter examples/users.json 'age>25' --format jsonl
```

## Stdin Examples

### Reading from pipes
```bash
# Query from stdin
cat examples/users.json | jsl .name

# Filter from stdin
cat examples/users.json | jsl filter 'age>25'

# Format from stdin
cat examples/users.json | jsl format

# Stats from stdin
cat examples/users.json | jsl stats

# Validate from stdin
cat examples/users.json | jsl validate
```

### Working with curl and APIs
```bash
# Query API results
curl -s https://api.example.com/users | jsl .data.*.email

# Filter API results
curl -s https://api.example.com/users | jsl filter 'status=active'
```

## Inline JSON Examples

### Quick testing without files
```bash
# Simple object
jsl '{"name":"Alice","age":30}' .name
# Output: "Alice"

# Array of objects
jsl '[{"id":1,"name":"Alice"},{"id":2,"name":"Bob"}]' .*.name
# Output: ["Alice", "Bob"]

# Nested structures
jsl '{"user":{"profile":{"name":"Alice"}}}' .user.profile.name
# Output: "Alice"

# Filtering inline JSON
jsl filter '[{"age":25},{"age":30},{"age":35}]' 'age>=30'
```

## Format Examples

### Pretty-print JSON
```bash
# From file
jsl format examples/users.json

# From stdin
cat examples/users.json | jsl format

# Inline
echo '{"name":"Alice"}' | jsl format
```

### Pretty-print JSONL
```bash
jsl format examples/users.jsonl
```

### Change output format
```bash
# Convert to JSONL on the fly
jsl format examples/users.json --output jsonl

# From stdin
cat examples/users.json | jsl format --output jsonl
```

## Convert Examples

### JSON to JSONL
```bash
# From file
jsl convert examples/users.json --to jsonl > /tmp/users.jsonl

# From stdin
cat examples/users.json | jsl convert --to jsonl > /tmp/users.jsonl
```

### JSONL to JSON
```bash
jsl convert examples/users.jsonl --to json > /tmp/users.json
```

### Convert without pretty printing
```bash
jsl convert examples/users.json --to jsonl --pretty=false > /tmp/compact.jsonl
```

## Statistics Examples

### Show file statistics
```bash
# From file
jsl stats examples/users.json

# From stdin
cat examples/users.json | jsl stats

# Output:
# File: examples/users.json
# Format: JSON
# Total records: 4
# 
# Fields:
#   age:
#     number: 4 (100.0%)
#   name:
#     string: 4 (100.0%)
#   ...
```

## Validation Examples

### Validate JSON file
```bash
# From file
jsl validate examples/users.json
# Output: ✅ Valid JSON file with 4 record(s)

# From stdin
cat examples/users.json | jsl validate
echo '{"valid":"json"}' | jsl validate
```

### Validate JSONL file
```bash
jsl validate examples/users.jsonl
# Output: ✅ Valid JSONL file with 4 record(s)
```

### Validate invalid file
```bash
echo '{"invalid": json}' > /tmp/bad.json
jsl validate /tmp/bad.json
# Output: ❌ Validation failed: ...
```

## Pipeline Examples (New Concise Syntax)

### Filter and then query
```bash
# New concise syntax (recommended)
jsl filter examples/users.json 'age>25' | jsl .name

# Traditional syntax
jsl filter examples/users.json --field age --op ">" --value 25 | jsl query - .name
```

### Query and then convert format
```bash
# New concise syntax
jsl examples/users.json . | jsl convert --to jsonl > /tmp/all_users.jsonl

# Traditional syntax
jsl query examples/users.json --path . | jsl convert - --to jsonl > /tmp/all_users.jsonl
```

### Complex pipeline
```bash
# Filter active users, extract names, and save as JSONL (new syntax)
jsl filter examples/users.json 'active=true' | jsl .name | jsl convert --to jsonl > /tmp/active_names.jsonl

# Traditional syntax
jsl filter examples/users.json --field active --op "=" --value true | \
  jsl query - --path name | \
  jsl convert - --to jsonl > /tmp/active_names.jsonl
```

### Chaining multiple filters
```bash
# Filter by age, then by city
jsl filter examples/users.json 'age>25' | jsl filter 'city~=New'
```

## Working with Standard Input

All commands now support reading from stdin automatically (no need for explicit filenames):

```bash
# Query from stdin
cat examples/users.json | jsl .name

curl https://api.example.com/users | jsl filter /dev/stdin --field age --op ">" --value 25
```

## Integration with Other Tools

### With jq
```bash
# Use jsl for initial filtering, then jq for complex transformations
jsl filter examples/users.json --field age --op ">" --value 25 | \
  jq '.[] | {name, age}'
```

### With grep
```bash
# Extract emails and search for domain
jsl query examples/users.json --path email | jq -r '.[]' | grep "example.com"
```

### With awk
```bash
# Get names and format with awk
jsl query examples/users.json --path name | jq -r '.[]' | awk '{print "User: " $0}'
```

## Advanced Patterns

### Extract multiple fields (using jq)
```bash
jsl query examples/users.json --path . | jq '.[] | {name, age}'
```

### Count filtered records
```bash
jsl filter examples/users.json --field active --op "=" --value true | jq 'length'
```

### Sort by field (using jq)
```bash
jsl query examples/users.json --path . | jq 'sort_by(.age)'
```

### Group by field (using jq)
```bash
jsl query examples/users.json --path . | jq 'group_by(.city)'
```

## Performance Tips

1. **Use JSONL for large files**: JSONL files can be processed line-by-line, which is more memory-efficient for large datasets.

2. **Filter first, then query**: When working with large files, filter records first to reduce the dataset size before extracting fields.

3. **Use --pretty=false for scripts**: Disable pretty printing in automated scripts to reduce output size.

4. **Stream processing**: For very large files, consider processing in chunks or using line-by-line tools.

## Common Patterns

### Extract unique values
```bash
jsl query examples/users.json --path city | jq -r '.[]' | sort -u
```

### Count records by field
```bash
jsl query examples/users.json --path city | jq -r '.[]' | sort | uniq -c
```

### Find min/max values
```bash
# Maximum age
jsl query examples/users.json --path age | jq 'max'

# Minimum age
jsl query examples/users.json --path age | jq 'min'
```

### Average calculation
```bash
jsl query examples/users.json --path age | jq 'add/length'
```

## Error Handling

### Check if file is valid before processing
```bash
if jsl validate data.json 2>/dev/null; then
  jsl query data.json --path name
else
  echo "Invalid JSON file"
fi
```

### Handle missing fields gracefully
```bash
# The query command skips records where the path doesn't exist
jsl query examples/users.json --path optional_field
```

## Shell Integration

### Bash function wrapper
```bash
# Add to ~/.bashrc
jslq() {
  jsl query "$1" --path "$2"
}

# Usage
jslq examples/users.json name
```

### Alias for common operations
```bash
alias jsl-names='jsl query --path name'
alias jsl-validate='jsl validate'

# Usage
jsl-names examples/users.json
```
