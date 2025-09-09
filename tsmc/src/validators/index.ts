import { BaseValidator, ValidatorOptions } from './base-validator.js';
import { LegacyValidator } from './legacy-validator.js';
import { SpyglassValidator } from './spyglass-validator.js';

export * from './base-validator.js';
export * from './legacy-validator.js';
export * from './spyglass-validator.js';

export type ValidatorType = 'legacy' | 'spyglass';

/**
 * Factory function to create validators
 */
export function createValidator(type: ValidatorType, options: ValidatorOptions = {}): BaseValidator {
  switch (type) {
    case 'legacy':
      return new LegacyValidator(options);
    case 'spyglass':
      return new SpyglassValidator(options);
    default:
      throw new Error(`Unknown validator type: ${type}`);
  }
}

/**
 * Legacy function for backward compatibility
 */
export function createValidatorWithVerbose(type: ValidatorType, verbose: boolean = false): BaseValidator {
  return createValidator(type, { verbose });
}

/**
 * Get list of available validator types
 */
export function getAvailableValidators(): ValidatorType[] {
  return ['legacy', 'spyglass'];
}

/**
 * Get the default validator type
 */
export function getDefaultValidator(): ValidatorType {
  return 'legacy';
}