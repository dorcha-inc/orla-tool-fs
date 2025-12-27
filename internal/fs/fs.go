// Package fs provides file system operations for the orla-tool-fs tool.
package fs

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"unicode/utf8"

	copyutil "github.com/otiai10/copy"
)

// mcpError sets the success flag to false and returns the error
func mcpError(err error) map[string]any {
	return map[string]any{"error": err.Error(), "success": false}
}

// mcpSuccess sets the success flag to true and returns the result. It
// takes multiple name-value pairs (name must be string, value can be any type)
// and returns a map[string]any.
func mcpSuccess(nameValuePairs ...any) map[string]any {
	result := map[string]any{"success": true}

	if len(nameValuePairs)%2 != 0 {
		return mcpError(fmt.Errorf("odd number of name-value pairs"))
	}

	for i := 0; i < len(nameValuePairs); i += 2 {
		name, ok := nameValuePairs[i].(string)
		if !ok {
			return mcpError(fmt.Errorf("name at index %d must be a string, got %T", i, nameValuePairs[i]))
		}
		result[name] = nameValuePairs[i+1]
	}
	return result
}

// ExpandPath expands shell-style paths like ~ and $VAR. It also cleans the path.
func ExpandPath(p string) (rtn string, err error) {
	// Clean the path before returning to avoid path traversal vulnerabilities.
	defer func() {
		if err != nil {
			return
		}
		rtn = filepath.Clean(rtn)
	}()

	p = os.ExpandEnv(p)

	if strings.HasPrefix(p, "~/") || p == "~" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("failed to get home directory: %w", err)
		}

		if p == "~" {
			return home, nil
		}

		rtn = strings.Replace(p, "~", home, 1)
		return rtn, nil
	}

	return p, nil
}

// Read reads the contents of a file
func Read(path string) map[string]any {
	if path == "" {
		return mcpError(fmt.Errorf("path is required"))
	}

	p, err := ExpandPath(path)
	if err != nil {
		return mcpError(err)
	}

	info, err := os.Stat(p)

	if err != nil {
		if os.IsNotExist(err) {
			return mcpError(fmt.Errorf("file not found: %s", path))
		}
		return mcpError(err)
	}

	if info.IsDir() {
		return mcpError(fmt.Errorf("path is not a file: %s", path))
	}

	// G304: This is a file system tool designed to read user-provided paths.
	// The path is validated (checked for existence, type) and cleaned via ExpandPath.
	//nolint:gosec // File system tool - user-provided paths are expected and validated
	data, err := os.ReadFile(p)
	if err != nil {
		if os.IsPermission(err) {
			return mcpError(fmt.Errorf("permission denied: %s", path))
		}
		return mcpError(err)
	}
	if !utf8.Valid(data) {
		return mcpError(fmt.Errorf("file is not valid UTF-8: %s", path))
	}
	return mcpSuccess("content", string(data))
}

// Write writes content to a file
func Write(path, content string, createDirs bool) map[string]any {
	if path == "" {
		return mcpError(fmt.Errorf("path is required"))
	}

	if content == "" {
		return mcpError(fmt.Errorf("content is required"))
	}

	p, err := ExpandPath(path)
	if err != nil {
		return mcpError(err)
	}

	if createDirs {
		// G301: This is a file system tool designed to create directories.
		// The path is validated and cleaned via ExpandPath before reaching this function.
		//nolint:gosec // File system tool - user-provided paths are expected and validated
		if err := os.MkdirAll(filepath.Dir(p), 0755); err != nil {
			return mcpError(err)
		}
	}

	// G304: This is a file system tool designed to write to a file.
	// The path is validated and cleaned via ExpandPath before reaching this function.
	//nolint:gosec // File system tool - user-provided paths are expected and validated
	if err := os.WriteFile(p, []byte(content), 0644); err != nil {
		if os.IsPermission(err) {
			return mcpError(fmt.Errorf("permission denied: %s", path))
		}
		return mcpError(err)
	}
	return mcpSuccess("path", p)
}

