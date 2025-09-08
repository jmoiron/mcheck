import { readFileSync } from 'fs';
import { glob } from 'glob';
import * as core from '@spyglassmc/core';
import type { ParsedMcdocFile } from './mcdoc-parser.js';
import type { DatapackFileInfo } from './path-mapper.js';
import { DatapackPathMapper } from './path-mapper.js';
import { TypeGraph } from './type-graph.js';

export interface JsonValidationResult {
  filePath: string;
  fileInfo: DatapackFileInfo;
  valid: boolean;
  errors: ValidationError[];
  warnings: ValidationError[];
  expectedType?: string;
  actualContent?: any;
}

export interface ValidationError {
  message: string;
  path?: string;
  line?: number;
  column?: number;
  severity: 'error' | 'warning';
}

export class JsonValidator {
  private pathMapper: DatapackPathMapper;
  private schemas: Map<string, ParsedMcdocFile> = new Map();
  private typeGraph: TypeGraph;
  private verbose: boolean;

  constructor(parsedSchemas: ParsedMcdocFile[], verbose: boolean = false) {
    this.pathMapper = new DatapackPathMapper();
    this.verbose = verbose;
    this.typeGraph = new TypeGraph(parsedSchemas);
    
    // Index schemas for quick lookup
    for (const schema of parsedSchemas) {
      this.schemas.set(schema.path, schema);
    }
  }

  /**
   * Discover JSON files in a datapack directory
   */
  async discoverJsonFiles(datpackPath: string): Promise<string[]> {
    const pattern = `${datpackPath}/**/*.json`;
    const files = await glob(pattern, { 
      ignore: ['**/node_modules/**'],
      absolute: true 
    });
    
    // Filter out metadata files
    return files.filter(f => !f.includes('/_'));
  }

  /**
   * Validate a single JSON file against mcdoc schemas
   */
  async validateJsonFile(filePath: string, datpackRoot: string = './worldgen'): Promise<JsonValidationResult> {
    const fileInfo = this.pathMapper.parseDatapackPath(filePath, datpackRoot);
    
    if (!fileInfo) {
      return {
        filePath,
        fileInfo: {} as DatapackFileInfo,
        valid: false,
        errors: [{
          message: `Unable to parse datapack path: ${filePath}`,
          severity: 'error'
        }],
        warnings: []
      };
    }

    // Load and parse JSON content
    let jsonContent: any;
    try {
      const content = readFileSync(filePath, 'utf-8');
      jsonContent = JSON.parse(content);
    } catch (error) {
      return {
        filePath,
        fileInfo,
        valid: false,
        errors: [{
          message: `JSON parse error: ${error}`,
          severity: 'error'
        }],
        warnings: []
      };
    }

    // Find the appropriate schema using the type graph
    const expectedType = this.typeGraph.getExpectedTypeName(fileInfo.registryType);
    if (!expectedType) {
      return {
        filePath,
        fileInfo,
        valid: false,
        errors: [{
          message: `No schema mapping found for registry type: ${fileInfo.registryType}`,
          severity: 'error'
        }],
        warnings: [],
        expectedType,
        actualContent: jsonContent
      };
    }

    // Use type graph for validation
    const validationResult = this.performTypeGraphValidation(jsonContent, fileInfo, expectedType);

    return {
      filePath,
      fileInfo,
      valid: validationResult.valid,
      errors: validationResult.errors,
      warnings: validationResult.warnings,
      expectedType,
      actualContent: jsonContent
    };
  }

