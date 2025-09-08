import * as core from '@spyglassmc/core';
import { NodeJsExternals } from '@spyglassmc/core/lib/nodejs.js';
import * as mcdoc from '@spyglassmc/mcdoc';
import * as json from '@spyglassmc/json';
import * as je from '@spyglassmc/java-edition';
import { resolve } from 'path';
import { pathToFileURL } from 'url';
import { readFileSync } from 'fs';
import { glob } from 'glob';
export class SpyglassValidator {
    project = null;
    verbose;
    constructor(verbose = false) {
        this.verbose = verbose;
    }
    /**
     * Initialize the Spyglass project with mcdoc schemas
     */
    async initialize(schemaPath) {
        const cacheRoot = resolve(process.cwd(), '.cache');
        const projectRoot = resolve(process.cwd(), schemaPath);
        if (this.verbose) {
            console.log(`Initializing Spyglass project...`);
            console.log(`Schema root: ${projectRoot}`);
            console.log(`Cache root: ${cacheRoot}`);
        }
        const logger = {
            log: (...args) => this.verbose ? console.log(...args) : {},
            info: (...args) => this.verbose ? console.info(...args) : {},
            warn: (...args) => console.warn(...args),
            error: (...args) => console.error(...args),
        };
        this.project = new core.Project({
            logger,
            profilers: new core.ProfilerFactory(logger, [
                'project#init',
                'project#ready',
                'project#ready#bind',
            ]),
            cacheRoot: core.fileUtil.ensureEndingSlash(pathToFileURL(cacheRoot).toString()),
            defaultConfig: core.ConfigService.merge(core.VanillaConfig, {
                env: { dependencies: [] },
            }),
            externals: NodeJsExternals,
            initializers: [
                mcdoc.initialize,
                je.initialize,
                json.getInitializer(),
            ],
            projectRoots: [core.fileUtil.ensureEndingSlash(pathToFileURL(projectRoot).toString())],
        });
        await this.project.ready();
        await this.project.cacheService.save();
        if (this.verbose) {
            console.log('Spyglass project initialized successfully');
        }
    }
    /**
     * Validate a JSON file against mcdoc schemas using Spyglass
     */
    async validateJsonFile(filePath, datpackRoot) {
        if (!this.project) {
            throw new Error('Spyglass project not initialized. Call initialize() first.');
        }
        // Parse the file path to get resource information
        const fileInfo = this.parseDatapackPath(filePath, datpackRoot);
        if (!fileInfo) {
            return {
                filePath,
                resourceId: filePath,
                valid: false,
                errors: [{
                        message: `Unable to parse datapack path: ${filePath}`,
                        severity: 'error'
                    }],
                warnings: []
            };
        }
        // Read the JSON file
        let jsonContent;
        try {
            jsonContent = readFileSync(filePath, 'utf-8');
        }
        catch (error) {
            return {
                filePath,
                resourceId: fileInfo.resourceId,
                valid: false,
                errors: [{
                        message: `Failed to read file: ${error}`,
                        severity: 'error'
                    }],
                warnings: []
            };
        }
        // Create a virtual document for Spyglass
        const fileUri = pathToFileURL(filePath).toString();
        try {
            // Parse the JSON using Spyglass JSON parser
            const source = new core.Source(jsonContent, fileUri);
            const ctx = new core.ParserContext(this.project.meta, source, this.project.logger, this.project.symbols);
            // Parse JSON
            const jsonResult = json.parser.file(ctx);
            if (!jsonResult || jsonResult.errors.length > 0) {
                return {
                    filePath,
                    resourceId: fileInfo.resourceId,
                    valid: false,
                    errors: jsonResult?.errors.map(err => ({
                        message: err.message,
                        range: err.range ? {
                            start: { line: err.range.start.line, character: err.range.start.character },
                            end: { line: err.range.end.line, character: err.range.end.character }
                        } : undefined,
                        severity: 'error',
                        code: err.code
                    })) || [{
                            message: 'Failed to parse JSON',
                            severity: 'error'
                        }],
                    warnings: []
                };
            }
            // Try to get the mcdoc type for this registry type
            const mcdocType = this.getMcdocTypeForRegistry(fileInfo.registryType);
            if (!mcdocType) {
                return {
                    filePath,
                    resourceId: fileInfo.resourceId,
                    valid: false,
                    errors: [{
                            message: `No mcdoc schema found for registry type: ${fileInfo.registryType}`,
                            severity: 'error'
                        }],
                    warnings: []
                };
            }
            // Validate using Spyglass mcdoc checker
            const checkerCtx = new core.CheckerContext(this.project.meta, source, this.project.logger, this.project.symbols);
            const validator = json.checker.index(mcdocType);
            // Clear any existing errors
            source.clearErrors();
            // Run validation
            validator(jsonResult.node, checkerCtx);
            // Convert errors to our format
            const errors = [];
            const warnings = [];
            for (const error of source.errors) {
                const converted = {
                    message: error.message,
                    range: error.range ? {
                        start: { line: error.range.start.line, character: error.range.start.character },
                        end: { line: error.range.end.line, character: error.range.end.character }
                    } : undefined,
                    severity: error.severity || 'error',
                    code: error.code
                };
                if (converted.severity === 'warning') {
                    warnings.push(converted);
                }
                else {
                    errors.push(converted);
                }
            }
            return {
                filePath,
                resourceId: fileInfo.resourceId,
                valid: errors.length === 0,
                errors,
                warnings
            };
        }
        catch (error) {
            return {
                filePath,
                resourceId: fileInfo.resourceId,
                valid: false,
                errors: [{
                        message: `Validation failed: ${error}`,
                        severity: 'error'
                    }],
                warnings: []
            };
        }
    }
    /**
     * Validate multiple JSON files
     */
    async validateAllJsonFiles(datpackPath) {
        const pattern = `${datpackPath}/**/*.json`;
        const files = await glob(pattern, {
            ignore: ['**/node_modules/**', '**/_*'],
            absolute: true
        });
        if (this.verbose) {
            console.log(`Found ${files.length} JSON files for validation`);
        }
        const results = [];
        for (const filePath of files) {
            try {
                const result = await this.validateJsonFile(filePath, datpackPath);
                results.push(result);
                if (this.verbose) {
                    const status = result.valid ? '✓' : '✗';
                    console.log(`${status} ${result.resourceId}`);
                    if (!result.valid && result.errors.length > 0) {
                        for (const error of result.errors.slice(0, 2)) {
                            console.log(`  - ${error.message}`);
                        }
                        if (result.errors.length > 2) {
                            console.log(`  - ... and ${result.errors.length - 2} more errors`);
                        }
                    }
                }
            }
            catch (error) {
                if (this.verbose) {
                    console.error(`Failed to validate ${filePath}:`, error);
                }
                results.push({
                    filePath,
                    resourceId: filePath,
                    valid: false,
                    errors: [{
                            message: `Validation failed: ${error}`,
                            severity: 'error'
                        }],
                    warnings: []
                });
            }
        }
        return results;
    }
    /**
     * Get mcdoc type for a registry type using Spyglass symbol system
     */
    getMcdocTypeForRegistry(registryType) {
        if (!this.project) {
            return null;
        }
        // Try various patterns to find the mcdoc type
        const patterns = [
            `minecraft:resource["${registryType}"]`,
            `minecraft:resource[${registryType}]`,
            registryType,
        ];
        // Also try without worldgen/ prefix for types like dimension
        if (registryType.startsWith('worldgen/')) {
            const withoutPrefix = registryType.substring('worldgen/'.length);
            patterns.push(`minecraft:resource["${withoutPrefix}"]`, `minecraft:resource[${withoutPrefix}]`, withoutPrefix);
        }
        for (const pattern of patterns) {
            try {
                // Look for dispatcher symbols
                const dispatchers = this.project.symbols.getVisibleSymbols('mcdoc/dispatcher');
                for (const [name, symbol] of Object.entries(dispatchers)) {
                    if (name === pattern && symbol.members) {
                        // Found a dispatcher, but we need the default type
                        // This is a simplified approach - in practice, we'd need to handle 
                        // the specific dispatch logic
                        continue;
                    }
                }
                // Look for direct type symbols
                const symbols = this.project.symbols.getVisibleSymbols('mcdoc');
                for (const [name, symbol] of Object.entries(symbols)) {
                    if (name.includes(registryType) && mcdoc.binder.TypeDefSymbolData.is(symbol.data)) {
                        return symbol.data.typeDef;
                    }
                }
            }
            catch (error) {
                // Continue trying other patterns
                if (this.verbose) {
                    console.log(`Failed to resolve pattern ${pattern}:`, error);
                }
            }
        }
        return null;
    }
    /**
     * Parse datapack file path (simplified version)
     */
    parseDatapackPath(filePath, datpackRoot) {
        try {
            const relativePath = filePath.replace(datpackRoot + '/', '');
            const pathParts = relativePath.split('/').filter(part => part.length > 0);
            if (pathParts.length < 2) {
                return null;
            }
            const resourceName = pathParts[pathParts.length - 1].replace('.json', '');
            const registryType = pathParts.slice(0, -1).join('/');
            // Add worldgen/ prefix if not present for consistency with mcdoc schemas
            const normalizedRegistryType = registryType.startsWith('worldgen/')
                ? registryType
                : `worldgen/${registryType}`;
            return {
                registryType: normalizedRegistryType,
                resourceName,
                resourceId: `minecraft:${resourceName}`
            };
        }
        catch (error) {
            return null;
        }
    }
    /**
     * Close the Spyglass project
     */
    async close() {
        if (this.project) {
            await this.project.close();
            this.project = null;
        }
    }
    /**
     * Generate a validation report
     */
    generateReport(results) {
        return {
            totalFiles: results.length,
            validFiles: results.filter(r => r.valid).length,
            invalidFiles: results.filter(r => !r.valid).length,
            totalErrors: results.reduce((sum, r) => sum + r.errors.length, 0),
            totalWarnings: results.reduce((sum, r) => sum + r.warnings.length, 0),
        };
    }
}
//# sourceMappingURL=spyglass-validator.js.map