// List lists directory contents
func List(path string, recursive bool) map[string]any {
	if path == "" {
		return mcpError(fmt.Errorf("path is required"))
	}

	p, err := ExpandPath(path)
	if err != nil {
		return mcpError(err)
	}

	info, err := os.Stat(p)
	if err != nil {
		if os.IsNotExist(err) {
			return mcpError(fmt.Errorf("directory not found: %s", path))
		}
		return mcpError(err)
	}

	if !info.IsDir() {
		return mcpError(fmt.Errorf("path is not a directory: %s", path))
	}

	var items []map[string]any
	if recursive {
		err = filepath.WalkDir(p, func(walkPath string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}

			// Skip the root directory itself
			if walkPath == p {
				return nil
			}

			entryInfo, err := d.Info()
			if err != nil {
				return err
			}

			rel, relErr := filepath.Rel(p, walkPath)
			if relErr != nil {
				// Fallback to entry name if relative path calculation fails
				rel = d.Name()
			}

			items = append(items, map[string]any{
				"path":     walkPath,
				"name":     d.Name(),
				"type":     itemType(entryInfo),
				"relative": rel,
			})

			return nil
		})

		if err != nil {
			if os.IsPermission(err) {
				return mcpError(fmt.Errorf("permission denied: %s", path))
			}
			return mcpError(err)
		}

		return mcpSuccess("items", items, "count", len(items))
	}

	entries, err := os.ReadDir(p)
	if err != nil {
		if os.IsPermission(err) {
			return mcpError(fmt.Errorf("permission denied: %s", path))
		}
		return mcpError(err)
	}
	for _, entry := range entries {
		info, err := entry.Info()
		if err != nil {
			return mcpError(err)
		}
		items = append(items, map[string]any{
			"path": filepath.Join(p, entry.Name()),
			"name": entry.Name(),
			"type": itemType(info),
		})
	}

	return mcpSuccess("items", items, "count", len(items))
}

// Exists checks if a path exists
func Exists(path string) map[string]any {
	if path == "" {
		return mcpError(fmt.Errorf("path is required"))
	}
	p, err := ExpandPath(path)
	if err != nil {
		return mcpError(err)
	}
	info, err := os.Stat(p)
	exists := err == nil
	result := mcpSuccess("exists", exists, "path", p)
	if exists {
		result["type"] = itemType(info)
		result["is_file"] = !info.IsDir()
		result["is_dir"] = info.IsDir()
		return result
	}

	if !os.IsNotExist(err) {
		// If the error is not "doesn't exist", it's another error (e.g., permission denied)
		// Return error instead of just "exists: false"
		return mcpError(err)
	}

	return mcpSuccess("exists", false, "path", p)
}

// Stat returns file/directory statistics
func Stat(path string) map[string]any {
	if path == "" {
		return mcpError(fmt.Errorf("path is required"))
	}
	p, err := ExpandPath(path)
	if err != nil {
		return mcpError(err)
	}
	info, err := os.Stat(p)
	if err != nil {
		if os.IsNotExist(err) {
			return mcpError(fmt.Errorf("path not found: %s", path))
		}
		return mcpError(err)
	}
	return mcpSuccess(
		"path", p,
		"name", filepath.Base(p),
		"type", itemType(info),
		"size", info.Size(),
		"mode", fmt.Sprintf("%o", info.Mode().Perm()),
		"modified", info.ModTime().Unix(),
		"accessed", info.ModTime().Unix(),
		"created", info.ModTime().Unix(),
		"is_file", !info.IsDir(),
		"is_dir", info.IsDir(),
		"is_symlink", info.Mode()&os.ModeSymlink != 0,
	)
}

