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
    range?: {
        min?: number;
        max?: number;
    };
}
export declare class TypeGraph {
    private types;
    private registryMappings;
    private dispatchMappings;
    constructor(parsedSchemas: ParsedMcdocFile[]);
    /**
     * Build a comprehensive type graph from all mcdoc files
     */
    private buildTypeGraph;
    /**
     * Extract type definitions from a parsed mcdoc schema
     */
    private extractTypesFromSchema;
    /**
     * Extract type definition (union or struct)
     */
    private extractTypeDefinition;
    /**
     * Extract union type alternatives from a type definition (legacy)
     */
    private extractUnionTypes;
    /**
     * Extract field definitions from a struct
     */
    private extractStructFields;
    /**
     * Extract enum values from an enum definition
     */
    private extractEnumValues;
    /**
     * Get the expected type name for a registry type
     */
    getExpectedTypeName(registryType: string): string | undefined;
    /**
     * Get type definition
     */
    getTypeDefinition(typeName: string): TypeDefinition | undefined;
    /**
     * Check if a value is valid for a given type
     */
    isValidForType(value: any, typeName: string): {
        valid: boolean;
        errors: string[];
    };
    /**
     * Get debug info about the type graph
     */
    getDebugInfo(): {
        totalTypes: number;
        registryMappings: Record<string, string>;
        typesByFile: Record<string, string[]>;
    };
}
