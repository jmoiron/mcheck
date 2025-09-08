import type { ParsedMcdocFile } from './mcdoc-parser.js';
import type { DatapackFileInfo } from './path-mapper.js';
export interface JsonValidationResult {
    filePath: string;
    fileInfo: DatapackFileInfo;
    valid: boolean;
    errors: ValidationError[];
    warnings: ValidationError[];
    expectedType?: string;
    actualContent?: any;
}
export interface ValidationError {
    message: string;
    path?: string;
    line?: number;
    column?: number;
    severity: 'error' | 'warning';
}
export declare class JsonValidator {
    private pathMapper;
    private schemas;
    private verbose;
    constructor(parsedSchemas: ParsedMcdocFile[], verbose?: boolean);
    /**
     * Discover JSON files in a datapack directory
     */
    discoverJsonFiles(datpackPath: string): Promise<string[]>;
    /**
     * Validate a single JSON file against mcdoc schemas
     */
    validateJsonFile(filePath: string, datpackRoot?: string): Promise<JsonValidationResult>;
    /**
     * Validate multiple JSON files
     */
    validateAllJsonFiles(datpackPath: string): Promise<JsonValidationResult[]>;
    /**
     * Perform basic validation (placeholder for full mcdoc validation)
     */
    private performBasicValidation;
    /**
     * Validate biome structure based on mcdoc schema
     */
    private validateBiome;
    /**
     * Validate a noise settings structure
     */
    private validateNoiseSettings;
    /**
     * Generate a validation report
     */
    generateReport(results: JsonValidationResult[]): {
        totalFiles: number;
        validFiles: number;
        invalidFiles: number;
        totalErrors: number;
        totalWarnings: number;
        byRegistryType: Record<string, {
            total: number;
            valid: number;
            invalid: number;
        }>;
    };
}