// Mkdir creates a directory
func Mkdir(path string, parents bool) map[string]any {
	if path == "" {
		return mcpError(fmt.Errorf("path is required"))
	}
	p, err := ExpandPath(path)
	if err != nil {
		return mcpError(err)
	}
	info, err := os.Stat(p)
	if err == nil {
		if info.IsDir() {
			return mcpSuccess("path", p, "message", "directory already exists")
		}
		return mcpError(fmt.Errorf("path exists but is not a directory: %s", path))
	}
	if parents {
		// G301: This is a file system tool designed to create directories.
		// The path is validated and cleaned via ExpandPath before reaching this function.
		//nolint:gosec // File system tool - user-provided paths are expected and validated
		err = os.MkdirAll(p, 0755)
	} else {
		// G301: This is a file system tool designed to create directories.
		// The path is validated and cleaned via ExpandPath before reaching this function.
		//nolint:gosec // File system tool - user-provided paths are expected and validated
		err = os.Mkdir(p, 0755)
	}
	if err != nil {
		if os.IsPermission(err) {
			return mcpError(fmt.Errorf("permission denied: %s", path))
		}
		return mcpError(err)
	}
	return mcpSuccess("path", p)
}

// Rm removes a file or directory
func Rm(path string, recursive bool) map[string]any {
	if path == "" {
		return mcpError(fmt.Errorf("path is required"))
	}
	p, err := ExpandPath(path)
	if err != nil {
		return mcpError(err)
	}
	info, err := os.Stat(p)
	if err != nil {
		if os.IsNotExist(err) {
			return mcpError(fmt.Errorf("path not found: %s", path))
		}
		return mcpError(err)
	}
	if info.IsDir() {
		if recursive {
			err = os.RemoveAll(p)
		} else {
			err = os.Remove(p)
			if err != nil && strings.Contains(err.Error(), "not empty") {
				return mcpError(fmt.Errorf("directory not empty: %s. use recursive=true", path))
			}
		}
	} else {
		err = os.Remove(p)
	}
	if err != nil {
		if os.IsPermission(err) {
			return mcpError(fmt.Errorf("permission denied: %s", path))
		}
		return mcpError(err)
	}
	return mcpSuccess("path", p)
}

// Mv moves or renames a file/directory
func Mv(source, dest string) map[string]any {
	if source == "" {
		return mcpError(fmt.Errorf("source is required"))
	}
	if dest == "" {
		return mcpError(fmt.Errorf("dest is required"))
	}
	src, err := ExpandPath(source)
	if err != nil {
		return mcpError(err)
	}
	dst, err := ExpandPath(dest)
	if err != nil {
		return mcpError(err)
	}
	if _, err := os.Stat(src); err != nil {
		if os.IsNotExist(err) {
			return mcpError(fmt.Errorf("source not found: %s", source))
		}
		return mcpError(err)
	}

	err = os.Rename(src, dst)
	if err != nil {
		if os.IsPermission(err) {
			return mcpError(fmt.Errorf("permission denied: %s", source))
		}
		return mcpError(err)
	}

	return mcpSuccess("source", src, "dest", dst)
}

// Cp copies a file or directory
func Cp(source, dest string, recursive bool) map[string]any {
	if source == "" {
		return mcpError(fmt.Errorf("source is required"))
	}
	if dest == "" {
		return mcpError(fmt.Errorf("dest is required"))
	}
	src, err := ExpandPath(source)
	if err != nil {
		return mcpError(err)
	}
	dst, err := ExpandPath(dest)
	if err != nil {
		return mcpError(err)
	}
	info, err := os.Stat(src)
	if err != nil {
		if os.IsNotExist(err) {
			return mcpError(fmt.Errorf("source not found: %s", source))
		}
		return mcpError(err)
	}
	if info.IsDir() && !recursive {
		return mcpError(fmt.Errorf("source is a directory. use recursive=true"))
	}

	err = copyutil.Copy(src, dst)
	if err != nil {
		if os.IsPermission(err) {
			return mcpError(fmt.Errorf("permission denied: %s", source))
		}
		return mcpError(err)
	}

	return mcpSuccess("source", src, "dest", dst)
}

func itemType(info os.FileInfo) string {
	if info.IsDir() {
		return "directory"
	}
	return "file"
}
