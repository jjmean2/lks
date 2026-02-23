package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

var Version = "dev"

type item struct {
	name            string
	selected        bool
	initialSelected bool
}

type model struct {
	items     []item
	cursor    int
	sourceDir string
	destDir   string
	quitting  bool
	confirmed bool
}

func initialModel(sourceDir, destDir string, items []item) model {
	return model{
		items:     items,
		sourceDir: sourceDir,
		destDir:   destDir,
	}
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "ctrl+d", "q":
			m.quitting = true
			return m, tea.Quit
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.items)-1 {
				m.cursor++
			}
		case " ":
			if len(m.items) > 0 {
				m.items[m.cursor].selected = !m.items[m.cursor].selected
			}
		case "enter":
			m.confirmed = true
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m model) View() string {
	if m.quitting || m.confirmed {
		return ""
	}

	var b strings.Builder

	b.WriteString("Link CLI Tools\n")
	b.WriteString("Use ↑/↓ arrows to navigate, Space to toggle, Enter to confirm.\n")
	b.WriteString("Press q, Ctrl+C, or Ctrl+D to quit without saving.\n\n")

	if len(m.items) == 0 {
		b.WriteString("No executable files found in the source directory.\n")
		return b.String()
	}

	for i, it := range m.items {
		cursor := "  "
		if m.cursor == i {
			cursor = "> "
		}

		checked := "○"
		if it.selected {
			checked = "●"
		}

		b.WriteString(fmt.Sprintf("%s%s %s\n", cursor, checked, it.name))
	}

	return b.String()
}

func getExecutables(dir string) ([]string, error) {
	var execs []string
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	for _, entry := range entries {
		if !entry.IsDir() {
			info, err := entry.Info()
			if err == nil {
				// Check if file is executable (owner, group, or other)
				if info.Mode()&0111 != 0 {
					execs = append(execs, entry.Name())
				}
			}
		}
	}
	return execs, nil
}

func isLinkedToSource(destDir, sourceDir, filename string) bool {
	destPath := filepath.Join(destDir, filename)
	sourcePath := filepath.Join(sourceDir, filename)

	fileInfo, err := os.Lstat(destPath)
	if err != nil {
		return false
	}

	if fileInfo.Mode()&os.ModeSymlink != 0 {
		target, err := os.Readlink(destPath)
		if err == nil {
			// On some systems Readlink returns relative path, resolve it
			if !filepath.IsAbs(target) {
				target = filepath.Join(destDir, target)
			}
			absTarget, _ := filepath.Abs(target)
			absSource, _ := filepath.Abs(sourcePath)

			if absTarget == absSource {
				return true
			}
		}
	}
	return false
}

func applyChanges(sourceDir, destDir string, items []item) error {
	absSourceDir, err := filepath.Abs(sourceDir)
	if err != nil {
		return err
	}
	absDestDir, err := filepath.Abs(destDir)
	if err != nil {
		return err
	}

	var errorMessages []string

	for _, it := range items {
		if it.selected == it.initialSelected {
			continue // No change
		}

		destPath := filepath.Join(absDestDir, it.name)
		sourcePath := filepath.Join(absSourceDir, it.name)

		if it.selected {
			// Newly selected: create symlink
			fileInfo, err := os.Lstat(destPath)
			if err == nil {
				// Destination file exists.
				if fileInfo.Mode()&os.ModeSymlink == 0 {
					errorMessages = append(errorMessages, fmt.Sprintf("Warning: Skipped '%s' because destination exists and is not a symlink.", destPath))
					continue
				} else {
					// It's a symlink, so we can safely remove it to replace with our symlink
					os.Remove(destPath)
				}
			}

			err = os.Symlink(sourcePath, destPath)
			if err != nil {
				errorMessages = append(errorMessages, fmt.Sprintf("Error creating symlink for '%s': %v", it.name, err))
			} else {
				fmt.Printf("Created symlink: %s -> %s\n", destPath, sourcePath)
			}
		} else {
			// Newly unselected: remove symlink ONLY if it points to source
			if isLinkedToSource(absDestDir, absSourceDir, it.name) {
				err := os.Remove(destPath)
				if err != nil {
					errorMessages = append(errorMessages, fmt.Sprintf("Error removing symlink for '%s': %v", it.name, err))
				} else {
					fmt.Printf("Removed symlink: %s\n", destPath)
				}
			}
		}
	}

	if len(errorMessages) > 0 {
		fmt.Println(strings.Join(errorMessages, "\n"))
	} else {
		// Just to indicate completion if changes were made and no errors occurred and no explicit print statements printed
		// actually, we print "Created symlink:" and "Removed symlink:" above, which is transparent enough.
	}

	return nil
}

func main() {
	var source string
	var dest string
	var versionFlag bool

	flag.StringVar(&source, "source", "", "Source directory (can also use -s)")
	flag.StringVar(&source, "s", "", "Source directory (shorthand for --source)")
	flag.StringVar(&dest, "destination", "", "Destination directory (can also use -d)")
	flag.StringVar(&dest, "d", "", "Destination directory (shorthand for --destination)")
	flag.BoolVar(&versionFlag, "version", false, "Print version information (can also use -v)")
	flag.BoolVar(&versionFlag, "v", false, "Print version information (shorthand for --version)")

	// Go's flag package handles -h and --help automatically.

	// Change custom usage description
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  -s, --source string\n\tSource directory containing executables\n")
		fmt.Fprintf(os.Stderr, "  -d, --destination string\n\tDestination directory for symlinks\n")
		fmt.Fprintf(os.Stderr, "  -v, --version\n\tShow version information\n")
		fmt.Fprintf(os.Stderr, "  -h, --help\n\tShow help message\n")
	}

	flag.Parse()

	if versionFlag {
		fmt.Printf("lks version %s\n", Version)
		os.Exit(0)
	}

	if source == "" || dest == "" {
		fmt.Println("Error: Both source (-s, --source) and destination (-d, --destination) are required.")
		flag.Usage()
		os.Exit(1)
	}

	absSourceDir, err := filepath.Abs(source)
	if err != nil {
		fmt.Printf("Error resolving source directory: %v\n", err)
		os.Exit(1)
	}
	absDestDir, err := filepath.Abs(dest)
	if err != nil {
		fmt.Printf("Error resolving destination directory: %v\n", err)
		os.Exit(1)
	}

	// Read executables from source
	executables, err := getExecutables(absSourceDir)
	if err != nil {
		fmt.Printf("Error reading source directory: %v\n", err)
		os.Exit(1)
	}

	// Prepare items
	var items []item
	for _, exe := range executables {
		selected := isLinkedToSource(absDestDir, absSourceDir, exe)
		items = append(items, item{
			name:            exe,
			selected:        selected,
			initialSelected: selected,
		})
	}

	p := tea.NewProgram(initialModel(absSourceDir, absDestDir, items))
	m, err := p.Run()
	if err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}

	if finalModel, ok := m.(model); ok {
		if finalModel.confirmed {
			// Print a newline to separate TUI from execution logs
			fmt.Println("\nApplying changes...")
			err := applyChanges(absSourceDir, absDestDir, finalModel.items)
			if err != nil {
				fmt.Printf("Error applying changes: %v\n", err)
			} else {
				fmt.Println("Done.")
			}
		} else if finalModel.quitting {
			fmt.Println("\nExited without making changes.")
		}
	}
}
