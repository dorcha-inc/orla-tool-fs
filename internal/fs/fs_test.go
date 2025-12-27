// Package fs provides file system operations for the orla-tool-fs tool.
package fs

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExpandPath(t *testing.T) {
	tests := []struct {
		name    string
		path    string
		wantErr bool
		want    string
	}{
		{
			name:    "simple path",
			path:    "/tmp/test",
			wantErr: false,
			want:    "/tmp/test",
		},
		{
			name:    "path with ..",
			path:    "/tmp/../test",
			wantErr: false,
			want:    "/test",
		},
		{
			name:    "home directory",
			path:    "~/test",
			wantErr: false,
		},
		{
			name:    "just tilde",
			path:    "~",
			wantErr: false,
		},
		{
			name:    "environment variable",
			path:    "$HOME/test",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ExpandPath(tt.path)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			if tt.want != "" {
				assert.Equal(t, tt.want, got)
			} else if tt.path == "~/test" {
				home, err := os.UserHomeDir()
				require.NoError(t, err)
				want := filepath.Join(home, "test")
				assert.Equal(t, want, got)
			} else if tt.path == "~" {
				want, err := os.UserHomeDir()
				require.NoError(t, err)
				assert.Equal(t, want, got)
			} else if tt.path == "$HOME/test" {
				assert.NotEqual(t, "$HOME/test", got, "should expand $HOME")
			}
		})
	}
}

func TestRead(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")
	testContent := "Hello, World!"

	//nolint:gosec // Test file - write to test file
	require.NoError(t, os.WriteFile(testFile, []byte(testContent), 0644))

	tests := []struct {
		name    string
		path    string
		wantErr bool
		want    string
	}{
		{
			name:    "read existing file",
			path:    testFile,
			wantErr: false,
			want:    testContent,
		},
		{
			name:    "empty path",
			path:    "",
			wantErr: true,
		},
		{
			name:    "non-existent file",
			path:    filepath.Join(tmpDir, "nonexistent.txt"),
			wantErr: true,
		},
		{
			name:    "directory instead of file",
			path:    tmpDir,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Read(tt.path)
			if tt.wantErr {
				//nolint:errcheck // Type assertion in test is safe
				assert.False(t, result["success"].(bool))
				return
			}
			//nolint:errcheck // Type assertion in test is safe
			assert.True(t, result["success"].(bool))
			if tt.want != "" {
				assert.Equal(t, tt.want, result["content"])
			}
		})
	}
}

func TestWrite(t *testing.T) {
	tmpDir := t.TempDir()

	tests := []struct {
		name       string
		path       string
		content    string
		createDirs bool
		wantErr    bool
		want       string
	}{
		{
			name:       "write to new file",
			path:       filepath.Join(tmpDir, "newfile.txt"),
			content:    "test content",
			createDirs: false,
			wantErr:    false,
			want:       "test content",
		},
		{
			name:       "write with create dirs",
			path:       filepath.Join(tmpDir, "subdir", "file.txt"),
			content:    "nested content",
			createDirs: true,
			wantErr:    false,
			want:       "nested content",
		},
		{
			name:       "empty path",
			path:       "",
			content:    "content",
			createDirs: false,
			wantErr:    true,
		},
		{
			name:       "empty content",
			path:       filepath.Join(tmpDir, "empty.txt"),
			content:    "",
			createDirs: false,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Write(tt.path, tt.content, tt.createDirs)
			if tt.wantErr {
				//nolint:errcheck // Type assertion in test is safe
				assert.False(t, result["success"].(bool))
				return
			}
			//nolint:errcheck // Type assertion in test is safe
			assert.True(t, result["success"].(bool))
			if tt.want != "" {
				//nolint:errcheck // Type assertion in test is safe
				path, ok := result["path"].(string)
				require.True(t, ok, "path should be a string")
				//nolint:gosec // Test file paths are safe - constructed from test temp directories
				data, err := os.ReadFile(path)
				require.NoError(t, err)
				assert.Equal(t, tt.want, string(data))
			}
		})
	}
}

