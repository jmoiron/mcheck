# TODO - TSMC JSON Datapack Validator

This document outlines the next steps for improving and extending the TSMC tool after the successful implementation of basic JSON validation against mcdoc schemas.

## ✅ Current Status

**MAJOR SUCCESS**: The core validation system is working! 
- 844 JSON files processed
- 601 files (71%) fully validated successfully
- All biome files (64/64) validate perfectly
- All configured features (194/194) validate perfectly
- All placed features (231/231) validate perfectly
- All structures (33/33) validate perfectly

## 🎯 Next Steps

### Phase 1: Improve Schema Coverage (High Priority)

#### 1.1 Add Missing Registry Type Mappings
- **Problem**: 243 files fail validation due to missing schema mappings for nested registry types
- **Files to update**: `src/path-mapper.ts`
- **Action Items**:
  ```typescript
  // Add these mappings to getDispatchKey() and getExpectedTypeName():
  'worldgen/noise': 'minecraft:resource["worldgen/noise"]',
  'worldgen/multi_noise_biome_source_parameter_list': 'minecraft:resource["worldgen/multi_noise_biome_source_parameter_list"]',
  'worldgen/flat_level_generator_preset': 'minecraft:resource["worldgen/flat_level_generator_preset"]',
  // Add deeply nested template pool mappings...
  ```

#### 1.2 Handle Nested Directory Structures
- **Problem**: Complex nested paths like `bastion/treasure/walls` not properly mapped
- **Solution**: Improve path parsing logic to handle arbitrary nesting levels
- **Files to update**: `src/path-mapper.ts` - `parseDatapackPath()` method

#### 1.3 Research Missing Schema Types
- **Investigation needed**: Examine `java/` schemas to find all available dispatch keys
- **Command to help**: `grep -r "dispatch.*minecraft:resource" java/`
- **Create comprehensive mapping**: Document all available schema types

### Phase 2: Enhanced Validation Logic (Medium Priority)

#### 2.1 Implement Full mcdoc-based Validation
- **Current state**: Using basic structural validation for biomes only
- **Goal**: Use parsed mcdoc AST for complete schema validation
- **Files to update**: `src/json-validator.ts` - `performBasicValidation()` method
- **Research needed**: Study Spyglass validation internals

#### 2.2 Add Version-Specific Validation
- **Enhancement**: Support different Minecraft versions (1.18, 1.19, 1.20, etc.)
- **Implementation**: 
  - Add `--mc-version` CLI option
  - Filter mcdoc schema fields based on version annotations
  - Update validation logic to respect version constraints

#### 2.3 Improve Error Messages
- **Current**: Generic error messages
- **Goal**: Specific, actionable error messages with suggested fixes
- **Examples**: 
  - "Field 'temperature' must be between -2.0 and 2.0, got 3.5"
  - "Required field 'effects.sky_color' is missing"
  - "Invalid mob category 'invalid_category', valid options are: [monster, creature, ...]"

### Phase 3: Developer Experience (Medium Priority)

#### 3.1 Add JSON Schema Generation
- **Feature**: Generate JSON Schema files from mcdoc schemas
- **Benefit**: IDE autocompletion and validation in editors
- **Implementation**: New command `tsmc generate-schema --output schemas/`

#### 3.2 Fix Mode
- **Feature**: Auto-fix common validation errors
- **Command**: `tsmc validate --fix`
- **Examples**:
  - Add missing required fields with default values
  - Fix common typos in mob categories
  - Normalize color values

#### 3.3 Watch Mode
- **Feature**: Continuous validation during development
- **Command**: `tsmc validate --watch`
- **Implementation**: Use file system watchers to re-validate on changes

### Phase 4: Extended Registry Support (Low Priority)

#### 4.1 Support Full Datapack Structure
- **Current**: Only worldgen files
- **Goal**: Support all datapack registries (advancements, recipes, loot tables, etc.)
- **Files to update**: Add more mappings to `src/path-mapper.ts`

#### 4.2 Resource Pack Validation
- **Feature**: Validate resource pack JSON files (models, blockstates, etc.)
- **Schema source**: Examine `java/assets/` directory
- **New command**: `tsmc validate-resourcepack`

#### 4.3 Custom Schema Support
- **Feature**: Allow users to provide custom mcdoc schemas
- **Use case**: Modded Minecraft validation
- **Implementation**: `--custom-schemas` option

### Phase 5: Performance & Scalability (Low Priority)

#### 5.1 Optimize Schema Parsing
- **Issue**: Currently parsing all 219 schemas for every validation
- **Solution**: Cache parsed schemas, incremental parsing
- **Target**: Sub-second validation for single files

#### 5.2 Parallel Validation
- **Enhancement**: Validate multiple files concurrently
- **Implementation**: Worker threads or async batching
- **Benefit**: Faster validation of large datapacks

#### 5.3 Memory Optimization
- **Analysis needed**: Profile memory usage during large validation runs
- **Goal**: Handle validation of thousands of files efficiently

## 🛠 Development Workflow

### Before Starting Any Work:
1. **Test current functionality**: `npm run build && node dist/index.js validate`
2. **Run schema parsing**: `node dist/index.js parse-schemas` to ensure no regressions
3. **Create feature branch**: Use descriptive branch names

### For Schema Mapping Work:
1. **Research first**: Use `grep` and file exploration to understand schema structure
2. **Test incrementally**: Add one registry type at a time
3. **Validate thoroughly**: Test with real datapack files

### For Validation Logic:
1. **Study Spyglass source**: Understand how official validation works
2. **Start small**: Implement one registry type fully before expanding
3. **Add comprehensive tests**: Create test cases for edge cases

## 📊 Success Metrics

### Phase 1 Success:
- **Target**: 90%+ files validate successfully (up from current 71%)
- **Measure**: `node dist/index.js validate` shows <50 validation errors

### Phase 2 Success:
- **Target**: Catch real validation errors in vanilla Minecraft files
- **Measure**: Find and report actual invalid values/structures

### Overall Project Success:
- **Adoption**: Tool used in CI pipelines for datapack development
- **Community**: Positive feedback from Minecraft modding community
- **Reliability**: <1% false positive rate on valid Minecraft data

## 🔧 Quick Wins (Start Here!)

1. **Add noise registry mapping** (30 minutes)
   - Update `src/path-mapper.ts` with `worldgen/noise` mapping
   - Test with `node dist/index.js validate --verbose`

2. **Improve error messages for biomes** (1 hour)
   - Add specific validation rules in `validateBiome()`
   - Add helpful suggestions for common errors

3. **Add flat level generator preset mapping** (30 minutes)
   - Another easy schema mapping addition
   - Quick validation improvement

Start with these quick wins to get familiar with the codebase, then tackle the larger architectural improvements!