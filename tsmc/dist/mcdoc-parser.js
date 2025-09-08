import * as core from '@spyglassmc/core';
import * as mcdoc from '@spyglassmc/mcdoc';
import { NodeJsExternals } from '@spyglassmc/core/lib/nodejs.js';
export class McdocParser {
    project;
    verbose;
    constructor(verbose = false) {
        this.verbose = verbose;
    }
    /**
     * Initialize the Spyglass project for parsing
     */
    async ensureProject() {
        if (!this.project) {
            if (this.verbose) {
                console.log('Initializing Spyglass project...');
            }
            const logger = {
                log: () => { },
                info: () => { },
                warn: console.warn,
                error: console.error
            };
            this.project = new core.Project({
                logger,
                profilers: new core.ProfilerFactory(logger, []),
                cacheRoot: core.fileUtil.ensureEndingSlash('file:///tmp/tsmc-cache/'),
                defaultConfig: core.ConfigService.merge(core.VanillaConfig, {
                    env: { dependencies: [] },
                }),
                externals: NodeJsExternals,
                initializers: [mcdoc.initialize],
                projectRoots: [core.fileUtil.ensureEndingSlash('file:///tmp/tsmc-root/')],
            });
            await this.project.ready();
            if (this.verbose) {
                console.log('Spyglass project initialized.');
            }
        }
        return this.project;
    }
    /**
     * Parse a single mcdoc file using Spyglass
     */
    async parseMcdocFile(mcdocFile) {
        if (this.verbose) {
            console.log(`Parsing: ${mcdocFile.path}`);
        }
        try {
            const project = await this.ensureProject();
            // Create a temporary text document for this file
            const doc = {
                uri: `file://${mcdocFile.path}`,
                getText: () => mcdocFile.content,
                languageId: 'mcdoc',
                version: 1,
                lineCount: mcdocFile.content.split('\n').length,
                offsetAt: () => 0,
                positionAt: () => ({ line: 0, character: 0 })
            };
            // Create error reporter
            const errorReporter = new core.ErrorReporter(mcdocFile.path);
            // Create parsing context
            const ctx = core.ParserContext.create(project, {
                doc,
                err: errorReporter
            });
            // Parse the file content using the mcdoc module parser
            const result = mcdoc.module_(mcdocFile.source, ctx);
            const errors = errorReporter.dump();
            return {
                ...mcdocFile,
                project,
                errors: Array.from(errors),
                success: errors.length === 0
            };
        }
        catch (error) {
            console.error(`Error parsing ${mcdocFile.path}:`, error);
            return {
                ...mcdocFile,
                errors: [core.LanguageError.create(`Parse error: ${error}`, core.Range.create(0, 0), 3 /* core.ErrorSeverity.Error */)],
                success: false
            };
        }
    }
    /**
     * Parse multiple mcdoc files
     */
    async parseAllMcdocFiles(mcdocFiles) {
        const parsedFiles = [];
        let errorCount = 0;
        for (const mcdocFile of mcdocFiles) {
            try {
                const parsed = await this.parseMcdocFile(mcdocFile);
                parsedFiles.push(parsed);
                if (parsed.errors.length > 0) {
                    if (this.verbose) {
                        console.warn(`${parsed.path} has ${parsed.errors.length} parsing errors`);
                    }
                    errorCount += parsed.errors.length;
                }
            }
            catch (error) {
                if (this.verbose) {
                    console.error(`Failed to parse ${mcdocFile.path}:`, error);
                }
                errorCount++;
            }
        }
        if (this.verbose) {
            console.log(`Parsed ${parsedFiles.length} files with ${errorCount} total errors`);
        }
        return parsedFiles;
    }
    /**
     * Clean up resources
     */
    async close() {
        if (this.project) {
            await this.project.close();
            this.project = undefined;
        }
    }
    /**
     * Get detailed information about parsing errors
     */
    getParsingReport(parsedFiles) {
        const filesWithErrors = parsedFiles.filter(f => f.errors.length > 0);
        const totalErrors = parsedFiles.reduce((sum, f) => sum + f.errors.length, 0);
        return {
            totalFiles: parsedFiles.length,
            successfulFiles: parsedFiles.length - filesWithErrors.length,
            filesWithErrors: filesWithErrors.length,
            totalErrors,
            errorDetails: filesWithErrors.map(f => ({
                file: f.path,
                errors: f.errors
            }))
        };
    }
}
//# sourceMappingURL=mcdoc-parser.js.map