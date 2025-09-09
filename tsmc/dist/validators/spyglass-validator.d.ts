import { BaseValidator, ValidationResult } from './base-validator.js';
/**
 * Spyglass-based validator using real Spyglass Project APIs
 */
export declare class SpyglassValidator extends BaseValidator {
    private project;
    private schemaPath;
    initialize(schemaPath: string, datpackPath?: string): Promise<void>;
    validateJsonFile(filePath: string, datpackRoot: string): Promise<ValidationResult>;
    validateAllJsonFiles(datpackPath: string): Promise<ValidationResult[]>;
    /**
     * Parse datapack file path to extract resource information
     */
    private parseDatapackPath;
    protected extractRegistryType(resourceId: string): string | null;
    /**
     * Enhance error messages from Spyglass to provide more context
     */
    private enhanceErrorMessage;
    /**
     * Extract context information from the error location in the JSON content
     */
    private getErrorContext;
    /**
     * Get the line number for an error based on its character offset
     */
    private getLineNumber;
    /**
     * Find the JSON path (like "noise_router.initial_density") for a given character offset
     */
    private findJsonPath;
    close(): Promise<void>;
    getValidatorName(): string;
}
