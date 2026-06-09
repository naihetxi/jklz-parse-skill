#!/usr/bin/env node
'use strict';

/**
 * Parse jklz-parse API SSE response
 *
 * Parses the Server-Sent Events (SSE) format response from the jklz-parse API
 * and extracts structured content.
 *
 * Usage:
 *   node parse-response.cjs < response.txt
 *   curl ... | node parse-response.cjs
 *
 * Output (JSON to stdout):
 *   {
 *     "content": "...",
 *     "toc": [["标题", 0], ...],
 *     "tables": [...],
 *     "job_id": "...",
 *     "file_id": "..."
 *   }
 */

const fs = require('fs');

// Parse SSE line
function parseSSELine(line) {
  if (!line.startsWith('data: ')) return null;

  const jsonStr = line.slice(6).trim();
  try {
    return JSON.parse(jsonStr);
  } catch (e) {
    return null;
  }
}

// Extract content from parse_return messages
function extractContent(messages) {
  const result = {
    content: '',
    html: '',
    toc: [],
    tables: [],
    job_id: null,
    file_id: null,
    file_name: null,
    slices: [],
    chunks: []
  };

  for (const msg of messages) {
    if (msg.code !== '200') continue;

    const data = msg.data || {};
    const type = data.type;
    const value = data.value;

    if (type === 'parse_return' && value) {
      if (value.content) result.content += value.content;
      if (value.html) result.html += value.html;
      if (value.toc && Array.isArray(value.toc)) result.toc = value.toc;
      if (value.table) {
        if (Array.isArray(value.table)) {
          result.tables.push(...value.table);
        } else {
          result.tables.push(value.table);
        }
      }
      if (value.slice) result.slices = value.slice;
      if (value.chunks) result.chunks = value.chunks;
      if (value.job_id) result.job_id = value.job_id;
      if (value.file_id) result.file_id = value.file_id;
      if (value.file_name) result.file_name = value.file_name;
    }
  }

  return result;
}

// Read from stdin
const lines = [];
const messages = [];

process.stdin.setEncoding('utf8');
process.stdin.on('data', (chunk) => {
  lines.push(...chunk.split('\n'));
});

process.stdin.on('end', () => {
  for (const line of lines) {
    const parsed = parseSSELine(line.trim());
    if (parsed) {
      messages.push(parsed);
    }
  }

  const result = extractContent(messages);

  // If no content but has job_id/file_id, fetch from API
  if ((!result.content || result.content === '') && result.job_id && result.file_id) {
    // Signal that fetch is needed
    result.fetch_needed = true;
  }

  console.log(JSON.stringify(result, null, 2));
});