func TestList(t *testing.T) {
	tmpDir := t.TempDir()

	// Create test structure
	//nolint:gosec // Test file permissions are acceptable for temporary test files
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "file1.txt"), []byte("content1"), 0644))
	//nolint:gosec // Test file permissions are acceptable for temporary test files
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "file2.txt"), []byte("content2"), 0644))
	//nolint:gosec // Test directory permissions are acceptable for temporary test files
	require.NoError(t, os.MkdirAll(filepath.Join(tmpDir, "subdir"), 0755))
	//nolint:gosec // Test file permissions are acceptable for temporary test files
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "subdir", "file3.txt"), []byte("content3"), 0644))

	tests := []struct {
		name        string
		path        string
		recursive   bool
		wantErr     bool
		minItems    int
		hasRelative bool
	}{
		{
			name:        "list non-recursive",
			path:        tmpDir,
			recursive:   false,
			wantErr:     false,
			minItems:    3, // file1, file2, subdir
			hasRelative: false,
		},
		{
			name:        "list recursive",
			path:        tmpDir,
			recursive:   true,
			wantErr:     false,
			minItems:    4, // file1, file2, subdir, subdir/file3
			hasRelative: true,
		},
		{
			name:      "empty path",
			path:      "",
			recursive: false,
			wantErr:   true,
		},
		{
			name:      "non-existent directory",
			path:      filepath.Join(tmpDir, "nonexistent"),
			recursive: false,
			wantErr:   true,
		},
		{
			name:      "file instead of directory",
			path:      filepath.Join(tmpDir, "file1.txt"),
			recursive: false,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := List(tt.path, tt.recursive)
			if tt.wantErr {
				//nolint:errcheck // Type assertion in test is safe
				assert.False(t, result["success"].(bool))
				return
			}
			//nolint:errcheck // Type assertion in test is safe
			assert.True(t, result["success"].(bool))
			items, ok := result["items"].([]map[string]any)
			require.True(t, ok, "items should be []map[string]any")
			assert.GreaterOrEqual(t, len(items), tt.minItems)
			if tt.hasRelative {
				hasRelative := false
				for _, item := range items {
					if _, ok := item["relative"]; ok {
						hasRelative = true
						break
					}
				}
				assert.True(t, hasRelative, "recursive list should include relative paths")
			}
		})
	}
}

func TestExists(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")
	//nolint:gosec // Test file permissions are acceptable for temporary test files
	require.NoError(t, os.WriteFile(testFile, []byte("test"), 0644))

	tests := []struct {
		name       string
		path       string
		wantErr    bool
		wantExists bool
		isFile     *bool
		isDir      *bool
	}{
		{
			name:       "existing file",
			path:       testFile,
			wantErr:    false,
			wantExists: true,
			isFile:     boolPtr(true),
		},
		{
			name:       "existing directory",
			path:       tmpDir,
			wantErr:    false,
			wantExists: true,
			isDir:      boolPtr(true),
		},
		{
			name:       "non-existent path",
			path:       filepath.Join(tmpDir, "nonexistent"),
			wantErr:    false,
			wantExists: false,
		},
		{
			name:    "empty path",
			path:    "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Exists(tt.path)
			if tt.wantErr {
				//nolint:errcheck // Type assertion in test is safe
				assert.False(t, result["success"].(bool))
				return
			}
			//nolint:errcheck // Type assertion in test is safe
			assert.True(t, result["success"].(bool))
			//nolint:errcheck // Type assertion in test is safe
			assert.Equal(t, tt.wantExists, result["exists"].(bool))
			if tt.isFile != nil {
				//nolint:errcheck // Type assertion in test is safe
				assert.Equal(t, *tt.isFile, result["is_file"].(bool))
			}
			if tt.isDir != nil {
				//nolint:errcheck // Type assertion in test is safe
				assert.Equal(t, *tt.isDir, result["is_dir"].(bool))
			}
		})
	}
}

