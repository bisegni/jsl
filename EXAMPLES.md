# JSL Usage Examples

This file demonstrates common usage patterns and examples for the jsl command-line tool.

## SQL-like Query Examples
 
 ### Select specific fields
 ```bash
 jsl examples/users.json "SELECT name, age"
 # Output: [{"name":"Alice", "age":30}, ...]
 ```
 
 ### Select with WHERE clause
 ```bash
 jsl examples/users.json "SELECT name, city WHERE age > 25"
 ```
 
 ### Select All with Condition
 ```bash
 jsl examples/users.json "SELECT * WHERE active = true"
 ```
 

 ### Working with Sensor Data
 
 The `examples/sensors.jsonl` file contains an array of sensor readings for each timestamp.
 
 ```bash
 # Select all sensor names from the array
 jsl examples/sensors.jsonl "SELECT sensors.*.name"
 
 # Filter ROWS where ANY sensor is of type 'temp' (returns the whole record if match is found)
 jsl examples/sensors.jsonl "SELECT * WHERE sensors.*.type = 'temp'"
 
 # Filter ARRAY ELEMENTS: Select only the sensors of type 'temp'
 # Syntax: array.*.key=value returns a list of matching sub-objects
 jsl examples/sensors.jsonl "SELECT sensors.*.type='temp'"
 
 # Filter and Extract: Get only the NAMES of 'temp' sensors
 # Syntax: array.*.filter.field
 jsl examples/sensors.jsonl "SELECT sensors.*.type='temp'.name"
 
 # Complex Filtering: Select rows where a specific room has sensors
 jsl examples/sensors.jsonl "SELECT * WHERE sensors.*.room = 'kitchen'"
 ```

 ### Select All (Wildcard)
 ```bash
 jsl examples/users.json "SELECT *"
 ```
 
 ## Filter Examples (using WHERE)
 
 ### Numeric comparisons
 ```bash
 # Greater than
 jsl examples/users.json "SELECT * WHERE age > 28"
 
 # Equal to
 jsl examples/users.json "SELECT * WHERE id = 2"
 ```
 
 ### String operations
 ```bash
 # Contains substring
 jsl examples/users.json "SELECT * WHERE name ~= 'li'"
 
 # Exact match
 jsl examples/users.json "SELECT * WHERE city = 'Boston'"
 ```
 
 ## Stdin Examples
 
 ### Reading from pipes
 
 ```bash
 # Query from stdin
 cat examples/users.json | jsl "SELECT name"
 
 # Filter from stdin
 cat examples/users.json | jsl "SELECT * WHERE age > 25"
 ```
 
 ### Working with curl and APIs
 
 ```bash
 # Query API results
 curl -s https://api.example.com/users | jsl "SELECT email"
 
 # Filter API results
 curl -s https://api.example.com/users | jsl "SELECT * WHERE status = 'active'"
 ```
 
 ## Inline JSON Examples
 
 ### Quick testing without files
 
 ```bash
 # Simple object
 jsl '{"name":"Alice","age":30}' "SELECT name"
 # Output: "Alice"
 
 # Array of objects
 jsl '[{"id":1,"name":"Alice"},{"id":2,"name":"Bob"}]' "SELECT name"
 # Output: ["Alice", "Bob"]
 
 # Filtering inline JSON
 jsl '[{"age":25},{"age":30},{"age":35}]' "SELECT * WHERE age >= 30"
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

## Pipeline Examples
 
 ### Filter and then query
 
 ```bash
 jsl examples/users.json "SELECT * WHERE age>25" | jsl "SELECT name"
 ```
 
 ### Query and then convert format
 
 ```bash
 jsl examples/users.json "SELECT *" | jsl convert --to jsonl > /tmp/all_users.jsonl
 ```
 
 ### Complex pipeline
 
 ```bash
 # Filter active users, extract names, and save as JSONL
 jsl examples/users.json "SELECT name WHERE active=true" | jsl convert --to jsonl > /tmp/active_names.jsonl
 ```
 
 ### Chaining multiple filters
 
 ```bash
 # Filter by age, then by city (conceptually)
 # Note: You can do this in one query: SELECT * WHERE age>25 AND city~='New' (if AND is supported, otherwise chain)
 jsl examples/users.json "SELECT * WHERE age>25" | jsl "SELECT * WHERE city ~= 'New'"
 ```
 
 ## Working with Standard Input
 
 All commands now support reading from stdin automatically:
 
 ```bash
 # Query from stdin
 cat examples/users.json | jsl "SELECT name"
 
 # Filter from stdin
 curl -s https://api.example.com/users | jsl "SELECT * WHERE age>25"
 ```
 
 ## Integration with Other Tools
 
 ### With jq
 
 ```bash
 # Use jsl for initial filtering, then jq for complex transformations
 jsl examples/users.json "SELECT * WHERE age>25" | jq '.[] | {name, age}'
 ```
 
 ### With grep
 
 ```bash
 # Extract emails and search for domain
 jsl examples/users.json "SELECT email" | jq -r '.[]' | grep "example.com"
 ```
 
 ### With awk
 
 ```bash
 # Get names and format with awk
 jsl examples/users.json "SELECT name" | jq -r '.[]' | awk '{print "User: " $0}'
 ```
 
 ## Advanced Patterns
 
 ### Extract multiple fields (using jq)
 
 ```bash
 jsl examples/users.json "SELECT *" | jq '.[] | {name, age}'
 ```
 
 ### Count filtered records
 
 ```bash
 jsl examples/users.json "SELECT * WHERE active=true" | jq 'length'
 ```
 
 ### Sort by field (using jq)
 
 ```bash
 jsl examples/users.json "SELECT *" | jq 'sort_by(.age)'
 ```
 
 ### Group by field (using jq)
 
 ```bash
 jsl examples/users.json "SELECT *" | jq 'group_by(.city)'
 ```
 
 ## Performance Tips
 
 1. **Use JSONL for large files**: JSONL files can be processed line-by-line, which is more memory-efficient for large datasets.
 
 2. **Filter early**: Use `WHERE` clauses to reduce the dataset size before piping to other tools.
 
 3. **Use --pretty=false for scripts**: Disable pretty printing in automated scripts to reduce output size.
 
 4. **Stream processing**: For very large files, consider processing in chunks.
 
 ## Common Patterns
 
 ### Extract unique values
 
 ```bash
 jsl examples/users.json "SELECT city" | jq -r '.[]' | sort -u
 ```
 
 ### Count records by field
 
 ```bash
 jsl examples/users.json "SELECT city" | jq -r '.[]' | sort | uniq -c
 ```
 
 ### Find min/max values
 
 ```bash
 # Maximum age
 jsl examples/users.json "SELECT age" | jq 'max'
 
 # Minimum age
 jsl examples/users.json "SELECT age" | jq 'min'
 ```
 
 ### Average calculation
 
 ```bash
 jsl examples/users.json "SELECT age" | jq 'add/length'
 ```
 
 ## Error Handling
 
 ### Check if file is valid before processing
 
 ```bash
 if jsl validate data.json 2>/dev/null; then
   jsl data.json "SELECT name"
 else
   echo "Invalid JSON file"
 fi
 ```
 
 ### Handle missing fields gracefully
 
 ```bash
 # The tool skips records where the path doesn't exist
 jsl examples/users.json "SELECT optional_field"
 ```
 
 ## Shell Integration
 
 ### Bash function wrapper
 
 ```bash
 # Add to ~/.bashrc
 jslq() {
   jsl "$1" "SELECT $2"
 }
 
 # Usage
 jslq examples/users.json "name"
 ```
 
 ### Alias for common operations
 
 ```bash
 alias jsl-names='jsl "SELECT name"'
 alias jsl-validate='jsl validate'
 
 # Usage
 jsl-names examples/users.json
 ```
