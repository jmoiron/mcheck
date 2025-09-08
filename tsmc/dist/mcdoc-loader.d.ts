#!/usr/bin/env node
import type { Source } from '@spyglassmc/core';
export interface McdocFile {
    path: string;
    content: string;
    source: Source;
}
export declare class McdocLoader {
    private readonly schemaRoot;
    private readonly verbose;
    constructor(schemaRoot?: string, verbose?: boolean);
    /**
     * Discover all .mcdoc files in the schema directory
     */
    discoverMcdocFiles(): Promise<string[]>;
    /**
     * Load a single mcdoc file and create a Source object
     */
    loadMcdocFile(filePath: string): McdocFile;
    /**
     * Load all discovered mcdoc files
     */
    loadAllMcdocFiles(): Promise<McdocFile[]>;
    /**
     * Get the schema root directory
     */
    getSchemaRoot(): string;
}
