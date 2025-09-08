import { BaseValidator } from './base-validator.js';
import { McdocLoader } from '../mcdoc-loader.js';
import { McdocParser } from '../mcdoc-parser.js';
import { JsonValidator } from '../json-validator.js';
/**
 * Legacy validator that uses our custom type graph system
 */
export class LegacyValidator extends BaseValidator {
    validator = null;
    parser = null;
    async initialize(schemaPath, datpackPath) {
        if (this.verbose) {
            console.log('📁 Loading mcdoc schemas...');
        }
        const loader = new McdocLoader(schemaPath, this.verbose);
        const mcdocFiles = await loader.loadAllMcdocFiles();
        if (mcdocFiles.length === 0) {
            throw new Error('No mcdoc schema files found!');
        }
        if (this.verbose) {
            console.log('⚡ Parsing mcdoc schemas...');
        }
        this.parser = new McdocParser(this.verbose);
        const parsedSchemas = await this.parser.parseAllMcdocFiles(mcdocFiles);
        const schemaReport = this.parser.getParsingReport(parsedSchemas);
        if (schemaReport.totalErrors > 0) {
            const errorMsg = `Schema parsing failed with ${schemaReport.totalErrors} errors`;
            if (this.verbose) {
                console.error(errorMsg);
                for (const detail of schemaReport.errorDetails) {
                    console.error(`${detail.file}:`);
                    for (const error of detail.errors) {
                        console.error(`  - ${error.message || error}`);
                    }
                }
            }
            throw new Error(errorMsg);
        }
        if (this.verbose) {
            console.log(`✅ All ${schemaReport.totalFiles} schemas parsed successfully`);
        }
        this.validator = new JsonValidator(parsedSchemas, this.verbose);
    }
    async validateJsonFile(filePath, datpackRoot) {
        if (!this.validator) {
            throw new Error('Validator not initialized. Call initialize() first.');
        }
        const result = await this.validator.validateJsonFile(filePath, datpackRoot);
        return {
            filePath: result.filePath,
            resourceId: result.fileInfo?.resourceId || result.filePath,
            valid: result.valid,
            errors: result.errors.map(err => ({
                message: err.message,
                range: err.line !== undefined ? {
                    start: { line: err.line - 1, character: err.column || 0 },
                    end: { line: err.line - 1, character: (err.column || 0) + 10 }
                } : undefined,
                severity: err.severity,
                code: err.path
            })),
            warnings: result.warnings.map(err => ({
                message: err.message,
                range: err.line !== undefined ? {
                    start: { line: err.line - 1, character: err.column || 0 },
                    end: { line: err.line - 1, character: (err.column || 0) + 10 }
                } : undefined,
                severity: err.severity,
                code: err.path
            }))
        };
    }
    async validateAllJsonFiles(datpackPath) {
        if (!this.validator) {
            throw new Error('Validator not initialized. Call initialize() first.');
        }
        const results = await this.validator.validateAllJsonFiles(datpackPath);
        return results.map(result => ({
            filePath: result.filePath,
            resourceId: result.fileInfo?.resourceId || result.filePath,
            valid: result.valid,
            errors: result.errors.map(err => ({
                message: err.message,
                range: err.line !== undefined ? {
                    start: { line: err.line - 1, character: err.column || 0 },
                    end: { line: err.line - 1, character: (err.column || 0) + 10 }
                } : undefined,
                severity: err.severity,
                code: err.path
            })),
            warnings: result.warnings.map(err => ({
                message: err.message,
                range: err.line !== undefined ? {
                    start: { line: err.line - 1, character: err.column || 0 },
                    end: { line: err.line - 1, character: (err.column || 0) + 10 }
                } : undefined,
                severity: err.severity,
                code: err.path
            }))
        }));
    }
    extractRegistryType(resourceId) {
        // The legacy validator doesn't have direct access to registry types
        // This would require refactoring the JsonValidator to expose this info
        return null;
    }
    async close() {
        if (this.parser) {
            await this.parser.close();
            this.parser = null;
        }
        this.validator = null;
    }
    getValidatorName() {
        return 'Legacy';
    }
}
//# sourceMappingURL=legacy-validator.js.map