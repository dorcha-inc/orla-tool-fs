# orla-tool-fs

A comprehensive file system operations tool for orla that provides reading, writing, listing, and managing files and directories.

---
[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat&logo=go)](https://golang.org/)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)
[![Go Report Card](https://goreportcard.com/badge/github.com/dorcha-inc/orla-tool-fs)](https://goreportcard.com/report/github.com/dorcha-inc/orla-tool-fs)
[![Build](https://github.com/dorcha-inc/orla-tool-fs/actions/workflows/build.yml/badge.svg)](https://github.com/dorcha-inc/orla-tool-fs/actions/workflows/build.yml)
[![Coverage](https://codecov.io/gh/dorcha-inc/orla-tool-fs/branch/main/graph/badge.svg)](https://codecov.io/gh/dorcha-inc/orla-tool-fs)
---

## Installation

Install from the orla registry:

```bash
orla install fs
```

This will install the latest version from the default registry. To install a specific version:

```bash
orla install --version v0.3.0
```

Or to use a custom registry:

```bash
orla install fs --registry https://github.com/dorcha-inc/orla-registry
```

## Usage

Once installed, the tool will be available as an MCP tool. It supports multiple file system operations through the `operation` parameter.

### Operations

#### `read` - Read file contents

Reads and returns the contents of a file.

**Required parameters:**
- `operation`: `"read"`
- `path`: Path to the file to read

**Example:**
```json
{
  "operation": "read",
  "path": "/path/to/file.txt"
}
```

**Response:**
```json
{
  "success": true,
  "content": "file contents here..."
}
```

#### `write` - Write file contents

Writes content to a file, optionally creating parent directories.

**Required parameters:**
- `operation`: `"write"`
- `path`: Path to the file to write
- `content`: Content to write to the file

**Optional parameters:**
- `create_dirs`: Create parent directories if they don't exist (default: `false`)

**Example:**
```json
{
  "operation": "write",
  "path": "/path/to/file.txt",
  "content": "Hello, world!",
  "create_dirs": true
}
```

#### `list` - List directory contents

Lists the contents of a directory.

**Required parameters:**
- `operation`: `"list"`
- `path`: Path to the directory to list

**Optional parameters:**
- `recursive`: List recursively (default: `false`)

**Example:**
```json
{
  "operation": "list",
  "path": "/path/to/directory",
  "recursive": false
}
```

#### `exists` - Check if path exists

Checks if a file or directory exists.

**Required parameters:**
- `operation`: `"exists"`
- `path`: Path to check

**Example:**
```json
{
  "operation": "exists",
  "path": "/path/to/file.txt"
}
```

#### `stat` - Get file/directory statistics

Returns detailed information about a file or directory.

**Required parameters:**
- `operation`: `"stat"`
- `path`: Path to the file or directory

**Example:**
```json
{
  "operation": "stat",
  "path": "/path/to/file.txt"
}
```

#### `mkdir` - Create directory

Creates a directory, optionally creating parent directories.

**Required parameters:**
- `operation`: `"mkdir"`
- `path`: Path to the directory to create

**Optional parameters:**
- `parents`: Create parent directories if they don't exist (default: `false`)

**Example:**
```json
{
  "operation": "mkdir",
  "path": "/path/to/directory",
  "parents": true
}
```

#### `rm` - Remove file or directory

Removes a file or directory.

**Required parameters:**
- `operation`: `"rm"`
- `path`: Path to remove

**Optional parameters:**
- `recursive`: Remove directories recursively (default: `false`)

**Example:**
```json
{
  "operation": "rm",
  "path": "/path/to/file.txt"
}
```

#### `mv` - Move or rename

Moves or renames a file or directory.

**Required parameters:**
- `operation`: `"mv"`
- `source`: Source path
- `dest`: Destination path

**Example:**
```json
{
  "operation": "mv",
  "source": "/path/to/old.txt",
  "dest": "/path/to/new.txt"
}
```

#### `cp` - Copy file or directory

Copies a file or directory.

**Required parameters:**
- `operation`: `"cp"`
- `source`: Source path
- `dest`: Destination path

**Optional parameters:**
- `recursive`: Copy directories recursively (default: `false`, required for directories)

**Example:**
```json
{
  "operation": "cp",
  "source": "/path/to/file.txt",
  "dest": "/path/to/copy.txt",
  "recursive": false
}
```

## Error Handling

All operations return a JSON response with a `success` field. If `success` is `false`, an `error` field will contain the error message.

**Example error response:**
```json
{
  "success": false,
  "error": "File not found: /path/to/file.txt"
}
```

## Development

To test locally before publishing:

1. Clone this repository
2. Create a git tag: `git tag v0.3.0`
3. Push the tag to GitHub
4. Ensure the tool is added to the registry's `registry.yaml` file
5. Test installation: `orla install fs`

**Note:** After installation, restart the orla server for the tool to be available.
