import { BaseValidator, ValidationResult } from './base-validator.js';
/**
 * Legacy validator that uses our custom type graph system
 */
export declare class LegacyValidator extends BaseValidator {
    private validator;
    private parser;
    initialize(schemaPath: string, datpackPath?: string): Promise<void>;
    validateJsonFile(filePath: string, datpackRoot: string): Promise<ValidationResult>;
    validateAllJsonFiles(datpackPath: string): Promise<ValidationResult[]>;
    protected extractRegistryType(resourceId: string): string | null;
    close(): Promise<void>;
    getValidatorName(): string;
}
