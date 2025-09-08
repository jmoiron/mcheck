import { BaseValidator } from './base-validator.js';
export * from './base-validator.js';
export * from './legacy-validator.js';
export * from './spyglass-validator.js';
export type ValidatorType = 'legacy' | 'spyglass';
/**
 * Factory function to create validators
 */
export declare function createValidator(type: ValidatorType, verbose?: boolean): BaseValidator;
/**
 * Get list of available validator types
 */
export declare function getAvailableValidators(): ValidatorType[];
/**
 * Get the default validator type
 */
export declare function getDefaultValidator(): ValidatorType;
