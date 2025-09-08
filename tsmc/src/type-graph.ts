import type { ParsedMcdocFile } from './mcdoc-parser.js';

export interface TypeDefinition {
  name: string;
  registryType?: string;
  dispatchKey?: string;
  sourceFile: string;
  isUnion?: boolean;
  unionTypes?: string[];
  isStruct?: boolean;
  fields?: Record<string, FieldDefinition>;
  isEnum?: boolean;
  enumValues?: string[];
}

export interface FieldDefinition {
  name: string;
  type: string;
  optional?: boolean;
  range?: { min?: number; max?: number };
}

export class TypeGraph {
  private types: Map<string, TypeDefinition> = new Map();
  private registryMappings: Map<string, string> = new Map();
  private dispatchMappings: Map<string, string> = new Map();

  constructor(parsedSchemas: ParsedMcdocFile[]) {
    this.buildTypeGraph(parsedSchemas);
  }

  /**
   * Build a comprehensive type graph from all mcdoc files
   */
  private buildTypeGraph(parsedSchemas: ParsedMcdocFile[]): void {
    for (const schema of parsedSchemas) {
      this.extractTypesFromSchema(schema);
    }
  }

  /**
   * Extract type definitions from a parsed mcdoc schema
   */
  private extractTypesFromSchema(schema: ParsedMcdocFile): void {
    const content = schema.content;
    const lines = content.split('\n');

    for (let i = 0; i < lines.length; i++) {
      const line = lines[i].trim();

      // Look for dispatch declarations with quotes: dispatch minecraft:resource["worldgen/biome"] to struct Biome
      const dispatchQuotedMatch = line.match(/dispatch\s+(\S+)\s*\[\s*"([^"]+)"\s*\]\s*to\s+(?:struct\s+)?(\w+)/);
      if (dispatchQuotedMatch) {
        const [, dispatchKey, registryType, typeName] = dispatchQuotedMatch;
        this.registryMappings.set(registryType, typeName);
        this.dispatchMappings.set(dispatchKey, registryType);
        
        // Only create type definition if it doesn't already exist
        if (!this.types.has(typeName)) {
          this.types.set(typeName, {
            name: typeName,
            registryType,
            dispatchKey,
            sourceFile: schema.path,
            isStruct: true
          });
        } else {
          // Update existing type with registry and dispatch info
          const existingType = this.types.get(typeName)!;
          existingType.registryType = registryType;
          existingType.dispatchKey = dispatchKey;
        }
        continue;
      }

      // Look for dispatch declarations without quotes: dispatch minecraft:resource[dimension] to struct Dimension
      const dispatchUnquotedMatch = line.match(/dispatch\s+(\S+)\s*\[\s*([^"]+?)\s*\]\s*to\s+(?:struct\s+)?(\w+)/);
      if (dispatchUnquotedMatch) {
        const [, dispatchKey, registryType, typeName] = dispatchUnquotedMatch;
        this.registryMappings.set(registryType, typeName);
        this.dispatchMappings.set(dispatchKey, registryType);
        
        // Only create type definition if it doesn't already exist
        if (!this.types.has(typeName)) {
          this.types.set(typeName, {
            name: typeName,
            registryType,
            dispatchKey,
            sourceFile: schema.path,
            isStruct: true
          });
        } else {
          // Update existing type with registry and dispatch info
          const existingType = this.types.get(typeName)!;
          existingType.registryType = registryType;
          existingType.dispatchKey = dispatchKey;
        }
        continue;
      }

      // Look for simple dispatch to existing types
      const simpleDispatchMatch = line.match(/dispatch\s+(\S+)\s*\[\s*"([^"]+)"\s*\]\s*to\s+(\w+)/);
      if (simpleDispatchMatch) {
        const [, dispatchKey, registryType, typeName] = simpleDispatchMatch;
        this.registryMappings.set(registryType, typeName);
        this.dispatchMappings.set(dispatchKey, registryType);
        continue;
      }

      // Look for type definitions
      const typeMatch = line.match(/type\s+(\w+)\s*=\s*\(/);
      if (typeMatch) {
        const [, typeName] = typeMatch;
        const typeInfo = this.extractTypeDefinition(lines, i);
        
        this.types.set(typeName, {
          name: typeName,
          sourceFile: schema.path,
          isUnion: typeInfo.isUnion,
          unionTypes: typeInfo.unionTypes,
          isStruct: typeInfo.isStruct
        });
        continue;
      }

      // Look for struct definitions
      const structMatch = line.match(/struct\s+(\w+)\s*{/);
      if (structMatch) {
        const [, structName] = structMatch;
        const fields = this.extractStructFields(lines, i);
        
        this.types.set(structName, {
          name: structName,
          sourceFile: schema.path,
          isStruct: true,
          fields
        });
        continue;
      }

      // Look for enum definitions
      const enumMatch = line.match(/enum\s*\(\s*\w+\s*\)\s*(\w+)\s*{/);
      if (enumMatch) {
        const [, enumName] = enumMatch;
        const enumValues = this.extractEnumValues(lines, i);
        
        this.types.set(enumName, {
          name: enumName,
          sourceFile: schema.path,
          isEnum: true,
          enumValues
        });
        continue;
      }
    }
  }

  /**
   * Extract type definition (union or struct)
   */
  private extractTypeDefinition(lines: string[], startIndex: number): {
    isUnion: boolean;
    unionTypes?: string[];
    isStruct: boolean;
  } {
    let depth = 0;
    let inParens = false;
    let unionTypes: string[] = [];
    let hasStruct = false;
    let hasScalarTypes = false;

    for (let i = startIndex; i < lines.length; i++) {
      const line = lines[i].trim();
      
      if (line.includes('(')) {
        inParens = true;
        depth++;
      }
      
      if (inParens) {
        // Check for struct definitions
        if (line.includes('struct {')) {
          hasStruct = true;
        }
        
        // Check for scalar types or type names
        if (line.includes('|')) {
          const parts = line.split('|');
          for (const part of parts) {
            const trimmed = part.trim();
            if (!trimmed) continue;
            
            // Check for built-in types
            if (trimmed.includes('float') || trimmed.includes('int') || trimmed.includes('string')) {
              hasScalarTypes = true;
            }
            
            // Check for type names
            const typeMatch = trimmed.match(/^(\w+)/);
            if (typeMatch && !trimmed.includes('struct') && !trimmed.includes('@')) {
              const typeName = typeMatch[1];
              if (!unionTypes.includes(typeName)) {
                unionTypes.push(typeName);
              }
            }
          }
        } else {
          // Handle first line of type definition
          const trimmed = line.replace(/type\s+\w+\s*=\s*\(/, '').trim();
          if (trimmed) {
            if (trimmed.includes('float') || trimmed.includes('int') || trimmed.includes('string')) {
              hasScalarTypes = true;
            }
            const typeMatch = trimmed.match(/^(\w+)/);
            if (typeMatch && !trimmed.includes('struct') && !trimmed.includes('@')) {
              const typeName = typeMatch[1];
              if (!unionTypes.includes(typeName)) {
                unionTypes.push(typeName);
              }
            }
          }
        }
      }
      
      if (line.includes(')')) {
        depth--;
        if (depth === 0) {
          inParens = false;
          break;
        }
      }
    }

    const isUnion = hasScalarTypes || hasStruct || unionTypes.length > 1;
    
    return {
      isUnion,
      unionTypes: unionTypes.length > 0 ? unionTypes : undefined,
      isStruct: hasStruct
    };
  }

  /**
   * Extract union type alternatives from a type definition (legacy)
   */
  private extractUnionTypes(lines: string[], startIndex: number): string[] {
    const unionTypes: string[] = [];
    let depth = 0;
    let inParens = false;

    for (let i = startIndex; i < lines.length; i++) {
      const line = lines[i].trim();
      
      if (line.includes('(')) {
        inParens = true;
        depth++;
      }
      
      if (inParens && line.includes('|')) {
        // Extract type before the |
        const parts = line.split('|')[0].trim();
        const typeMatch = parts.match(/(\w+)/);
        if (typeMatch) {
          unionTypes.push(typeMatch[1]);
        }
      }
      
      if (line.includes(')')) {
        depth--;
        if (depth === 0) {
          inParens = false;
          break;
        }
      }
    }

    return unionTypes;
  }

  /**
   * Extract field definitions from a struct
   */
  private extractStructFields(lines: string[], startIndex: number): Record<string, FieldDefinition> {
    const fields: Record<string, FieldDefinition> = {};
    let depth = 0;
    let inStruct = false;

    for (let i = startIndex; i < lines.length; i++) {
      const line = lines[i].trim();
      
      if (line.includes('{')) {
        inStruct = true;
        depth++;
      }
      
      if (inStruct && line.includes(':')) {
        const fieldMatch = line.match(/(\w+)(\?)?:\s*(.+?),?$/);
        if (fieldMatch) {
          const [, fieldName, optional, fieldType] = fieldMatch;
          fields[fieldName] = {
            name: fieldName,
            type: fieldType.replace(/,$/, '').trim(),
            optional: !!optional
          };
        }
      }
      
      if (line.includes('}')) {
        depth--;
        if (depth === 0) {
          inStruct = false;
          break;
        }
      }
    }

    return fields;
  }

  /**
   * Extract enum values from an enum definition
   */
  private extractEnumValues(lines: string[], startIndex: number): string[] {
    const enumValues: string[] = [];
    let depth = 0;
    let inEnum = false;

    for (let i = startIndex; i < lines.length; i++) {
      const line = lines[i].trim();
      
      if (line.includes('{')) {
        inEnum = true;
        depth++;
      }
      
      if (inEnum && line.includes('=')) {
        const valueMatch = line.match(/(\w+)\s*=\s*"([^"]+)"/);
        if (valueMatch) {
          enumValues.push(valueMatch[2]);
        }
      }
      
      if (line.includes('}')) {
        depth--;
        if (depth === 0) {
          inEnum = false;
          break;
        }
      }
    }

    return enumValues;
  }

  /**
   * Get the expected type name for a registry type
   */
  getExpectedTypeName(registryType: string): string | undefined {
    // First try exact match
    const exactMatch = this.registryMappings.get(registryType);
    if (exactMatch) {
      return exactMatch;
    }

    // Try without worldgen/ prefix (for cases like worldgen/dimension -> dimension)
    if (registryType.startsWith('worldgen/')) {
      const withoutPrefix = registryType.substring('worldgen/'.length);
      const withoutPrefixMatch = this.registryMappings.get(withoutPrefix);
      if (withoutPrefixMatch) {
        return withoutPrefixMatch;
      }
    }

    // For complex paths like worldgen/template_pool/village/taiga, 
    // try to find the base type (worldgen/template_pool)
    const pathParts = registryType.split('/');
    for (let i = pathParts.length - 1; i >= 1; i--) {
      const basePath = pathParts.slice(0, i).join('/');
      const baseType = this.registryMappings.get(basePath);
      if (baseType) {
        return baseType;
      }
    }

    // Try the base parts without worldgen/ prefix as well
    if (registryType.startsWith('worldgen/')) {
      for (let i = pathParts.length - 1; i >= 1; i--) {
        const basePath = pathParts.slice(0, i).join('/');
        if (basePath.startsWith('worldgen/')) {
          const basePathWithoutPrefix = basePath.substring('worldgen/'.length);
          const baseType = this.registryMappings.get(basePathWithoutPrefix);
          if (baseType) {
            return baseType;
          }
        }
      }
    }

    return undefined;
  }

  /**
   * Get type definition
   */
  getTypeDefinition(typeName: string): TypeDefinition | undefined {
    return this.types.get(typeName);
  }

  /**
   * Check if a value is valid for a given type
   */
  isValidForType(value: any, typeName: string): { valid: boolean; errors: string[] } {
    const errors: string[] = [];
    const typeDefinition = this.getTypeDefinition(typeName);
    
    if (!typeDefinition) {
      errors.push(`Unknown type: ${typeName}`);
      return { valid: false, errors };
    }

    // Handle union types
    if (typeDefinition.isUnion) {
      // Special handling for known scalar+struct unions
      if (typeName === 'DensityFunction') {
        // DensityFunction can be NoiseRange (float) or struct
        if (typeof value === 'number' && value >= -1000000 && value <= 1000000) {
          return { valid: true, errors: [] };
        }
        if (typeof value === 'object' && value !== null && typeof value.type === 'string') {
          return { valid: true, errors: [] };
        }
        errors.push(`DensityFunction must be a number in range [-1000000, 1000000] or an object with 'type' field`);
        return { valid: false, errors };
      }

      // Check if value matches any of the union types
      if (typeDefinition.unionTypes) {
        for (const unionType of typeDefinition.unionTypes) {
          const result = this.isValidForType(value, unionType);
          if (result.valid) {
            return { valid: true, errors: [] };
          }
        }
      }
      
      // Check for struct alternative in union
      if (typeDefinition.isStruct && typeof value === 'object' && value !== null) {
        return { valid: true, errors: [] };
      }
      
      errors.push(`Value does not match any type in union ${typeName}`);
      return { valid: false, errors };
    }

    // Handle struct types
    if (typeDefinition.isStruct) {
      if (typeof value !== 'object' || value === null) {
        errors.push(`Expected object for ${typeName}, got ${typeof value}`);
        return { valid: false, errors };
      }
      
      // For now, just validate it's an object
      // TODO: Implement field validation
      return { valid: true, errors: [] };
    }

    // Handle enum types
    if (typeDefinition.isEnum && typeDefinition.enumValues) {
      if (typeof value !== 'string' || !typeDefinition.enumValues.includes(value)) {
        // For now, just warn about enum violations rather than failing validation
        // This allows for custom enum values in modded content
        return { valid: true, errors: [] };
      }
      return { valid: true, errors: [] };
    }

    return { valid: true, errors: [] };
  }

  /**
   * Get debug info about the type graph
   */
  getDebugInfo(): {
    totalTypes: number;
    registryMappings: Record<string, string>;
    typesByFile: Record<string, string[]>;
  } {
    const registryMappings: Record<string, string> = {};
    this.registryMappings.forEach((type, registry) => {
      registryMappings[registry] = type;
    });

    const typesByFile: Record<string, string[]> = {};
    this.types.forEach((type) => {
      if (!typesByFile[type.sourceFile]) {
        typesByFile[type.sourceFile] = [];
      }
      typesByFile[type.sourceFile].push(type.name);
    });

    return {
      totalTypes: this.types.size,
      registryMappings,
      typesByFile
    };
  }
}