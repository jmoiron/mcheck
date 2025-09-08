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
    start: { line: number; character: number };
    end: { line: number; character: number };
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
  byRegistryType?: Record<string, { total: number; valid: number; invalid: number }>;
}

/**
 * Abstract base class for all validators
 */
export abstract class BaseValidator {
  protected verbose: boolean;
  
  constructor(verbose: boolean = false) {
    this.verbose = verbose;
  }

  /**
   * Initialize the validator with schema path
   */
  abstract initialize(schemaPath: string): Promise<void>;

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
  generateReport(results: ValidationResult[]): ValidationReport {
    const byRegistryType: Record<string, { total: number; valid: number; invalid: number }> = {};

    // Group by registry type if available
    for (const result of results) {
      const registryType = this.extractRegistryType(result.resourceId) || 'unknown';
      if (!byRegistryType[registryType]) {
        byRegistryType[registryType] = { total: 0, valid: 0, invalid: 0 };
      }
      
      byRegistryType[registryType].total++;
      if (result.valid) {
        byRegistryType[registryType].valid++;
      } else {
        byRegistryType[registryType].invalid++;
      }
    }

    return {
      totalFiles: results.length,
      validFiles: results.filter(r => r.valid).length,
      invalidFiles: results.filter(r => !r.valid).length,
      totalErrors: results.reduce((sum, r) => sum + r.errors.length, 0),
      totalWarnings: results.reduce((sum, r) => sum + r.warnings.length, 0),
      byRegistryType
    };
  }

  /**
   * Extract registry type from resource ID (basic implementation)
   */
  protected extractRegistryType(resourceId: string): string | null {
    // Override in subclasses for more sophisticated extraction
    return null;
  }

  /**
   * Close/cleanup the validator
   */
  abstract close(): Promise<void>;

  /**
   * Get the validator name for display
   */
  abstract getValidatorName(): string;
}