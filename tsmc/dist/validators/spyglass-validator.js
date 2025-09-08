import { BaseValidator } from './base-validator.js';
import { glob } from 'glob';
import { readFileSync } from 'fs';
import { basename, resolve } from 'path';
import { pathToFileURL } from 'url';
import * as core from '@spyglassmc/core';
import * as mcdoc from '@spyglassmc/mcdoc';
import * as json from '@spyglassmc/json';
import * as javaEdition from '@spyglassmc/java-edition';
import { NodeJsExternals } from '@spyglassmc/core/lib/nodejs.js';
/**
 * Spyglass-based validator using real Spyglass Project APIs
 */
export class SpyglassValidator extends BaseValidator {
    project = null;
    schemaPath = '';
    async initialize(schemaPath, datpackPath) {
        this.schemaPath = schemaPath;
        if (this.verbose) {
            console.log('🔧 Initializing Spyglass validator...');
            console.log(`Schema directory: ${schemaPath}`);
            if (datpackPath) {
                console.log(`Datapack directory: ${datpackPath}`);
            }
        }
        try {
            // Create Spyglass logger
            const logger = {
                log: (...args) => this.verbose ? console.log(...args) : {},
                info: (...args) => this.verbose ? console.info(...args) : {},
                warn: (...args) => console.warn(...args),
                error: (...args) => console.error(...args),
            };
            // Set up project paths
            const cacheRoot = resolve(process.cwd(), '.cache');
            const schemaRoot = resolve(process.cwd(), schemaPath);
            // Build project roots array - include both schema and datapack if provided
            const projectRoots = [
                core.fileUtil.ensureEndingSlash(pathToFileURL(schemaRoot).toString())
            ];
            if (datpackPath) {
                const datpackRoot = resolve(process.cwd(), datpackPath);
                projectRoots.push(core.fileUtil.ensureEndingSlash(pathToFileURL(datpackRoot).toString()));
            }
            // Create Spyglass project
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
                    gameVersion: '1.21.8', // Specify concrete Minecraft version
                }),
                externals: NodeJsExternals,
                initializers: [
                    mcdoc.initialize,
                    json.getInitializer(),
                    javaEdition.initialize,
                ],
                projectRoots,
            });
            // Wait for project to be ready
            await this.project.ready();
            await this.project.cacheService.save();
            if (this.verbose) {
                console.log('✅ Spyglass validator initialized');
            }
        }
        catch (error) {
            throw new Error(`Failed to initialize Spyglass project: ${error}`);
        }
    }
    async validateJsonFile(filePath, datpackRoot) {
        if (!this.project) {
            throw new Error('Validator not initialized. Call initialize() first.');
        }
        // Parse basic file info
        const fileInfo = this.parseDatapackPath(filePath, datpackRoot);
        const resourceId = fileInfo?.resourceId || basename(filePath, '.json');
        try {
            // Read file content
            const content = readFileSync(filePath, 'utf-8');
            // Basic JSON syntax validation first
            try {
                JSON.parse(content);
            }
            catch (jsonError) {
                return {
                    filePath,
                    resourceId,
                    valid: false,
                    errors: [{
                            message: `JSON syntax error: ${jsonError}`,
                            severity: 'error',
                            range: undefined
                        }],
                    warnings: []
                };
            }
            // Convert file path to URI for Spyglass
            const fileUri = pathToFileURL(filePath).toString();
            // Check if this is a file Spyglass can validate by looking at the project's supported files
            const isFileSupported = !this.project.shouldExclude(fileUri, 'json');
            if (isFileSupported) {
                // Use Spyglass to process the file as a client-managed document
                await this.project.onDidOpen(fileUri, 'json', 1, content);
                // Get the processed document and node
                const docAndNode = await this.project.ensureClientManagedChecked(fileUri);
                if (docAndNode) {
                    // Extract errors from the file node using Spyglass API
                    const errors = core.FileNode.getErrors(docAndNode.node);
                    // Clean up the managed document
                    this.project.onDidClose(fileUri);
                    return {
                        filePath,
                        resourceId,
                        valid: errors.length === 0,
                        errors: errors.map((err) => ({
                            message: err.message,
                            // TODO: Convert Spyglass Range (offset-based) to LSP Position (line/character-based)
                            range: undefined,
                            severity: 'error',
                            code: err.info?.codeAction?.title
                        })),
                        warnings: [] // Spyglass doesn't separate warnings from errors
                    };
                }
                // Clean up in case of failure
                this.project.onDidClose(fileUri);
            }
            // If Spyglass can't validate this file, fall back to JSON syntax validation
            if (this.verbose) {
                console.log(`  File ${resourceId} not supported by Spyglass, using syntax validation only`);
            }
            return {
                filePath,
                resourceId,
                valid: true, // JSON syntax is valid
                errors: [],
                warnings: []
            };
        }
        catch (error) {
            return {
                filePath,
                resourceId,
                valid: false,
                errors: [{
                        message: `Validation error: ${error}`,
                        severity: 'error'
                    }],
                warnings: []
            };
        }
    }
    async validateAllJsonFiles(datpackPath) {
        if (!this.project) {
            throw new Error('Validator not initialized. Call initialize() first.');
        }
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
        return results;
    }
    /**
     * Parse datapack file path to extract resource information
     */
    parseDatapackPath(filePath, datpackRoot) {
        try {
            const relativePath = filePath.replace(datpackRoot + '/', '');
            const pathParts = relativePath.split('/').filter(part => part.length > 0);
            // Handle flat structure like: dimension_type/file.json
            if (pathParts.length >= 2) {
                const resourceName = pathParts[pathParts.length - 1].replace('.json', '');
                const registryType = pathParts.slice(0, -1).join('/');
                return {
                    registryType,
                    resourceName,
                    resourceId: `minecraft:${resourceName}` // Default to minecraft namespace
                };
            }
            return null;
        }
        catch (error) {
            return null;
        }
    }
    extractRegistryType(resourceId) {
        // TODO: Extract from actual validation context
        return 'unknown';
    }
    async close() {
        if (this.verbose) {
            console.log('Closing Spyglass validator...');
        }
        if (this.project) {
            await this.project.close();
            this.project = null;
        }
    }
    getValidatorName() {
        return 'Spyglass';
    }
}
//# sourceMappingURL=spyglass-validator.js.map