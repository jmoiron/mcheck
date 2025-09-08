#!/usr/bin/env node
import { Command } from 'commander';
import { McdocLoader } from './mcdoc-loader.js';
import { McdocParser } from './mcdoc-parser.js';
import { createValidator, getAvailableValidators, getDefaultValidator } from './validators/index.js';
const program = new Command();
program
    .name('tsmc')
    .description('TypeScript Minecraft datapack validator using mcdoc schemas')
    .version('1.0.0')
    .option('--schema-path <path>', 'Path to schema directory', './java');
program
    .command('parse-schemas')
    .description('Parse and validate all mcdoc schema files')
    .option('-v, --verbose', 'Enable verbose output')
    .action(async (options, command) => {
    try {
        const schemaPath = command.parent?.opts().schemaPath || './java';
        console.log('🔍 Starting mcdoc schema parsing...');
        console.log(`Schema directory: ${schemaPath}`);
        // Initialize loader and parser
        const loader = new McdocLoader(schemaPath, options.verbose);
        const parser = new McdocParser(options.verbose);
        // Load all mcdoc files
        console.log('\n📁 Loading mcdoc files...');
        const mcdocFiles = await loader.loadAllMcdocFiles();
        if (mcdocFiles.length === 0) {
            console.log('❌ No mcdoc files found!');
            process.exit(1);
        }
        // Parse all files
        console.log('\n⚡ Parsing mcdoc files...');
        const parsedFiles = await parser.parseAllMcdocFiles(mcdocFiles);
        // Generate report
        const report = parser.getParsingReport(parsedFiles);
        // Clean up
        await parser.close();
        console.log('\n📊 Parsing Report:');
        console.log(`  Total files: ${report.totalFiles}`);
        console.log(`  Successful: ${report.successfulFiles}`);
        console.log(`  With errors: ${report.filesWithErrors}`);
        console.log(`  Total errors: ${report.totalErrors}`);
        if (options.verbose && report.errorDetails.length > 0) {
            console.log('\n🐛 Error Details:');
            for (const detail of report.errorDetails) {
                console.log(`  ${detail.file}:`);
                for (const error of detail.errors) {
                    console.log(`    - ${error.message || error}`);
                }
            }
        }
        if (report.totalErrors === 0) {
            console.log('\n✅ All schema files parsed successfully!');
        }
        else {
            console.log(`\n⚠️  Completed with ${report.totalErrors} errors`);
            if (!options.verbose) {
                console.log('Use --verbose flag to see error details');
            }
            process.exit(1);
        }
    }
    catch (error) {
        console.error('❌ Error:', error);
        process.exit(1);
    }
});
program
    .command('info')
    .description('Show information about the schema directory')
    .action(async (options, command) => {
    try {
        const schemaPath = command.parent?.opts().schemaPath || './java';
        const loader = new McdocLoader(schemaPath);
        const files = await loader.discoverMcdocFiles();
        console.log(`📍 Schema root: ${loader.getSchemaRoot()}`);
        console.log(`📄 Found ${files.length} mcdoc files`);
        if (files.length > 0) {
            console.log('\n📝 Files:');
            const relativePaths = files.map(f => f.replace(loader.getSchemaRoot() + '/', ''));
            relativePaths.sort().forEach(path => {
                console.log(`  ${path}`);
            });
        }
    }
    catch (error) {
        console.error('❌ Error:', error);
        process.exit(1);
    }
});
program
    .command('validate')
    .description('Validate JSON datapack files against mcdoc schemas')
    .option('-d, --datapack <path>', 'Path to datapack directory', './worldgen')
    .option('-v, --verbose', 'Enable verbose output with detailed validation results')
    .option('--file <path>', 'Validate a single file instead of entire directory')
    .option('--validator <type>', `Validator to use (${getAvailableValidators().join(', ')})`, getDefaultValidator())
    .action(async (options, command) => {
    try {
        const schemaPath = command.parent?.opts().schemaPath || './java';
        const datpackPath = options.datapack || './worldgen';
        const validatorType = options.validator;
        // Validate validator type
        if (!getAvailableValidators().includes(validatorType)) {
            console.error(`❌ Invalid validator type: ${validatorType}`);
            console.error(`Available validators: ${getAvailableValidators().join(', ')}`);
            process.exit(1);
        }
        if (options.verbose) {
            console.log('🔍 Starting JSON validation...');
            console.log(`Schema directory: ${schemaPath}`);
            console.log(`Datapack directory: ${datpackPath}`);
            console.log(`Validator: ${validatorType}`);
        }
        // Create and initialize validator
        const validator = createValidator(validatorType, options.verbose);
        try {
            await validator.initialize(schemaPath, datpackPath);
            if (options.verbose) {
                console.log(`\n🔎 Validating JSON files using ${validator.getValidatorName()} validator...`);
            }
            let validationResults;
            if (options.file) {
                if (options.verbose) {
                    console.log(`Validating single file: ${options.file}`);
                }
                const result = await validator.validateJsonFile(options.file, datpackPath);
                validationResults = [result];
            }
            else {
                validationResults = await validator.validateAllJsonFiles(datpackPath);
            }
            // Generate report
            const report = validator.generateReport(validationResults);
            // Unix-like behavior: only show failures, silent on success
            if (report.totalErrors === 0) {
                // Silent success - no output for valid files
                if (options.verbose) {
                    console.log(`\n✅ All files are valid! (${report.totalFiles} files checked)`);
                    if (report.totalWarnings > 0) {
                        console.log(`⚠️  ${report.totalWarnings} warnings found`);
                    }
                }
            }
            else {
                // Show only the failed files and their errors
                for (const result of validationResults) {
                    if (!result.valid) {
                        console.log(`${result.resourceId}:`);
                        for (const error of result.errors) {
                            if (error.range) {
                                console.log(`  ${error.message} (line ${error.range.start.line + 1})`);
                            }
                            else {
                                console.log(`  ${error.message}`);
                            }
                        }
                    }
                }
                if (options.verbose) {
                    console.log(`\n📊 Validation Summary:`);
                    console.log(`  Total files: ${report.totalFiles}`);
                    console.log(`  Valid files: ${report.validFiles}`);
                    console.log(`  Invalid files: ${report.invalidFiles}`);
                    console.log(`  Total errors: ${report.totalErrors}`);
                    console.log(`  Total warnings: ${report.totalWarnings}`);
                }
                process.exit(1);
            }
        }
        finally {
            // Clean up
            await validator.close();
        }
    }
    catch (error) {
        console.error('❌ Error:', error);
        process.exit(1);
    }
});
program
    .command('list-validators')
    .description('List available validators')
    .action(() => {
    console.log('Available validators:');
    for (const validatorType of getAvailableValidators()) {
        const isDefault = validatorType === getDefaultValidator();
        const marker = isDefault ? ' (default)' : '';
        console.log(`  - ${validatorType}${marker}`);
    }
});
program.parse();
//# sourceMappingURL=index.js.map