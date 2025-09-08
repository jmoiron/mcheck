/**
 * Base interface that all validators must implement
 */
export interface ValidationResult {
    filePath: string;
    resourceId: string;
    valid: boolean;
    errors: ValidationError[];
    warnings: ValidationError[];
}
export interface ValidationError {
    message: string;
    range?: {
        start: {
            line: number;
            character: number;
        };
        end: {
            line: number;
            character: number;
        };
    };
    severity: 'error' | 'warning';
    code?: string;
}
export interface ValidationReport {
    totalFiles: number;
    validFiles: number;
    invalidFiles: number;
    totalErrors: number;
    totalWarnings: number;
    byRegistryType?: Record<string, {
        total: number;
        valid: number;
        invalid: number;
    }>;
}
/**
 * Abstract base class for all validators
 */
export declare abstract class BaseValidator {
    protected verbose: boolean;
    constructor(verbose?: boolean);
    /**
     * Initialize the validator with schema path and optional datapack path
     */
    abstract initialize(schemaPath: string, datpackPath?: string): Promise<void>;
    /**
     * Validate a single JSON file
     */
    abstract validateJsonFile(filePath: string, datpackRoot: string): Promise<ValidationResult>;
    /**
     * Validate multiple JSON files
     */
    abstract validateAllJsonFiles(datpackPath: string): Promise<ValidationResult[]>;
    /**
     * Generate a validation report from results
     */
    generateReport(results: ValidationResult[]): ValidationReport;
    /**
     * Extract registry type from resource ID (basic implementation)
     */
    protected extractRegistryType(resourceId: string): string | null;
    /**
     * Close/cleanup the validator
     */
    abstract close(): Promise<void>;
    /**
     * Get the validator name for display
     */
    abstract getValidatorName(): string;
}
