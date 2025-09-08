#!/usr/bin/env node

import { readFileSync, existsSync } from 'fs';
import { join, resolve, relative } from 'path';
import { glob } from 'glob';
import type { Source } from '@spyglassmc/core';
import { Source as SourceImpl } from '@spyglassmc/core';

export interface McdocFile {
  path: string;
  content: string;
  source: Source;
}

export class McdocLoader {
  private readonly schemaRoot: string;
  private readonly verbose: boolean;

  constructor(schemaRoot: string = './java', verbose: boolean = false) {
    this.schemaRoot = resolve(schemaRoot);
    this.verbose = verbose;
    if (!existsSync(this.schemaRoot)) {
      throw new Error(`Schema root directory does not exist: ${this.schemaRoot}`);
    }
  }

  /**
   * Discover all .mcdoc files in the schema directory
   */
  async discoverMcdocFiles(): Promise<string[]> {
    const pattern = join(this.schemaRoot, '**/*.mcdoc');
    const files = await glob(pattern, { 
      ignore: ['**/node_modules/**'],
      absolute: true 
    });
    
    if (this.verbose) {
      console.log(`Found ${files.length} mcdoc files in ${this.schemaRoot}`);
    }
    return files;
  }

  /**
   * Load a single mcdoc file and create a Source object
   */
  loadMcdocFile(filePath: string): McdocFile {
    if (!existsSync(filePath)) {
      throw new Error(`Mcdoc file does not exist: ${filePath}`);
    }

    const content = readFileSync(filePath, 'utf-8');
    const relativePath = relative(this.schemaRoot, filePath);
    
    // Create a Source object for the file content
    const source = new SourceImpl(content);
    
    return {
      path: relativePath,
      content,
      source
    };
  }

  /**
   * Load all discovered mcdoc files
   */
  async loadAllMcdocFiles(): Promise<McdocFile[]> {
    const filePaths = await this.discoverMcdocFiles();
    const mcdocFiles: McdocFile[] = [];

    for (const filePath of filePaths) {
      try {
        const mcdocFile = this.loadMcdocFile(filePath);
        mcdocFiles.push(mcdocFile);
        if (this.verbose) {
          console.log(`Loaded: ${mcdocFile.path}`);
        }
      } catch (error) {
        console.error(`Failed to load ${filePath}:`, error);
      }
    }

    return mcdocFiles;
  }

  /**
   * Get the schema root directory
   */
  getSchemaRoot(): string {
    return this.schemaRoot;
  }
}