  /**
   * Validate multiple JSON files
   */
  async validateAllJsonFiles(datpackPath: string): Promise<JsonValidationResult[]> {
    const jsonFiles = await this.discoverJsonFiles(datpackPath);
    const results: JsonValidationResult[] = [];

    if (this.verbose) {
      console.log(`Found ${jsonFiles.length} JSON files for validation`);
    }

    for (const filePath of jsonFiles) {
      try {
        const result = await this.validateJsonFile(filePath, datpackPath);
        results.push(result);
        
        if (this.verbose) {
          const status = result.valid ? '✓' : '✗';
          console.log(`${status} ${result.fileInfo.resourceId || filePath}`);
          
          if (!result.valid && result.errors.length > 0) {
            for (const error of result.errors.slice(0, 2)) { // Show first 2 errors
              console.log(`  - ${error.message}`);
            }
            if (result.errors.length > 2) {
              console.log(`  - ... and ${result.errors.length - 2} more errors`);
            }
          }
        }
      } catch (error) {
        if (this.verbose) {
          console.error(`Failed to validate ${filePath}:`, error);
        }
        results.push({
          filePath,
          fileInfo: {} as DatapackFileInfo,
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
   * Perform validation using the type graph
   */
  private performTypeGraphValidation(jsonContent: any, fileInfo: DatapackFileInfo, expectedType: string): {
    valid: boolean;
    errors: ValidationError[];
    warnings: ValidationError[];
  } {
    const warnings: ValidationError[] = [];
    
    // Use type graph to validate
    const validationResult = this.typeGraph.isValidForType(jsonContent, expectedType);
    
    const errors: ValidationError[] = validationResult.errors.map(error => ({
      message: error,
      severity: 'error' as const
    }));

    // Add specific validation for known types
    if (expectedType === 'Biome' && validationResult.valid) {
      const biomeValidation = this.validateBiome(jsonContent);
      errors.push(...biomeValidation.errors);
      warnings.push(...biomeValidation.warnings);
    }

    if (expectedType === 'NoiseSettings' && validationResult.valid) {
      const noiseValidation = this.validateNoiseSettings(jsonContent);
      errors.push(...noiseValidation.errors);
      warnings.push(...noiseValidation.warnings);
    }

    return {
      valid: errors.length === 0,
      errors,
      warnings
    };
  }

  /**
   * Perform basic validation (fallback for specific types)
   */
  private performBasicValidation(jsonContent: any, fileInfo: DatapackFileInfo, expectedType: string): {
    valid: boolean;
    errors: ValidationError[];
    warnings: ValidationError[];
  } {
    const errors: ValidationError[] = [];
    const warnings: ValidationError[] = [];

    // Basic validation based on known structure for biomes
    if (expectedType === 'Biome') {
      return this.validateBiome(jsonContent);
    }

    // Basic validation for noise settings
    if (expectedType === 'NoiseSettings') {
      return this.validateNoiseSettings(jsonContent);
    }

    // For other types, just check if it's a valid object
    if (typeof jsonContent !== 'object' || jsonContent === null) {
      errors.push({
        message: `Expected object for ${expectedType}, got ${typeof jsonContent}`,
        severity: 'error'
      });
    }

    return {
      valid: errors.length === 0,
      errors,
      warnings
    };
  }

  /**
   * Validate biome structure based on mcdoc schema
   */
  private validateBiome(biome: any): {
    valid: boolean;
    errors: ValidationError[];
    warnings: ValidationError[];
  } {
    const errors: ValidationError[] = [];
    const warnings: ValidationError[] = [];

    // Required fields based on the mcdoc schema
    const requiredFields = ['temperature', 'downfall', 'effects', 'spawners', 'spawn_costs', 'carvers', 'features'];
    
    for (const field of requiredFields) {
      if (!(field in biome)) {
        errors.push({
          message: `Missing required field: ${field}`,
          path: field,
          severity: 'error'
        });
      }
    }

    // Validate temperature range (should be reasonable)
    if (typeof biome.temperature === 'number') {
      if (biome.temperature < -2.0 || biome.temperature > 2.0) {
        warnings.push({
          message: `Temperature ${biome.temperature} is outside typical range [-2.0, 2.0]`,
          path: 'temperature',
          severity: 'warning'
        });
      }
    } else if ('temperature' in biome) {
      errors.push({
        message: `Temperature must be a number, got ${typeof biome.temperature}`,
        path: 'temperature',
        severity: 'error'
      });
    }

    // Validate downfall range (should be between 0 and 1)
    if (typeof biome.downfall === 'number') {
      if (biome.downfall < 0.0 || biome.downfall > 1.0) {
        warnings.push({
          message: `Downfall ${biome.downfall} is outside typical range [0.0, 1.0]`,
          path: 'downfall',
          severity: 'warning'
        });
      }
    } else if ('downfall' in biome) {
      errors.push({
        message: `Downfall must be a number, got ${typeof biome.downfall}`,
        path: 'downfall',
        severity: 'error'
      });
    }

    // Validate effects structure
    if (biome.effects && typeof biome.effects === 'object') {
      const requiredEffects = ['sky_color', 'fog_color', 'water_color', 'water_fog_color'];
      for (const effect of requiredEffects) {
        if (!(effect in biome.effects)) {
          errors.push({
            message: `Missing required effect: ${effect}`,
            path: `effects.${effect}`,
            severity: 'error'
          });
        } else if (typeof biome.effects[effect] !== 'number') {
          errors.push({
            message: `Effect ${effect} must be a number (color), got ${typeof biome.effects[effect]}`,
            path: `effects.${effect}`,
            severity: 'error'
          });
        }
      }
    } else if ('effects' in biome) {
      errors.push({
        message: `Effects must be an object, got ${typeof biome.effects}`,
        path: 'effects',
        severity: 'error'
      });
    }

    // Validate spawners structure
    if (biome.spawners && typeof biome.spawners === 'object') {
      const validCategories = ['monster', 'creature', 'ambient', 'axolotls', 'underground_water_creature', 'water_creature', 'water_ambient', 'misc'];
      for (const [category, spawners] of Object.entries(biome.spawners)) {
        if (!validCategories.includes(category)) {
          warnings.push({
            message: `Unknown mob category: ${category}`,
            path: `spawners.${category}`,
            severity: 'warning'
          });
        }
        
        if (Array.isArray(spawners)) {
          for (let i = 0; i < spawners.length; i++) {
            const spawner = spawners[i];
            if (typeof spawner === 'object' && spawner !== null) {
              const requiredSpawnerFields = ['type', 'weight', 'minCount', 'maxCount'];
              for (const field of requiredSpawnerFields) {
                if (!(field in spawner)) {
                  errors.push({
                    message: `Missing required spawner field: ${field}`,
                    path: `spawners.${category}[${i}].${field}`,
                    severity: 'error'
                  });
                }
              }
            }
          }
        }
      }
    } else if ('spawners' in biome) {
      errors.push({
        message: `Spawners must be an object, got ${typeof biome.spawners}`,
        path: 'spawners',
        severity: 'error'
      });
    }

    return {
      valid: errors.length === 0,
      errors,
      warnings
    };
  }

  /**
   * Validate a noise settings structure
   */
  private validateNoiseSettings(noiseSettings: any): {
    valid: boolean;
    errors: ValidationError[];
    warnings: ValidationError[];
  } {
    const errors: ValidationError[] = [];
    const warnings: ValidationError[] = [];

    // Required fields for noise settings
    const requiredFields = [
      'sea_level',
      'disable_mob_generation', 
      'aquifers_enabled',
      'ore_veins_enabled',
      'default_block',
      'default_fluid',
      'noise',
      'noise_router',
      'surface_rule',
      'spawn_target'
    ];

    // Check for required fields
    for (const field of requiredFields) {
      if (!(field in noiseSettings)) {
        errors.push({
          message: `Missing required field: ${field}`,
          path: field,
          severity: 'error'
        });
      }
    }

    // Validate field types
    if ('sea_level' in noiseSettings && typeof noiseSettings.sea_level !== 'number') {
      errors.push({
        message: 'sea_level must be a number',
        path: 'sea_level',
        severity: 'error'
      });
    }

    if ('disable_mob_generation' in noiseSettings && typeof noiseSettings.disable_mob_generation !== 'boolean') {
      errors.push({
        message: 'disable_mob_generation must be a boolean',
        path: 'disable_mob_generation',
        severity: 'error'
      });
    }

    if ('aquifers_enabled' in noiseSettings && typeof noiseSettings.aquifers_enabled !== 'boolean') {
      errors.push({
        message: 'aquifers_enabled must be a boolean',
        path: 'aquifers_enabled',
        severity: 'error'
      });
    }

    if ('ore_veins_enabled' in noiseSettings && typeof noiseSettings.ore_veins_enabled !== 'boolean') {
      errors.push({
        message: 'ore_veins_enabled must be a boolean',
        path: 'ore_veins_enabled',
        severity: 'error'
      });
    }

    // Validate noise_router structure
    if ('noise_router' in noiseSettings) {
      const noiseRouter = noiseSettings.noise_router;
      if (typeof noiseRouter !== 'object') {
        errors.push({
          message: 'noise_router must be an object',
          path: 'noise_router',
          severity: 'error'
        });
      } else {
        const requiredRouterFields = [
          'barrier', 'fluid_level_floodedness', 'fluid_level_spread', 'lava',
          'temperature', 'vegetation', 'continents', 'erosion', 'depth', 'ridges',
          'initial_density_without_jaggedness', 'final_density', 'vein_toggle', 'vein_ridged', 'vein_gap'
        ];
        
        for (const field of requiredRouterFields) {
          if (!(field in noiseRouter)) {
            errors.push({
              message: `Missing required noise_router field: ${field}`,
              path: `noise_router.${field}`,
              severity: 'error'
            });
          }
        }
      }
    }

    // Validate spawn_target structure (should be an array)
    if ('spawn_target' in noiseSettings) {
      if (!Array.isArray(noiseSettings.spawn_target)) {
        errors.push({
          message: 'spawn_target must be an array',
          path: 'spawn_target',
          severity: 'error'
        });
      }
    }

    // Validate surface_rule structure
    if ('surface_rule' in noiseSettings) {
      const surfaceRule = noiseSettings.surface_rule;
      if (typeof surfaceRule !== 'object') {
        errors.push({
          message: 'surface_rule must be an object',
          path: 'surface_rule',
          severity: 'error'
        });
      } else if (!surfaceRule.type || typeof surfaceRule.type !== 'string') {
        errors.push({
          message: 'surface_rule must have a string type field',
          path: 'surface_rule.type',
          severity: 'error'
        });
      }
    }

    return {
      valid: errors.length === 0,
      errors,
      warnings
    };
  }

  /**
   * Get debug information about the type system
   */
  getTypeGraphDebugInfo() {
    return this.typeGraph.getDebugInfo();
  }

  /**
   * Generate a validation report
   */
  generateReport(results: JsonValidationResult[]): {
    totalFiles: number;
    validFiles: number;
    invalidFiles: number;
    totalErrors: number;
    totalWarnings: number;
    byRegistryType: Record<string, { total: number; valid: number; invalid: number }>;
  } {
    const report = {
      totalFiles: results.length,
      validFiles: results.filter(r => r.valid).length,
      invalidFiles: results.filter(r => !r.valid).length,
      totalErrors: results.reduce((sum, r) => sum + r.errors.length, 0),
      totalWarnings: results.reduce((sum, r) => sum + r.warnings.length, 0),
      byRegistryType: {} as Record<string, { total: number; valid: number; invalid: number }>
    };

    // Group by registry type
    for (const result of results) {
      const registryType = result.fileInfo.registryType || 'unknown';
      if (!report.byRegistryType[registryType]) {
        report.byRegistryType[registryType] = { total: 0, valid: 0, invalid: 0 };
      }
      
      report.byRegistryType[registryType].total++;
      if (result.valid) {
        report.byRegistryType[registryType].valid++;
      } else {
        report.byRegistryType[registryType].invalid++;
      }
    }

    return report;
  }
}