#!/usr/bin/env python3
"""
File system tool for orla
Provides comprehensive file and directory operations
"""

import argparse
import json
import shutil
import sys
from argparse import Namespace
from pathlib import Path
from typing import Any, Dict


def read_file(path: str) -> Dict[str, Any]:
    """Read file contents"""
    file_path = Path(path)

    if not file_path.exists():
        return {"error": f"File not found: {path}", "success": False}

    if not file_path.is_file():
        return {"error": f"Path is not a file: {path}", "success": False}

    try:
        content = file_path.read_text(encoding="utf-8")
        return {"content": content, "success": True}
    except UnicodeDecodeError:
        return {
            "error": f"File contains binary data or invalid UTF-8: {path}",
            "success": False,
        }
    except PermissionError:
        return {"error": f"Permission denied reading file: {path}", "success": False}
    except Exception as e:
        return {"error": f"Failed to read file: {e}", "success": False}


def write_file(path: str, content: str, create_dirs: bool = False) -> Dict[str, Any]:
    """Write file contents"""
    file_path = Path(path)

    if create_dirs:
        file_path.parent.mkdir(parents=True, exist_ok=True)

    try:
        file_path.write_text(content, encoding="utf-8")
        return {"success": True, "path": str(file_path)}
    except PermissionError:
        return {"error": f"Permission denied writing file: {path}", "success": False}
    except Exception as e:
        return {"error": f"Failed to write file: {e}", "success": False}


def list_directory(path: str, recursive: bool = False) -> Dict[str, Any]:
    """List directory contents"""
    dir_path = Path(path)

    if not dir_path.exists():
        return {"error": f"Directory not found: {path}", "success": False}

    if not dir_path.is_dir():
        return {"error": f"Path is not a directory: {path}", "success": False}

    try:
        items = []
        if recursive:
            for item in dir_path.rglob("*"):
                items.append(
                    {
                        "path": str(item),
                        "name": item.name,
                        "type": "directory" if item.is_dir() else "file",
                        "relative": str(item.relative_to(dir_path)),
                    }
                )
        else:
            for item in sorted(dir_path.iterdir()):
                items.append(
                    {
                        "path": str(item),
                        "name": item.name,
                        "type": "directory" if item.is_dir() else "file",
                    }
                )

        return {"items": items, "count": len(items), "success": True}
    except PermissionError:
        return {
            "error": f"Permission denied reading directory: {path}",
            "success": False,
        }
    except Exception as e:
        return {"error": f"Failed to list directory: {e}", "success": False}


def path_exists(path: str) -> Dict[str, Any]:
    """Check if path exists"""
    file_path = Path(path)
    exists = file_path.exists()

    result = {"exists": exists, "path": path, "success": True}

    if exists:
        result["type"] = "directory" if file_path.is_dir() else "file"
        result["is_file"] = file_path.is_file()
        result["is_dir"] = file_path.is_dir()

    return result


def get_stat(path: str) -> Dict[str, Any]:
    """Get file/directory statistics"""
    file_path = Path(path)

    if not file_path.exists():
        return {"error": f"Path not found: {path}", "success": False}

    try:
        stat_info = file_path.stat()

        return {
            "path": str(file_path),
            "name": file_path.name,
            "type": "directory" if file_path.is_dir() else "file",
            "size": stat_info.st_size,
            "mode": oct(stat_info.st_mode)[-3:],
            "modified": stat_info.st_mtime,
            "accessed": stat_info.st_atime,
            "created": stat_info.st_ctime,
            "is_file": file_path.is_file(),
            "is_dir": file_path.is_dir(),
            "is_symlink": file_path.is_symlink(),
            "success": True,
        }
    except PermissionError:
        return {"error": f"Permission denied accessing path: {path}", "success": False}
    except Exception as e:
        return {"error": f"Failed to get stat: {e}", "success": False}


def make_directory(path: str, parents: bool = False) -> Dict[str, Any]:
    """Create directory"""
    dir_path = Path(path)

    if dir_path.exists():
        if dir_path.is_dir():
            return {
                "success": True,
                "path": str(dir_path),
                "message": "Directory already exists",
            }
        else:
            return {
                "error": f"Path exists but is not a directory: {path}",
                "success": False,
            }

    try:
        dir_path.mkdir(parents=parents, exist_ok=True)
        return {"success": True, "path": str(dir_path)}
    except PermissionError:
        return {
            "error": f"Permission denied creating directory: {path}",
            "success": False,
        }
    except Exception as e:
        return {"error": f"Failed to create directory: {e}", "success": False}


def remove_path(path: str, recursive: bool = False) -> Dict[str, Any]:
    """Remove file or directory"""
    file_path = Path(path)

    if not file_path.exists():
        return {"error": f"Path not found: {path}", "success": False}

    try:
        if file_path.is_dir():
            if recursive:
                shutil.rmtree(file_path)
            else:
                file_path.rmdir()
        else:
            file_path.unlink()

        return {"success": True, "path": str(file_path)}
    except PermissionError:
        return {"error": f"Permission denied removing path: {path}", "success": False}
    except OSError as e:
        if "Directory not empty" in str(e):
            return {
                "error": f"Directory not empty: {path}. Use recursive=true to remove.",
                "success": False,
            }
        return {"error": f"Failed to remove path: {e}", "success": False}
    except Exception as e:
        return {"error": f"Failed to remove path: {e}", "success": False}