func TestStat(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")
	testContent := "test content"
	//nolint:gosec // Test file permissions are acceptable for temporary test files
	require.NoError(t, os.WriteFile(testFile, []byte(testContent), 0644))

	tests := []struct {
		name     string
		path     string
		wantErr  bool
		wantType string
		isFile   *bool
		isDir    *bool
		wantSize *int64
	}{
		{
			name:     "stat existing file",
			path:     testFile,
			wantErr:  false,
			wantType: "file",
			isFile:   boolPtr(true),
			wantSize: int64Ptr(int64(len(testContent))),
		},
		{
			name:     "stat existing directory",
			path:     tmpDir,
			wantErr:  false,
			wantType: "directory",
			isDir:    boolPtr(true),
		},
		{
			name:    "non-existent path",
			path:    filepath.Join(tmpDir, "nonexistent"),
			wantErr: true,
		},
		{
			name:    "empty path",
			path:    "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Stat(tt.path)
			if tt.wantErr {
				//nolint:errcheck // Type assertion in test is safe
				assert.False(t, result["success"].(bool))
				return
			}
			//nolint:errcheck // Type assertion in test is safe
			//nolint:errcheck // Type assertion in test is safe
			assert.True(t, result["success"].(bool))
			if tt.wantType != "" {
				assert.Equal(t, tt.wantType, result["type"])
			}
			if tt.isFile != nil {
				//nolint:errcheck // Type assertion in test is safe
				assert.Equal(t, *tt.isFile, result["is_file"].(bool))
			}
			if tt.isDir != nil {
				//nolint:errcheck // Type assertion in test is safe
				assert.Equal(t, *tt.isDir, result["is_dir"].(bool))
			}
			if tt.wantSize != nil {
				//nolint:errcheck // Type assertion in test is safe
				assert.Equal(t, *tt.wantSize, result["size"].(int64))
			}
		})
	}
}

func TestMkdir(t *testing.T) {
	tmpDir := t.TempDir()

	tests := []struct {
		name    string
		path    string
		parents bool
		wantErr bool
		check   func(t *testing.T, result map[string]any, path string)
	}{
		{
			name:    "create single directory",
			path:    filepath.Join(tmpDir, "newdir"),
			parents: false,
			wantErr: false,
			check: func(t *testing.T, result map[string]any, path string) {
				//nolint:errcheck // Type assertion in test is safe
				assert.True(t, result["success"].(bool))
				//nolint:errcheck // Type assertion in test is safe
				//nolint:gosec // Test file paths are safe - constructed from test temp directories
				info, err := os.Stat(result["path"].(string))
				require.NoError(t, err)
				assert.True(t, info.IsDir())
			},
		},
		{
			name:    "create nested directory with parents",
			path:    filepath.Join(tmpDir, "parent", "child", "grandchild"),
			parents: true,
			wantErr: false,
			check: func(t *testing.T, result map[string]any, path string) {
				//nolint:errcheck // Type assertion in test is safe
				assert.True(t, result["success"].(bool))
			},
		},
		{
			name:    "create nested directory without parents",
			path:    filepath.Join(tmpDir, "parent2", "child2"),
			parents: false,
			wantErr: true,
		},
		{
			name:    "directory already exists",
			path:    tmpDir,
			parents: false,
			wantErr: false,
			check: func(t *testing.T, result map[string]any, path string) {
				//nolint:errcheck // Type assertion in test is safe
				assert.True(t, result["success"].(bool))
				assert.Equal(t, "directory already exists", result["message"])
			},
		},
		{
			name:    "empty path",
			path:    "",
			parents: false,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Mkdir(tt.path, tt.parents)
			if tt.wantErr {
				//nolint:errcheck // Type assertion in test is safe
				assert.False(t, result["success"].(bool))
				return
			}
			if tt.check != nil {
				tt.check(t, result, tt.path)
			}
		})
	}
}

