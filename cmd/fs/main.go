package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"

	"github.com/dorcha-inc/orla-tool-fs/internal/fs"
	"github.com/spf13/cobra"
)

func mcpFatalError(err error) {
	mcpOutput(map[string]any{"error": err.Error(), "success": false})
	os.Exit(1)
}

func getFlagOrFatal(cmd *cobra.Command, flag string) string {
	// Get flag from root persistent flags (where MCP sets them)
	value, err := cmd.Root().PersistentFlags().GetString(flag)
	if err != nil {
		mcpFatalError(fmt.Errorf("failed to get flag %s: %w", flag, err))
	}
	return value
}

func main() {
	rootCmd := &cobra.Command{
		Use:   "fs",
		Short: "File system operations tool",
		Long:  "A comprehensive file system operations tool for orla",
	}

	var operation string
	rootCmd.PersistentFlags().StringVar(&operation, "operation", "", "Operation: read, write, list, exists, stat, mkdir, rm, mv, cp")

	rootCmd.PersistentFlags().String("path", "", "Path to file or directory")
	rootCmd.PersistentFlags().String("source", "", "Source path (for mv, cp)")
	rootCmd.PersistentFlags().String("dest", "", "Destination path (for mv, cp)")
	rootCmd.PersistentFlags().String("content", "", "Content to write")
	rootCmd.PersistentFlags().String("recursive", "false", "Recursive operation")
	rootCmd.PersistentFlags().String("parents", "false", "Create parent directories")
	rootCmd.PersistentFlags().String("create-dirs", "false", "Create parent directories")

	// Add subcommands for each operation
	subCommandMap := map[string]*cobra.Command{
		"read":   newReadCmd(),
		"write":  newWriteCmd(),
		"list":   newListCmd(),
		"exists": newExistsCmd(),
		"stat":   newStatCmd(),
		"mkdir":  newMkdirCmd(),
		"rm":     newRmCmd(),
		"mv":     newMvCmd(),
		"cp":     newCpCmd(),
	}

	for _, cmd := range subCommandMap {
		rootCmd.AddCommand(cmd)
	}

	rootCmd.RunE = func(cmd *cobra.Command, args []string) error {
		if operation != "" {
			subCmd, ok := subCommandMap[operation]
			if !ok {
				return fmt.Errorf("unknown operation: %s", operation)
			}
			return subCmd.RunE(subCmd, args)
		}
		return cmd.Help()
	}

	err := rootCmd.Execute()
	if err != nil {
		mcpFatalError(err)
	}
}

func newReadCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "read",
		Short: "Read file contents",
		RunE: func(cmd *cobra.Command, args []string) error {
			// Get path from root persistent flags
			path := getFlagOrFatal(cmd, "path")
			result := fs.Read(path)
			mcpOutput(result)
			if !getBool(result, "success") {
				os.Exit(1)
			}
			return nil
		},
	}
	// Flags are inherited from root persistent flags
	return cmd
}

func newWriteCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "write",
		Short: "Write file contents",
		RunE: func(cmd *cobra.Command, args []string) error {
			path := getFlagOrFatal(cmd, "path")
			content := getFlagOrFatal(cmd, "content")
			createDirsStr := getFlagOrFatal(cmd, "create-dirs")
			createDirs := toBool(createDirsStr)

			result := fs.Write(path, content, createDirs)

			mcpOutput(result)
			if !getBool(result, "success") {
				os.Exit(1)
			}

			return nil
		},
	}
	return cmd
}

func newListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List directory contents",
		RunE: func(cmd *cobra.Command, args []string) error {
			path := getFlagOrFatal(cmd, "path")
			recursiveStr := getFlagOrFatal(cmd, "recursive")
			recursive := toBool(recursiveStr)
			result := fs.List(path, recursive)

			mcpOutput(result)
			if !getBool(result, "success") {
				os.Exit(1)
			}
			return nil
		},
	}
	return cmd
}

func newExistsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "exists",
		Short: "Check if path exists",
		RunE: func(cmd *cobra.Command, args []string) error {
			path := getFlagOrFatal(cmd, "path")
			result := fs.Exists(path)
			mcpOutput(result)
			return nil
		},
	}
	return cmd
}

func newStatCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "stat",
		Short: "Get file/directory statistics",
		RunE: func(cmd *cobra.Command, args []string) error {
			path := getFlagOrFatal(cmd, "path")
			result := fs.Stat(path)
			mcpOutput(result)
			if !getBool(result, "success") {
				os.Exit(1)
			}
			return nil
		},
	}
	return cmd
}

func newMkdirCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "mkdir",
		Short: "Create directory",
		RunE: func(cmd *cobra.Command, args []string) error {
			path := getFlagOrFatal(cmd, "path")
			parentsStr := getFlagOrFatal(cmd, "parents")
			parents := toBool(parentsStr)
			result := fs.Mkdir(path, parents)
			mcpOutput(result)
			if !getBool(result, "success") {
				os.Exit(1)
			}
			return nil
		},
	}
	// Flags are inherited from root persistent flags
	return cmd
}

func newRmCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "rm",
		Short: "Remove file or directory",
		RunE: func(cmd *cobra.Command, args []string) error {
			path := getFlagOrFatal(cmd, "path")
			recursiveStr := getFlagOrFatal(cmd, "recursive")
			recursive := toBool(recursiveStr)
			result := fs.Rm(path, recursive)
			mcpOutput(result)
			if !getBool(result, "success") {
				os.Exit(1)
			}
			return nil
		},
	}
	return cmd
}

func newMvCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "mv",
		Short: "Move or rename file/directory",
		RunE: func(cmd *cobra.Command, args []string) error {
			source := getFlagOrFatal(cmd, "source")
			dest := getFlagOrFatal(cmd, "dest")
			result := fs.Mv(source, dest)
			mcpOutput(result)
			if !getBool(result, "success") {
				os.Exit(1)
			}
			return nil
		},
	}
	return cmd
}

func newCpCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cp",
		Short: "Copy file or directory",
		RunE: func(cmd *cobra.Command, args []string) error {
			source := getFlagOrFatal(cmd, "source")
			dest := getFlagOrFatal(cmd, "dest")
			recursiveStr := getFlagOrFatal(cmd, "recursive")
			recursive := toBool(recursiveStr)
			result := fs.Cp(source, dest, recursive)
			mcpOutput(result)
			if !getBool(result, "success") {
				os.Exit(1)
			}
			return nil
		},
	}
	return cmd
}

func toBool(s string) bool {
	b, err := strconv.ParseBool(s)
	if err != nil {
		mcpFatalError(fmt.Errorf("failed to parse bool %s: %w", s, err))
	}
	return b
}

func getBool(m map[string]any, key string) bool {
	v, ok := m[key].(bool)
	if !ok {
		mcpFatalError(fmt.Errorf("value %s is not a boolean: %v", key, v))
	}
	return ok && v
}

func mcpOutput(v any) {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	err := enc.Encode(v)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error encoding output: %v\n", err)
		os.Exit(1)
	}
}
