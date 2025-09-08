export interface SpyglassValidationResult {
    filePath: string;
    resourceId: string;
    valid: boolean;
    errors: SpyglassValidationError[];
    warnings: SpyglassValidationError[];
}
export interface SpyglassValidationError {
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
export declare class SpyglassValidator {
    private project;
    private verbose;
    constructor(verbose?: boolean);
    /**
     * Initialize the Spyglass project with mcdoc schemas
     */
    initialize(schemaPath: string): Promise<void>;
    /**
     * Validate a JSON file against mcdoc schemas using Spyglass
     */
    validateJsonFile(filePath: string, datpackRoot: string): Promise<SpyglassValidationResult>;
    /**
     * Validate multiple JSON files
     */
    validateAllJsonFiles(datpackPath: string): Promise<SpyglassValidationResult[]>;
    /**
     * Get mcdoc type for a registry type using Spyglass symbol system
     */
    private getMcdocTypeForRegistry;
    /**
     * Parse datapack file path (simplified version)
     */
    private parseDatapackPath;
    /**
     * Close the Spyglass project
     */
    close(): Promise<void>;
    /**
     * Generate a validation report
     */
    generateReport(results: SpyglassValidationResult[]): {
        totalFiles: number;
        validFiles: number;
        invalidFiles: number;
        totalErrors: number;
        totalWarnings: number;
    };
}