func TestRm(t *testing.T) {
	tmpDir := t.TempDir()

	tests := []struct {
		name      string
		setup     func() string
		recursive bool
		wantErr   bool
		check     func(t *testing.T, result map[string]any, path string)
	}{
		{
			name: "remove file",
			setup: func() string {
				testFile := filepath.Join(tmpDir, "file.txt")
				//nolint:gosec // Test file permissions are acceptable for temporary test files
				require.NoError(t, os.WriteFile(testFile, []byte("test"), 0644))
				return testFile
			},
			recursive: false,
			wantErr:   false,
			check: func(t *testing.T, result map[string]any, path string) {
				//nolint:errcheck // Type assertion in test is safe
				assert.True(t, result["success"].(bool))
				//nolint:gosec // Test file paths are safe - constructed from test temp directories
				_, err := os.Stat(path)
				assert.True(t, os.IsNotExist(err))
			},
		},
		{
			name: "remove empty directory",
			setup: func() string {
				testDir := filepath.Join(tmpDir, "emptydir")
				//nolint:gosec // Test directory permissions are acceptable for temporary test files
				require.NoError(t, os.Mkdir(testDir, 0755))
				return testDir
			},
			recursive: false,
			wantErr:   false,
			check: func(t *testing.T, result map[string]any, path string) {
				//nolint:errcheck // Type assertion in test is safe
				assert.True(t, result["success"].(bool))
			},
		},
		{
			name: "remove directory with files recursively",
			setup: func() string {
				testDir := filepath.Join(tmpDir, "dirwithfiles")
				//nolint:gosec // Test directory permissions are acceptable for temporary test files
				require.NoError(t, os.Mkdir(testDir, 0755))
				//nolint:gosec // Test file permissions are acceptable for temporary test files
				require.NoError(t, os.WriteFile(filepath.Join(testDir, "file.txt"), []byte("test"), 0644))
				return testDir
			},
			recursive: true,
			wantErr:   false,
			check: func(t *testing.T, result map[string]any, path string) {
				//nolint:errcheck // Type assertion in test is safe
				assert.True(t, result["success"].(bool))
			},
		},
		{
			name: "remove non-empty directory without recursive",
			setup: func() string {
				testDir := filepath.Join(tmpDir, "nonemptydir")
				//nolint:gosec // Test directory permissions are acceptable for temporary test files
				require.NoError(t, os.Mkdir(testDir, 0755))
				//nolint:gosec // Test file permissions are acceptable for temporary test files
				require.NoError(t, os.WriteFile(filepath.Join(testDir, "file.txt"), []byte("test"), 0644))
				return testDir
			},
			recursive: false,
			wantErr:   true,
		},
		{
			name: "remove non-existent path",
			setup: func() string {
				return filepath.Join(tmpDir, "nonexistent")
			},
			recursive: false,
			wantErr:   true,
		},
		{
			name: "empty path",
			setup: func() string {
				return ""
			},
			recursive: false,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := tt.setup()
			result := Rm(path, tt.recursive)
			if tt.wantErr {
				//nolint:errcheck // Type assertion in test is safe
				assert.False(t, result["success"].(bool))
				return
			}
			if tt.check != nil {
				tt.check(t, result, path)
			}
		})
	}
}

func TestMv(t *testing.T) {
	tmpDir := t.TempDir()

	tests := []struct {
		name    string
		setup   func() (string, string)
		wantErr bool
		check   func(t *testing.T, result map[string]any, source, dest string)
	}{
		{
			name: "move file",
			setup: func() (string, string) {
				source := filepath.Join(tmpDir, "source.txt")
				dest := filepath.Join(tmpDir, "dest.txt")
				//nolint:gosec // Test file permissions are acceptable for temporary test files
				require.NoError(t, os.WriteFile(source, []byte("test"), 0644))
				return source, dest
			},
			wantErr: false,
			check: func(t *testing.T, result map[string]any, source, dest string) {
				//nolint:errcheck // Type assertion in test is safe
				assert.True(t, result["success"].(bool))
				//nolint:gosec // Test file paths are safe - constructed from test temp directories
				_, err := os.Stat(source)
				assert.True(t, os.IsNotExist(err))
				//nolint:gosec // Test file paths are safe - constructed from test temp directories
				_, err = os.Stat(dest)
				assert.NoError(t, err)
			},
		},
		{
			name: "rename directory",
			setup: func() (string, string) {
				source := filepath.Join(tmpDir, "sourcedir")
				dest := filepath.Join(tmpDir, "destdir")
				//nolint:gosec // Test directory permissions are acceptable for temporary test files
				require.NoError(t, os.Mkdir(source, 0755))
				//nolint:gosec // Test file permissions are acceptable for temporary test files
				require.NoError(t, os.WriteFile(filepath.Join(source, "file.txt"), []byte("test"), 0644))
				return source, dest
			},
			wantErr: false,
			check: func(t *testing.T, result map[string]any, source, dest string) {
				//nolint:errcheck // Type assertion in test is safe
				assert.True(t, result["success"].(bool))
			},
		},
		{
			name: "empty source",
			setup: func() (string, string) {
				return "", filepath.Join(tmpDir, "dest.txt")
			},
			wantErr: true,
		},
		{
			name: "empty dest",
			setup: func() (string, string) {
				return filepath.Join(tmpDir, "source.txt"), ""
			},
			wantErr: true,
		},
		{
			name: "non-existent source",
			setup: func() (string, string) {
				return filepath.Join(tmpDir, "nonexistent.txt"), filepath.Join(tmpDir, "dest.txt")
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			source, dest := tt.setup()
			result := Mv(source, dest)
			if tt.wantErr {
				//nolint:errcheck // Type assertion in test is safe
				assert.False(t, result["success"].(bool))
				return
			}
			if tt.check != nil {
				tt.check(t, result, source, dest)
			}
		})
	}
}

