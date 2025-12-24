# orla-tool-file-read

A file reading tool for orla that reads and returns the contents of a file.

## Installation

Install from the orla registry:

```bash
orla install readfile
```

This will install the latest version from the default registry. To install a specific version:

```bash
orla install readfile@v0.1.0
```

Or to use a custom registry:

```bash
orla install readfile --registry https://github.com/dorcha-inc/orla-registry
```

## Usage

Once installed, the tool will be available as an MCP tool. It takes a `path` argument and returns the file contents.

Arguments:

- `path` (required): Path to the file to read

Example:

```json
{
  "path": "/path/to/file.txt"
}
```

The tool will:
- Read the file contents
- Return the contents as text (UTF-8)
- Handle errors gracefully (file not found, permission denied, etc.)

**Note:** After installation, restart the orla server for the tool to be available.

## Development

To test locally before publishing:

1. Clone this repository
2. Create a git tag: `git tag v0.1.0`
3. Push the tag to GitHub
4. Ensure the tool is added to the registry's `registry.yaml` file
5. Test installation: `orla install readfile`

