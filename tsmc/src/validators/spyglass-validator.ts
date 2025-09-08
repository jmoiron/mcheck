import { BaseValidator, ValidationResult } from './base-validator.js';
import { glob } from 'glob';
import { readFileSync } from 'fs';
import { basename } from 'path';

/**
 * Spyglass-based validator (stub implementation for now)
 * TODO: Implement using actual Spyglass APIs
 */
export class SpyglassValidator extends BaseValidator {
  private initialized = false;
  private schemaPath = '';

  async initialize(schemaPath: string): Promise<void> {
    this.schemaPath = schemaPath;
    
    if (this.verbose) {
      console.log('🔧 Initializing Spyglass validator...');
      console.log(`Schema directory: ${schemaPath}`);
      console.log('⚠️  Using stub implementation - all files will pass validation');
    }

    // TODO: Initialize actual Spyglass Project here
    // const project = new core.Project({...});
    // await project.ready();

    this.initialized = true;
    
    if (this.verbose) {
      console.log('✅ Spyglass validator initialized');
    }
  }

  async validateJsonFile(filePath: string, datpackRoot: string): Promise<ValidationResult> {
    if (!this.initialized) {
      throw new Error('Validator not initialized. Call initialize() first.');
    }

    // Parse basic file info
    const fileInfo = this.parseDatapackPath(filePath, datpackRoot);
    
    // Read and validate JSON syntax
    let jsonContent: any;
    try {
      const content = readFileSync(filePath, 'utf-8');
      jsonContent = JSON.parse(content);
    } catch (error) {
      return {
        filePath,
        resourceId: fileInfo?.resourceId || basename(filePath, '.json'),
        valid: false,
        errors: [{
          message: `JSON parse error: ${error}`,
          severity: 'error'
        }],
        warnings: []
      };
    }

    // TODO: Implement actual Spyglass validation here
    // 1. Parse JSON with json.parser.file()
    // 2. Look up mcdoc type from project.symbols
    // 3. Validate with json.checker.index(mcdocType)
    
    // Stub: Always pass validation for now
    return {
      filePath,
      resourceId: fileInfo?.resourceId || basename(filePath, '.json'),
      valid: true,
      errors: [],
      warnings: []
    };
  }

  async validateAllJsonFiles(datpackPath: string): Promise<ValidationResult[]> {
    if (!this.initialized) {
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

    const results: ValidationResult[] = [];
    
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
  private parseDatapackPath(filePath: string, datpackRoot: string) {
    try {
      const relativePath = filePath.replace(datpackRoot + '/', '');
      const pathParts = relativePath.split('/').filter(part => part.length > 0);
      
      if (pathParts.length < 2) {
        return null;
      }

      const resourceName = pathParts[pathParts.length - 1].replace('.json', '');
      const registryType = pathParts.slice(0, -1).join('/');
      
      return {
        registryType,
        resourceName,
        resourceId: `minecraft:${resourceName}`
      };
    } catch (error) {
      return null;
    }
  }

  protected extractRegistryType(resourceId: string): string | null {
    // TODO: Extract from actual validation context
    return 'unknown';
  }

  async close(): Promise<void> {
    if (this.verbose) {
      console.log('Closing Spyglass validator...');
    }
    
    // TODO: Close Spyglass project
    // if (this.project) {
    //   await this.project.close();
    // }
    
    this.initialized = false;
  }

  getValidatorName(): string {
    return 'Spyglass (stub)';
  }
}