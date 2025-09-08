# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Goal

Create a command line tool that can verify JSON files in Minecraft datapacks as being valid against mcdoc schemas. The project aims to enable CI pipeline integration and LLM-assisted datapack development verification.

## Project Structure

### Root Directory Structure
- `java/` - Contains mcdoc schema definitions for Minecraft data structures
- `Spyglass/` - TypeScript monorepo with various packages for datapack validation (THIRD PARTY - READ ONLY)
- `spyglass.json` - Environment configuration file

### Key Spyglass Packages (Reference Only)
- `packages/core/` - Core parsing, AST nodes, and service infrastructure
- `packages/mcdoc/` - Parser and processor for mcdoc schema language
- `packages/mcdoc-cli/` - Existing CLI tool for mcdoc operations (reference implementation)
- `packages/java-edition/` - Java Edition specific mcdoc attributes
- `packages/json/` - JSON parsing and validation utilities
- `packages/nbt/` - NBT format support with mcdoc attributes
- `packages/language-server/` - Language server implementation
- `packages/vscode-extension/` - VS Code extension

### Schema Structure (java/ directory)
Contains hierarchical mcdoc schemas organized by:
- `assets/` - Client-side assets (models, textures, sounds, etc.)
- `data/` - Server-side data (advancements, recipes, loot tables, etc.)
- `server/` - Server configuration schemas
- `world/` - World data structures (blocks, entities, items)
- `util/` - Common utility schemas

## Build System & Commands

### Root Level (Spyglass/)
- `npm run build` - Build all packages using wireit
- `npm run build:dev` - Development build with watch mode
- `npm run watch` - Alias for build:dev --watch
- `npm run clean` - Clean TypeScript build artifacts
- `npm run clean:full` - Full clean including all generated files
- `npm run test` - Run tests using nyc/mocha
- `npm run test:local` - Local test configuration
- `npm run lint` - ESLint with max 0 warnings
- `npm run lint:fix` - Auto-fix linting issues
- `npm run fmt` - Format code using dprint
- `npm run fmt:check` - Check formatting without fixing

### Development Environment
- Node.js: ^18.18.0 || >=20.0.0
- Uses workspaces for monorepo management
- TypeScript compilation with project references
- Uses wireit for build orchestration

### Key Packages for CLI Tool Development
1. **@spyglassmc/core** - Essential parsing and validation infrastructure
2. **@spyglassmc/mcdoc** - mcdoc language parser and processor
3. **@spyglassmc/mcdoc-cli** - Existing CLI foundation (reference only)
4. **@spyglassmc/java-edition** - Java Edition specific validation
5. **@spyglassmc/json** - JSON format handling

### Architecture Notes
- Uses TypeScript project references for build dependency management
- Monorepo structure with shared build configuration
- ESLint enforces consistent imports and coding standards
- Test framework: Mocha with NYC coverage
- Source maps enabled for debugging

### Development Workflow
1. Use `npm run build:dev` for incremental development
2. Run tests with `npm run test:local` during development
3. Lint code with `npm run lint:fix` before commits
4. Format code with `npm run fmt`

## Implementation Strategy

**IMPORTANT**: The JSON validator CLI must be implemented OUTSIDE the Spyglass directory. Spyglass is third-party code used as a reference and dependency source only. DO NOT modify any files in the Spyglass directory.

The existing `mcdoc-cli` package provides a foundation but focuses on schema generation. For datapack validation:

1. **Core Dependencies**: Use published npm packages from @spyglassmc/* for parsing infrastructure
2. **Schema Loading**: Use @spyglassmc/mcdoc to parse schema files from java/
3. **Validation Logic**: Integrate JSON validation against loaded schemas  
4. **CLI Interface**: Create new standalone CLI tool with options for:
   - Minecraft version selection
   - Schema path configuration
   - Datapack file validation
   - Output formats (JSON, text, etc.)

### Key Files to Examine (Reference Only)
- `Spyglass/packages/mcdoc-cli/src/index.ts` - Existing CLI implementation patterns
- `Spyglass/packages/core/src/service/` - Core services architecture
- `Spyglass/packages/mcdoc/src/` - mcdoc parser implementation

### CLI Tool Location
Create the new JSON validator CLI tool in the root directory (alongside java/ and Spyglass/), consuming the Spyglass packages as external dependencies via npm.