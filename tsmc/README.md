# tsmc - TypeScript Minecraft Datapack Validator

A command-line tool for validating Minecraft datapack JSON files against mcdoc schemas using the Spyglass parser.

## Installation

```bash
npm install
npm run build
```

## Usage

The `tsmc` tool provides commands for working with mcdoc schema files and validating Minecraft datapacks.

### Global Options

- `--schema-path <path>` - Path to schema directory (default: `./java`)

### Commands

#### `info` - Schema Information

Shows information about the schema directory and lists all discovered mcdoc files.

```bash
# Use default schema path (./java)
tsmc info

# Use custom schema path
tsmc --schema-path ./custom/schemas info
```

**Example output:**
```
📍 Schema root: /path/to/schemas
📄 Found 219 mcdoc files

📝 Files:
  pack.mcdoc
  util/text.mcdoc
  data/advancement/mod.mcdoc
  ...
```

#### `parse-schemas` - Parse and Validate Schemas

Parse and validate all mcdoc schema files using the Spyglass parser.

**Options:**
- `-v, --verbose` - Enable verbose output with detailed error information

```bash
# Parse with default schema path
tsmc parse-schemas

# Parse with verbose output
tsmc parse-schemas --verbose

# Parse with custom schema path
tsmc --schema-path ./custom/schemas parse-schemas
```

**Example output:**
```
🔍 Starting mcdoc schema parsing...
Schema directory: ./java

📁 Loading mcdoc files...
Found 219 mcdoc files in /path/to/schemas
Loaded: pack.mcdoc
Loaded: util/text.mcdoc
...

⚡ Parsing mcdoc files...
Parsing: pack.mcdoc
Initializing Spyglass project...
Spyglass project initialized.
...

📊 Parsing Report:
  Total files: 219
  Successful: 219
  With errors: 0
  Total errors: 0

✅ All schema files parsed successfully!
```

### Help

Get help for any command:

```bash
tsmc --help
tsmc parse-schemas --help
```

## Project Structure

```
tsmc/
├── src/                 # TypeScript source code
│   ├── index.ts        # CLI entry point
│   ├── mcdoc-loader.ts # Schema file discovery and loading
│   └── mcdoc-parser.ts # Spyglass integration for parsing
├── dist/               # Compiled JavaScript output
├── java/               # Default mcdoc schema directory
├── Spyglass/           # Third-party Spyglass codebase (read-only)
└── package.json
```

## Error Handling

The tool provides comprehensive error handling:

- **Invalid schema paths**: Detected and reported with clear error messages
- **Non-existent directories**: Tool exits with descriptive error
- **Parsing errors**: Collected and reported with file location information
- **Verbose mode**: Use `--verbose` to see detailed error information

## Examples

### Basic Schema Validation

```bash
# Validate all schemas in the default ./java directory
tsmc parse-schemas
```

### Custom Schema Directory

```bash
# Use a different schema directory
tsmc --schema-path /path/to/minecraft-schemas parse-schemas
```

### Troubleshooting with Verbose Output

```bash
# Get detailed error information
tsmc parse-schemas --verbose
```

### Check Schema Directory Contents

```bash
# List all schema files in a directory
tsmc --schema-path ./custom/schemas info
```

## Dependencies

The tool uses the following key dependencies:

- **@spyglassmc/core** - Core Spyglass parsing infrastructure
- **@spyglassmc/mcdoc** - mcdoc language parser
- **@spyglassmc/java-edition** - Java Edition specific validation
- **commander** - CLI framework
- **glob** - File pattern matching

## Development

This tool is built as a standalone CLI that consumes the Spyglass packages as external dependencies. The Spyglass directory contains the third-party codebase for reference but should not be modified.

To build:
```bash
npm run build
```

To run in development:
```bash
npm run dev  # Starts TypeScript compiler in watch mode
```