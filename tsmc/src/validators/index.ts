import { BaseValidator } from './base-validator.js';
import { LegacyValidator } from './legacy-validator.js';
import { SpyglassValidator } from './spyglass-validator.js';

export * from './base-validator.js';
export * from './legacy-validator.js';
export * from './spyglass-validator.js';

export type ValidatorType = 'legacy' | 'spyglass';

/**
 * Factory function to create validators
 */
export function createValidator(type: ValidatorType, verbose: boolean = false): BaseValidator {
  switch (type) {
    case 'legacy':
      return new LegacyValidator(verbose);
    case 'spyglass':
      return new SpyglassValidator(verbose);
    default:
      throw new Error(`Unknown validator type: ${type}`);
  }
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