func TestCp(t *testing.T) {
	tmpDir := t.TempDir()

	tests := []struct {
		name      string
		setup     func() (string, string)
		recursive bool
		wantErr   bool
		check     func(t *testing.T, result map[string]any, source, dest string)
	}{
		{
			name: "copy file",
			setup: func() (string, string) {
				source := filepath.Join(tmpDir, "source.txt")
				dest := filepath.Join(tmpDir, "dest.txt")
				//nolint:gosec // Test file permissions are acceptable for temporary test files
				require.NoError(t, os.WriteFile(source, []byte("test content"), 0644))
				return source, dest
			},
			recursive: false,
			wantErr:   false,
			check: func(t *testing.T, result map[string]any, source, dest string) {
				//nolint:errcheck // Type assertion in test is safe
				assert.True(t, result["success"].(bool))
				//nolint:gosec // Test file paths are safe - constructed from test temp directories
				_, err := os.Stat(source)
				assert.NoError(t, err)
				//nolint:gosec // Test file paths are safe - constructed from test temp directories
				_, err = os.Stat(dest)
				assert.NoError(t, err)
				//nolint:gosec // Test file paths are safe - constructed from test temp directories
				sourceData, err := os.ReadFile(source)
				require.NoError(t, err)
				//nolint:gosec // Test file paths are safe - constructed from test temp directories
				destData, err := os.ReadFile(dest)
				require.NoError(t, err)
				assert.Equal(t, sourceData, destData)
			},
		},
		{
			name: "copy directory recursively",
			setup: func() (string, string) {
				source := filepath.Join(tmpDir, "sourcedir")
				dest := filepath.Join(tmpDir, "destdir")
				//nolint:gosec // Test directory permissions are acceptable for temporary test files
				require.NoError(t, os.Mkdir(source, 0755))
				//nolint:gosec // Test file permissions are acceptable for temporary test files
				require.NoError(t, os.WriteFile(filepath.Join(source, "file1.txt"), []byte("content1"), 0644))
				//nolint:gosec // Test directory permissions are acceptable for temporary test files
				require.NoError(t, os.Mkdir(filepath.Join(source, "subdir"), 0755))
				//nolint:gosec // Test file permissions are acceptable for temporary test files
				require.NoError(t, os.WriteFile(filepath.Join(source, "subdir", "file2.txt"), []byte("content2"), 0644))
				return source, dest
			},
			recursive: true,
			wantErr:   false,
			check: func(t *testing.T, result map[string]any, source, dest string) {
				//nolint:errcheck // Type assertion in test is safe
				assert.True(t, result["success"].(bool))
				destFile := filepath.Join(dest, "subdir", "file2.txt")
				//nolint:gosec // Test file paths are safe - constructed from test temp directories
				_, err := os.Stat(destFile)
				assert.NoError(t, err)
			},
		},
		{
			name: "copy directory without recursive",
			setup: func() (string, string) {
				source := filepath.Join(tmpDir, "sourcedir2")
				dest := filepath.Join(tmpDir, "destdir2")
				//nolint:gosec // Test directory permissions are acceptable for temporary test files
				require.NoError(t, os.Mkdir(source, 0755))
				return source, dest
			},
			recursive: false,
			wantErr:   true,
		},
		{
			name: "empty source",
			setup: func() (string, string) {
				return "", filepath.Join(tmpDir, "dest.txt")
			},
			recursive: false,
			wantErr:   true,
		},
		{
			name: "empty dest",
			setup: func() (string, string) {
				return filepath.Join(tmpDir, "source.txt"), ""
			},
			recursive: false,
			wantErr:   true,
		},
		{
			name: "non-existent source",
			setup: func() (string, string) {
				return filepath.Join(tmpDir, "nonexistent.txt"), filepath.Join(tmpDir, "dest.txt")
			},
			recursive: false,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			source, dest := tt.setup()
			result := Cp(source, dest, tt.recursive)
			if tt.wantErr {
				//nolint:errcheck // Type assertion in test is safe
				assert.False(t, result["success"].(bool))
				return
			}
			if tt.check != nil {
				tt.check(t, result, source, dest)
			}
		})
	}
}

// Helper functions
func boolPtr(b bool) *bool {
	return &b
}

func int64Ptr(i int64) *int64 {
	return &i
}
