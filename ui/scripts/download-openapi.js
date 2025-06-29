#!/usr/bin/env node

import fs from 'fs';
import path from 'path';
import { fileURLToPath } from 'url';

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);

const API_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8888';
const OUTPUT_PATH = path.join(__dirname, '..', 'openapi.json');

async function downloadOpenAPISchema() {
  try {
    console.log(`Downloading OpenAPI schema from ${API_URL}/openapi.json...`);
    
    const response = await fetch(`${API_URL}/openapi.json`);
    
    if (!response.ok) {
      throw new Error(`Failed to download schema: ${response.status} ${response.statusText}`);
    }
    
    const schema = await response.json();
    
    // Write the schema to file
    fs.writeFileSync(OUTPUT_PATH, JSON.stringify(schema, null, 2));
    
    console.log(`✅ OpenAPI schema downloaded to ${OUTPUT_PATH}`);
    console.log(`📊 Schema contains ${Object.keys(schema.paths || {}).length} endpoints`);
    
  } catch (error) {
    console.error('❌ Failed to download OpenAPI schema:', error.message);
    console.log('\nMake sure your HashPost server is running:');
    console.log('  make dev');
    console.log('\nOr run the server directly:');
    console.log('  go run ./cmd/server');
    process.exit(1);
  }
}

downloadOpenAPISchema(); 