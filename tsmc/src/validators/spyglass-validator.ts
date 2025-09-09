import { BaseValidator, ValidationResult, ValidationError } from './base-validator.js';
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
  private project: core.Project | null = null;
  private schemaPath = '';

  async initialize(schemaPath: string, datpackPath?: string): Promise<void> {
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
      const logger: core.Logger = {
        log: (...args: any[]) => this.verbose ? console.log(...args) : {},
        info: (...args: any[]) => this.verbose ? console.info(...args) : {},
        warn: (...args: any[]) => console.warn(...args),
        error: (...args: any[]) => console.error(...args),
      };

      // Set up project paths
      const cacheRoot = resolve(process.cwd(), '.cache');
      
      // The schema path should contain a 'java' directory for ::java:: references to work
      const schemaRoot = resolve(process.cwd(), schemaPath);
      const javaDir = resolve(schemaRoot, 'java');
      
      // Check if java directory exists in the schema path
      try {
        const fs = await import('fs/promises');
        await fs.access(javaDir);
        // java directory exists - this is the expected structure
        if (this.verbose) {
          console.log(`✅ Found java directory in schema path: ${javaDir}`);
        }
      } catch {
        // java directory doesn't exist, warn the user
        console.warn(`⚠️  Warning: No 'java' directory found in schema path '${schemaPath}'. This may cause ::java:: references to fail.`);
        console.warn(`   Expected structure: ${schemaPath}/java/`);
      }
      
      // Always use schema root as project root (contains java/ directory)
      const projectRootPath = schemaRoot;
      
      // Build project roots array
      const projectRoots = [
        core.fileUtil.ensureEndingSlash(pathToFileURL(projectRootPath).toString())
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
    } catch (error) {
      throw new Error(`Failed to initialize Spyglass project: ${error}`);
    }
  }

  async validateJsonFile(filePath: string, datpackRoot: string): Promise<ValidationResult> {
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
      } catch (jsonError) {
        return {
          filePath,
          resourceId,
          valid: false,
          errors: [{
            message: `JSON syntax error: ${jsonError}`,
            severity: 'error' as const,
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
          
          // Process errors based on options
          const processedErrors: ValidationError[] = [];
          const processedWarnings: ValidationError[] = [];
          
          for (const err of errors) {
            const isUndeclaredSymbol = this.isUndeclaredSymbolError(err);
            
            if (this.options.ignoreUndeclaredSymbols && isUndeclaredSymbol) {
              // Convert undeclaredSymbol errors to warnings
              processedWarnings.push({
                message: this.enhanceErrorMessage(err, content),
                range: undefined,
                severity: 'warning' as const,
                code: err.info?.codeAction?.title
              });
            } else {
              // Keep as error
              processedErrors.push({
                message: this.enhanceErrorMessage(err, content),
                range: undefined,
                severity: 'error' as const,
                code: err.info?.codeAction?.title
              });
            }
          }
          
          return {
            filePath,
            resourceId,
            valid: processedErrors.length === 0,
            errors: processedErrors,
            warnings: processedWarnings
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
      
    } catch (error) {
      return {
        filePath,
        resourceId,
        valid: false,
        errors: [{
          message: `Validation error: ${error}`,
          severity: 'error' as const
        }],
        warnings: []
      };
    }
  }

  async validateAllJsonFiles(datpackPath: string): Promise<ValidationResult[]> {
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

    const results: ValidationResult[] = [];
    
    for (const filePath of files) {
      const result = await this.validateJsonFile(filePath, datpackPath);
      results.push(result);
      
      if (this.verbose) {
        const status = result.valid ? '✓' : '✗';
        const displayPath = result.filePath.replace(process.cwd() + '/', '');
        console.log(`${status} ${displayPath}`);
        
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
    } catch (error) {
      return null;
    }
  }

  protected extractRegistryType(resourceId: string): string | null {
    // TODO: Extract from actual validation context
    return 'unknown';
  }

  /**
   * Check if an error is an undeclaredSymbol error that should be handled specially
   */
  private isUndeclaredSymbolError(error: core.LanguageError): boolean {
    const message = error.message;
    
    // Spyglass uses "(rule: undeclaredSymbol)" in error messages for undeclared symbol errors
    return message.includes('(rule: undeclaredSymbol)');
  }

  /**
   * Enhance error messages from Spyglass to provide more context
   */
  private enhanceErrorMessage(error: core.LanguageError, content: string): string {
    let message = error.message;
    
    // Handle "Expected nothing" errors by providing context about what was found
    if (message === 'Expected nothing') {
      const contextInfo = this.getErrorContext(error, content);
      if (contextInfo) {
        message = `Unexpected field or value: ${contextInfo}`;
      }
    }
    
    // Handle "Expected an int" and similar type errors
    if (message.startsWith('Expected a') || message.startsWith('Expected an')) {
      const contextInfo = this.getErrorContext(error, content);
      const lineNumber = this.getLineNumber(error, content);
      
      if (contextInfo) {
        message = `${message} for ${contextInfo}`;
      } else if (lineNumber) {
        message = `${message} (line ${lineNumber})`;
      }
    }
    
    // Add source context if available
    if (error.source && error.source !== 'mcdoc') {
      message = `[${error.source}] ${message}`;
    }
    
    return message;
  }

  /**
   * Extract context information from the error location in the JSON content
   */
  private getErrorContext(error: core.LanguageError, content: string): string | null {
    try {
      const startOffset = error.range.start;
      const endOffset = error.range.end;
      
      if (startOffset >= 0 && endOffset <= content.length) {
        // Get the problematic text
        let errorText = '';
        if (endOffset > startOffset) {
          errorText = content.substring(startOffset, endOffset);
        }
        
        // Find the JSON path to this error location
        const jsonPath = this.findJsonPath(content, startOffset);
        
        // Find the line containing the error for more context
        const beforeError = content.substring(0, Math.max(startOffset, endOffset));
        const lines = beforeError.split('\n');
        const currentLine = lines[lines.length - 1] || '';
        
        // Look for key names in the current line
        const keyMatch = currentLine.match(/"([^"]+)"\s*:/);
        if (keyMatch) {
          const keyName = keyMatch[1];
          // Avoid duplication if the json path already ends with the key name
          let pathContext = jsonPath;
          if (pathContext && !pathContext.endsWith(`.${keyName}`)) {
            pathContext = `${pathContext}.${keyName}`;
          } else if (!pathContext) {
            pathContext = keyName;
          }
          return `"${pathContext}" field`;
        }
        
        // Look for properties or objects that shouldn't be there
        const objectMatch = currentLine.match(/(\w+)\s*:\s*\{/);
        if (objectMatch) {
          const pathContext = jsonPath ? ` in ${jsonPath}` : '';
          return `"${objectMatch[1]}" object${pathContext} (not expected here)`;
        }
        
        // Look for array structures
        const arrayMatch = currentLine.match(/(\w+)\s*:\s*\[/);
        if (arrayMatch) {
          const pathContext = jsonPath ? ` in ${jsonPath}` : '';
          return `"${arrayMatch[1]}" array${pathContext} (not expected here)`;
        }
        
        // Fall back to showing the exact text if it's reasonable
        const trimmedError = errorText.trim();
        if (trimmedError.length > 0 && trimmedError.length < 100) {
          const pathContext = jsonPath ? ` in ${jsonPath}` : '';
          return `"${trimmedError}"${pathContext} (not valid here)`;
        }
        
        // If no specific context, try to extract the nearest field name
        const nearbyKeyMatch = beforeError.match(/"([^"]+)"\s*:\s*[^,}]*$/);
        if (nearbyKeyMatch) {
          const pathContext = jsonPath ? ` -> ${nearbyKeyMatch[1]}` : nearbyKeyMatch[1];
          return `content in "${pathContext}" field`;
        }
        
        // Last resort: just show the JSON path if we have one
        if (jsonPath) {
          return `invalid content in ${jsonPath}`;
        }
      }
    } catch (e) {
      // Ignore context extraction errors
    }
    
    return null;
  }

  /**
   * Get the line number for an error based on its character offset
   */
  private getLineNumber(error: core.LanguageError, content: string): number | null {
    try {
      const startOffset = error.range.start;
      if (startOffset >= 0 && startOffset <= content.length) {
        const beforeError = content.substring(0, startOffset);
        const lineNumber = beforeError.split('\n').length;
        return lineNumber;
      }
    } catch (e) {
      // Ignore line number calculation errors
    }
    return null;
  }

  /**
   * Find the JSON path (like "noise_router.initial_density") for a given character offset
   */
  private findJsonPath(content: string, offset: number): string | null {
    try {
      // Parse the JSON to build a proper path context
      const jsonObj = JSON.parse(content);
      const beforeError = content.substring(0, offset);
      
      // Count braces and find keys to build path
      const path: string[] = [];
      let braceCount = 0;
      let inString = false;
      let escapeNext = false;
      let keyStart = -1;
      
      for (let i = 0; i < beforeError.length; i++) {
        const char = beforeError[i];
        
        if (escapeNext) {
          escapeNext = false;
          continue;
        }
        
        if (char === '\\') {
          escapeNext = true;
          continue;
        }
        
        if (char === '"') {
          if (!inString) {
            keyStart = i + 1; // Start of key content
          } else if (keyStart !== -1) {
            // End of string - check if this is a key (followed by :)
            const key = beforeError.substring(keyStart, i);
            const afterQuote = beforeError.substring(i + 1).trim();
            if (afterQuote.startsWith(':')) {
              // This is a key
              if (braceCount === 0) {
                path[0] = key; // Root level key
              } else {
                path[braceCount - 1] = key; // Nested key
              }
            }
            keyStart = -1;
          }
          inString = !inString;
        }
        
        if (!inString) {
          if (char === '{') {
            braceCount++;
          } else if (char === '}') {
            braceCount--;
            if (braceCount >= 0) {
              path.length = braceCount; // Trim path to current nesting level
            }
          }
        }
      }
      
      // Clean up and return the path
      const cleanPath = path.filter(key => key && key.length > 0);
      return cleanPath.length > 0 ? cleanPath.join('.') : null;
    } catch (e) {
      // Fallback to simpler regex-based approach if JSON parsing fails
      try {
        const beforeError = content.substring(0, offset);
        const keys = [];
        const keyRegex = /"([^"]+)"\s*:/g;
        let match;
        
        while ((match = keyRegex.exec(beforeError)) !== null) {
          keys.push(match[1]);
        }
        
        // Return the last few keys as context (usually the most relevant)
        return keys.slice(-2).join('.');
      } catch (fallbackError) {
        return null;
      }
    }
  }

  async close(): Promise<void> {
    if (this.verbose) {
      console.log('Closing Spyglass validator...');
    }
    
    if (this.project) {
      await this.project.close();
      this.project = null;
    }
  }

  getValidatorName(): string {
    return 'Spyglass';
  }
}