def move_path(source: str, dest: str) -> Dict[str, Any]:
    """Move or rename file/directory"""
    source_path = Path(source)
    dest_path = Path(dest)

    if not source_path.exists():
        return {"error": f"Source path not found: {source}", "success": False}

    try:
        shutil.move(str(source_path), str(dest_path))
        return {"success": True, "source": str(source_path), "dest": str(dest_path)}
    except PermissionError:
        return {"error": f"Permission denied moving path: {source}", "success": False}
    except Exception as e:
        return {"error": f"Failed to move path: {e}", "success": False}


def copy_path(source: str, dest: str, recursive: bool = False) -> Dict[str, Any]:
    """Copy file or directory"""
    source_path = Path(source)
    dest_path = Path(dest)

    if not source_path.exists():
        return {"error": f"Source path not found: {source}", "success": False}

    try:
        if source_path.is_dir():
            if recursive:
                shutil.copytree(str(source_path), str(dest_path), dirs_exist_ok=True)
            else:
                return {
                    "error": "Source is a directory. Use recursive=true to copy.",
                    "success": False,
                }
        else:
            shutil.copy2(str(source_path), str(dest_path))

        return {"success": True, "source": str(source_path), "dest": str(dest_path)}
    except PermissionError:
        return {"error": f"Permission denied copying path: {source}", "success": False}
    except Exception as e:
        return {"error": f"Failed to copy path: {e}", "success": False}


# Operation wrapper functions that handle validation
def read_op(args: Namespace) -> Dict[str, Any]:
    """Read operation with validation"""
    if not args.path:
        return {"error": "path is required for read operation", "success": False}
    return read_file(args.path)


def write_op(args: Namespace) -> Dict[str, Any]:
    """Write operation with validation"""
    if not args.path:
        return {"error": "path is required for write operation", "success": False}
    if args.content is None:
        return {"error": "content is required for write operation", "success": False}
    return write_file(args.path, args.content, args.create_dirs)


def list_op(args: Namespace) -> Dict[str, Any]:
    """List operation with validation"""
    if not args.path:
        return {"error": "path is required for list operation", "success": False}
    return list_directory(args.path, args.recursive)


def exists_op(args: Namespace) -> Dict[str, Any]:
    """Exists operation with validation"""
    if not args.path:
        return {"error": "path is required for exists operation", "success": False}
    return path_exists(args.path)


def stat_op(args: Namespace) -> Dict[str, Any]:
    """Stat operation with validation"""
    if not args.path:
        return {"error": "path is required for stat operation", "success": False}
    return get_stat(args.path)


def mkdir_op(args: Namespace) -> Dict[str, Any]:
    """Mkdir operation with validation"""
    if not args.path:
        return {"error": "path is required for mkdir operation", "success": False}
    return make_directory(args.path, args.parents)


def rm_op(args: Namespace) -> Dict[str, Any]:
    """Remove operation with validation"""
    if not args.path:
        return {"error": "path is required for rm operation", "success": False}
    return remove_path(args.path, args.recursive)


def mv_op(args: Namespace) -> Dict[str, Any]:
    """Move operation with validation"""
    if not args.source:
        return {"error": "source is required for mv operation", "success": False}
    if not args.dest:
        return {"error": "dest is required for mv operation", "success": False}
    return move_path(args.source, args.dest)


def cp_op(args: Namespace) -> Dict[str, Any]:
    """Copy operation with validation"""
    if not args.source:
        return {"error": "source is required for cp operation", "success": False}
    if not args.dest:
        return {"error": "dest is required for cp operation", "success": False}
    return copy_path(args.source, args.dest, args.recursive)


def main():
    parser = argparse.ArgumentParser(description="File system operations tool")
    parser.add_argument(
        "--operation",
        required=True,
        choices=["read", "write", "list", "exists", "stat", "mkdir", "rm", "mv", "cp"],
        help="Operation to perform",
    )

    # Common arguments
    parser.add_argument("--path", help="Path to file or directory")
    parser.add_argument("--source", help="Source path (for mv, cp)")
    parser.add_argument("--dest", help="Destination path (for mv, cp)")
    parser.add_argument("--content", help="Content to write (for write)")
    parser.add_argument(
        "--recursive",
        action="store_true",
        help="Recursive operation (for list, rm, cp)",
    )
    parser.add_argument(
        "--parents", action="store_true", help="Create parent directories (for mkdir)"
    )
    parser.add_argument(
        "--create-dirs",
        action="store_true",
        help="Create parent directories (for write)",
    )

    args = parser.parse_args()

    operation_map = {
        "read": read_op,
        "write": write_op,
        "list": list_op,
        "exists": exists_op,
        "stat": stat_op,
        "mkdir": mkdir_op,
        "rm": rm_op,
        "mv": mv_op,
        "cp": cp_op,
    }

    operation_func = operation_map.get(args.operation)
    if operation_func is None:
        result = {"error": f"Unknown operation: {args.operation}", "success": False}
    else:
        try:
            result = operation_func(args)
        except Exception as e:
            result = {"error": f"Unexpected error: {e}", "success": False}

    # Output result as JSON
    print(json.dumps(result, indent=2))

    # Exit with error code if operation failed
    if not result.get("success", False):
        sys.exit(1)
    else:
        sys.exit(0)


if __name__ == "__main__":
    main()
