# lks

`lks` is a terminal-based interactive CLI tool written in Go that helps you manage symlinks between directories effortlessly.

If you have a source directory containing executable scripts or binaries and want to safely and interactively link them to a system destination path (e.g., `/usr/local/bin` or `~/bin`), `lks` provides an intuitive Terminal UI (TUI) to selectively manage these links.

## ✨ Features
- **Interactive TUI**: Navigate and toggle symlinks easily using arrow keys and the `<space>` bar.
- **Tri-State Link Awareness**: Visually informs you about the current status of the destination file:
  - `●`: Safely linked to your target source.
  - `○`: Not linked (or safely removed).
  - `!`: **Path Mismatch** - A symlink exists with the same name, but it points to a *different* unexpected location.
- **Bulk Actions**: Select/Deselect all in one go using the `<a>` key, or reset to initial scan state with `<r>`.

## 📦 Installation

To install `lks` from source, ensure you have Go 1.22+ installed and run:

```bash
go install github.com/jjmean2/lks@latest
```

## 🚀 Usage

`lks` requires both a source directory containing your executables and a destination directory where the symlinks should be placed. 

Run `lks` using the `-s` (source) and `-d` (destination) flags:

```bash
lks -s /path/to/my/scripts -d /usr/local/bin
```

### Key Bindings

Once the TUI starts, you can use the following keys:
- `<up>` / `<down>` or `k` / `j`: Move the cursor up and down the list.
- `<space>`: Toggle the link state for the currently highlighted executable.
  - Toggling an unlinked item (`○`) will queue it to be linked (`●`).
  - Toggling an already linked item (`●`) will queue it to be removed (`○`).
  - Toggling a mismatched link (`!`) will queue it to forcefully overwrite with the correct link (`●`) -> remove (`○`) -> back to ignore (`!`).
- `<a>`: Toggle all items on or off.
- `<r>`: Reset all selections to their initial state when the program first started.
- `<enter>`: Confirm and effectively apply all changes.
- `<q>`, `Ctrl+C`, `Ctrl+D`: Abort and quit without making any changes.

## 🛠️ Options and Flags
```
A CLI tool to interactively manage symlinks between directories.

Usage: lks -s <src> -d <dest>

Options:
  -s, --source <path>
        Source directory containing executables
  -d, --destination <path>
        Destination directory for symlinks
  -v, --version
        Show version information
  -h, --help
        Show help message
```

## 📄 License
This project is licensed under the MIT License.
