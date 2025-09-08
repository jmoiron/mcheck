/**
 * Abstract base class for all validators
 */
export class BaseValidator {
    verbose;
    constructor(verbose = false) {
        this.verbose = verbose;
    }
    /**
     * Generate a validation report from results
     */
    generateReport(results) {
        const byRegistryType = {};
        // Group by registry type if available
        for (const result of results) {
            const registryType = this.extractRegistryType(result.resourceId) || 'unknown';
            if (!byRegistryType[registryType]) {
                byRegistryType[registryType] = { total: 0, valid: 0, invalid: 0 };
            }
            byRegistryType[registryType].total++;
            if (result.valid) {
                byRegistryType[registryType].valid++;
            }
            else {
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
    extractRegistryType(resourceId) {
        // Override in subclasses for more sophisticated extraction
        return null;
    }
}
//# sourceMappingURL=base-validator.js.map