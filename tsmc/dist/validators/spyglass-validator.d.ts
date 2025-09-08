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
    close(): Promise<void>;
    getValidatorName(): string;
}
