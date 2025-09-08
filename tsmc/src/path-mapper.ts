import { join, relative, dirname, basename } from 'path';

export interface DatapackFileInfo {
  /** The file path relative to the datapack root */
  relativePath: string;
  /** The namespace extracted from the path (e.g., 'minecraft') */
  namespace: string;
  /** The registry type (e.g., 'worldgen/biome') */
  registryType: string;
  /** The resource name (e.g., 'forest') */
  resourceName: string;
  /** The full resource identifier (e.g., 'minecraft:forest') */
  resourceId: string;
  /** The mcdoc dispatch key (e.g., 'minecraft:resource["worldgen/biome"]') */
  dispatchKey?: string;
}

export class DatapackPathMapper {
  /**
   * Parse a datapack file path to extract registry information
   */
  parseDatapackPath(filePath: string, datpackRoot: string = './worldgen'): DatapackFileInfo | null {
    const relativePath = relative(datpackRoot, filePath);
    
    // Split path into components
    const pathParts = relativePath.split('/').filter(part => part.length > 0);
    
    // For worldgen files, the structure is: registryType/resourceName.json
    // For data packs, it's: data/namespace/registryType/resourceName.json
    
    let namespace: string;
    let registryType: string;
    let resourceName: string;
    
    if (pathParts.length >= 2) {
      if (pathParts[0] === 'data' && pathParts.length >= 4) {
        // Full datapack structure: data/namespace/registryType/resourceName.json
        // But handle the case where namespace might actually be part of registry type
        const potentialNamespace = pathParts[1];
        const potentialRegistryType = pathParts.slice(2, -1).join('/');
        
        // Check if this looks like a worldgen registry type pattern
        if (potentialNamespace === 'worldgen' || potentialRegistryType.startsWith('worldgen/') || 
            (pathParts.length >= 3 && pathParts[2] === 'worldgen')) {
          // This is actually: data/worldgen/registryType/resourceName.json
          // Treat it as a simplified structure with default minecraft namespace
          namespace = 'minecraft';
          registryType = pathParts.slice(1, -1).join('/'); // Skip 'data', include everything else
          resourceName = basename(pathParts[pathParts.length - 1], '.json');
        } else {
          // True datapack structure: data/minecraft/worldgen/biome/file.json
          namespace = potentialNamespace;
          registryType = potentialRegistryType;
          resourceName = basename(pathParts[pathParts.length - 1], '.json');
        }
      } else if (pathParts.length >= 2) {
        // Simplified structure: registryType/resourceName.json (for our worldgen files)
        namespace = 'minecraft'; // Default namespace
        registryType = pathParts.slice(0, -1).join('/');
        // Add worldgen/ prefix if not already present
        if (!registryType.startsWith('worldgen/')) {
          registryType = `worldgen/${registryType}`;
        }
        resourceName = basename(pathParts[pathParts.length - 1], '.json');
      } else {
        return null;
      }
    } else {
      return null;
    }
    
    // Skip metadata files
    if (resourceName.startsWith('_')) {
      return null;
    }
    
    const resourceId = `${namespace}:${resourceName}`;
    
    return {
      relativePath,
      namespace,
      registryType,
      resourceName,
      resourceId,
      dispatchKey: this.getDispatchKey(registryType)
    };
  }
  
  /**
   * Get the mcdoc dispatch key for a registry type
   */
  private getDispatchKey(registryType: string): string | undefined {
    // Map registry types to their mcdoc dispatch keys
    const dispatchMap: Record<string, string> = {
      'worldgen/biome': 'minecraft:resource["worldgen/biome"]',
      'worldgen/configured_carver': 'minecraft:resource["worldgen/configured_carver"]',
      'worldgen/configured_feature': 'minecraft:resource["worldgen/configured_feature"]',
      'worldgen/placed_feature': 'minecraft:resource["worldgen/placed_feature"]',
      'worldgen/structure': 'minecraft:resource["worldgen/structure"]',
      'worldgen/structure_set': 'minecraft:resource["worldgen/structure_set"]',
      'worldgen/processor_list': 'minecraft:resource["worldgen/processor_list"]',
      'worldgen/template_pool': 'minecraft:resource["worldgen/template_pool"]',
      'worldgen/noise_settings': 'minecraft:resource["worldgen/noise_settings"]',
      'worldgen/density_function': 'minecraft:resource["worldgen/density_function"]',
      'worldgen/world_preset': 'minecraft:resource["worldgen/world_preset"]',
      'worldgen/noise': 'minecraft:resource["worldgen/noise"]',
      'worldgen/flat_level_generator_preset': 'minecraft:resource["worldgen/flat_level_generator_preset"]',
      'worldgen/multi_noise_biome_source_parameter_list': 'minecraft:resource["worldgen/multi_noise_biome_source_parameter_list"]',
      // Add more mappings as needed
    };
    
    return dispatchMap[registryType];
  }
  
  /**
   * Get the expected mcdoc type name for a registry type
   */
  getExpectedTypeName(registryType: string): string | undefined {
    // Map registry types to their expected struct names
    const typeMap: Record<string, string> = {
      'worldgen/biome': 'Biome',
      'worldgen/configured_carver': 'ConfiguredCarver',
      'worldgen/configured_feature': 'ConfiguredFeature',
      'worldgen/placed_feature': 'PlacedFeature',
      'worldgen/structure': 'Structure',
      'worldgen/structure_set': 'StructureSet',
      'worldgen/processor_list': 'ProcessorList',
      'worldgen/template_pool': 'TemplatePool',
      'worldgen/noise_settings': 'NoiseSettings',
      'worldgen/density_function': 'DensityFunction',
      'worldgen/world_preset': 'WorldPreset',
      'worldgen/noise': 'Noise',
      'worldgen/flat_level_generator_preset': 'FlatLevelGeneratorPreset',
      'worldgen/multi_noise_biome_source_parameter_list': 'MultiNoiseBiomeSourceParameterList',
      // Add more mappings as needed
    };
    
    return typeMap[registryType];
  }
}