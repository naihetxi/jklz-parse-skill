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

function parseJSONObjects(input) {
  const messages = [];
  let buf = '';
  let inString = false;
  let escaped = false;
  let depth = 0;
  let started = false;

  for (const ch of input) {
    if (!started) {
      if (ch === '{') {
        started = true;
        depth = 1;
        buf = ch;
      }
      continue;
    }

    buf += ch;

    if (escaped) {
      escaped = false;
      continue;
    }
    if (ch === '\\') {
      escaped = true;
      continue;
    }
    if (ch === '"') {
      inString = !inString;
      continue;
    }
    if (inString) continue;

    if (ch === '{') depth += 1;
    if (ch === '}') depth -= 1;

    if (depth === 0) {
      try {
        messages.push(JSON.parse(buf));
      } catch (e) {
        // Ignore malformed fragments in streamed output.
      }
      buf = '';
      started = false;
    }
  }

  return messages;
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

    if ((type === 'parse_return' || type === 'parseReturn') && value) {
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
      if (value.job_id || value.jobId) result.job_id = value.job_id || value.jobId;
      if (value.file_id || value.fileId) result.file_id = value.file_id || value.fileId;
      if (value.file_name || value.fileName) result.file_name = value.file_name || value.fileName;
    }
  }

  return result;
}

// Read from stdin
const chunks = [];

process.stdin.setEncoding('utf8');
process.stdin.on('data', (chunk) => {
  chunks.push(chunk);
});

process.stdin.on('end', () => {
  const messages = parseJSONObjects(chunks.join(''));
  const result = extractContent(messages);

  // If no content but has job_id/file_id, fetch from API
  if ((!result.content || result.content === '') && result.job_id && result.file_id) {
    // Signal that fetch is needed
    result.fetch_needed = true;
  }

  console.log(JSON.stringify(result, null, 2));
});
