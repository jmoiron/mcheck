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
export declare class DatapackPathMapper {
    /**
     * Parse a datapack file path to extract registry information
     */
    parseDatapackPath(filePath: string, datpackRoot?: string): DatapackFileInfo | null;
    /**
     * Get the mcdoc dispatch key for a registry type
     */
    private getDispatchKey;
    /**
     * Get the expected mcdoc type name for a registry type
     */
    getExpectedTypeName(registryType: string): string | undefined;
}
