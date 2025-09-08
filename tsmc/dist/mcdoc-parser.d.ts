import * as core from '@spyglassmc/core';
import type { McdocFile } from './mcdoc-loader.js';
export interface ParsedMcdocFile extends McdocFile {
    project?: core.Project;
    errors: core.LanguageError[];
    success: boolean;
}
export declare class McdocParser {
    private project?;
    private verbose;
    constructor(verbose?: boolean);
    /**
     * Initialize the Spyglass project for parsing
     */
    private ensureProject;
    /**
     * Parse a single mcdoc file using Spyglass
     */
    parseMcdocFile(mcdocFile: McdocFile): Promise<ParsedMcdocFile>;
    /**
     * Parse multiple mcdoc files
     */
    parseAllMcdocFiles(mcdocFiles: McdocFile[]): Promise<ParsedMcdocFile[]>;
    /**
     * Clean up resources
     */
    close(): Promise<void>;
    /**
     * Get detailed information about parsing errors
     */
    getParsingReport(parsedFiles: ParsedMcdocFile[]): {
        totalFiles: number;
        successfulFiles: number;
        filesWithErrors: number;
        totalErrors: number;
        errorDetails: Array<{
            file: string;
            errors: any[];
        }>;
    };
}
