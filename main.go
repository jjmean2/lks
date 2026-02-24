package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"runtime/debug"
	"sort"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	Version  = "dev"                                                            // Can still be overridden by ldflags
	keyStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("73"))             // Softer Cyan
	redStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("167")).Bold(true) // Muted Red
)

func getVersion() string {
	if Version != "dev" {
		return Version
	}

	info, ok := debug.ReadBuildInfo()
	if !ok {
		return "unknown"
	}

	// If installed via 'go install github.com/user/repo@version', info.Main.Version has the version tag
	if info.Main.Version != "" && info.Main.Version != "(devel)" {
		return info.Main.Version
	}

	// If built locally (Go 1.18+), extract vcs revision
	var revision string
	var modified bool
	for _, setting := range info.Settings {
		switch setting.Key {
		case "vcs.revision":
			revision = setting.Value
		case "vcs.modified":
			modified = setting.Value == "true"
		}
	}

	if revision != "" {
		if len(revision) > 7 {
			revision = revision[:7]
		}
		if modified {
			return revision + "-dirty"
		}
		return revision
	}

	return "dev"
}

type StringSliceFlag []string

func (i *StringSliceFlag) String() string {
	return strings.Join(*i, ", ")
}

func (i *StringSliceFlag) Set(value string) error {
	*i = append(*i, value)
	return nil
}

type LinkState int

const (
	StateNone LinkState = iota
	StateLinked
	StateInvalid
)

type item struct {
	// Common
	name string

	// Normal items
	isNormal     bool
	sourceDir    string
	state        LinkState
	initialState LinkState
	target       string

	// Radio items
	isRadio         bool
	radioGroup      string // The name of the executable
	radioValue      string // The path it points to, or "none"
	radioLabel      string // The text to display
	selected        bool
	initialSelected bool
}

type model struct {
	items      []item
	cursor     int
	sourceDirs []string
	destDir    string
	quitting   bool
	confirmed  bool
}

func initialModel(sourceDirs []string, destDir string, items []item) model {
	return model{
		items:      items,
		sourceDirs: sourceDirs,
		destDir:    destDir,
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
				it := &m.items[m.cursor]
				if it.isNormal {
					currentState := it.state
					initialState := it.initialState

					if initialState == StateInvalid {
						// Toggle sequence: ! -> ● -> ○ -> !
						if currentState == StateInvalid {
							it.state = StateLinked
						} else if currentState == StateLinked {
							it.state = StateNone
						} else if currentState == StateNone {
							it.state = StateInvalid
						}
					} else {
						// Toggle sequence: ● -> ○ -> ●
						if currentState == StateLinked {
							it.state = StateNone
						} else if currentState == StateNone {
							it.state = StateLinked
						}
					}
				} else if it.isRadio {
					if !it.selected {
						group := it.radioGroup
						for i := range m.items {
							if m.items[i].isRadio && m.items[i].radioGroup == group {
								m.items[i].selected = (i == m.cursor)
							}
						}
					}
				}
			}
		case "a":
			if len(m.items) > 0 {
				allLinked := true
				for _, it := range m.items {
					if it.isNormal && it.state != StateLinked {
						allLinked = false
						break
					}
				}
				for i := range m.items {
					if m.items[i].isNormal {
						if allLinked {
							m.items[i].state = StateNone
						} else {
							m.items[i].state = StateLinked
						}
					}
				}
			}
		case "r":
			if len(m.items) > 0 {
				it := &m.items[m.cursor]
				if it.isNormal {
					it.state = it.initialState
				} else if it.isRadio {
					group := it.radioGroup
					for i := range m.items {
						if m.items[i].isRadio && m.items[i].radioGroup == group {
							m.items[i].selected = m.items[i].initialSelected
						}
					}
				}
			}
		case "R":
			for i := range m.items {
				if m.items[i].isNormal {
					m.items[i].state = m.items[i].initialState
				} else if m.items[i].isRadio {
					m.items[i].selected = m.items[i].initialSelected
				}
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

	b.WriteString("Link Selection\n\n")

	pathStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("186")).Bold(true) // Softer Yellow

	var srcStrs []string
	for _, s := range m.sourceDirs {
		srcStrs = append(srcStrs, pathStyle.Render(s))
	}
	src := strings.Join(srcStrs, ", ")
	dst := pathStyle.Render(m.destDir)

	b.WriteString(fmt.Sprintf("The selected executables in %s\nwill be linked to %s.\n\n", src, dst))

	up := keyStyle.Render("<up>")
	down := keyStyle.Render("<down>")
	space := keyStyle.Render("<space>")
	aKey := keyStyle.Render("<a>")
	rKey := keyStyle.Render("<r>")
	RKey := keyStyle.Render("<R>")
	enter := keyStyle.Render("<enter>")
	qKey := keyStyle.Render("<q>")
	ctrlC := keyStyle.Render("<ctrl+c>")
	ctrlD := keyStyle.Render("<ctrl+d>")

	b.WriteString(fmt.Sprintf("Press %s/%s to navigate, %s to toggle, %s to select all/none.\n", up, down, space, aKey))
	b.WriteString(fmt.Sprintf("Press %s/%s to reset current/all, %s to confirm, %s,%s,%s to quit without saving.\n\n", rKey, RKey, enter, qKey, ctrlC, ctrlD))

	if len(m.items) == 0 {
		b.WriteString("No executable files found in the source directories.\n")
		return b.String()
	}

	currentSourceDir := ""
	currentRadioGroup := ""
	for i, it := range m.items {
		if it.isNormal {
			if it.sourceDir != currentSourceDir {
				b.WriteString(fmt.Sprintf("\n%s:\n", pathStyle.Render(it.sourceDir)))
				currentSourceDir = it.sourceDir
			}

			cursor := "  "
			if m.cursor == i {
				cursor = "> "
			}

			checked := "○"
			targetInfo := ""

			if it.initialState == StateInvalid {
				targetInfo = fmt.Sprintf(" [%s]", it.target)
			}

			if it.state == StateLinked {
				checked = "●"
			} else if it.state == StateInvalid {
				checked = "!"
			}

			b.WriteString(fmt.Sprintf("  %s%s %s%s\n", cursor, checked, it.name, targetInfo))
		} else if it.isRadio {
			if it.radioGroup != currentRadioGroup {
				b.WriteString(fmt.Sprintf("\n%s (duplicate in multiple sources):\n", redStyle.Render(it.radioGroup)))
				currentRadioGroup = it.radioGroup
			}

			cursor := "  "
			if m.cursor == i {
				cursor = "> "
			}

			checked := "○"
			if it.selected {
				checked = "●"
			}
			b.WriteString(fmt.Sprintf("  %s%s %s\n", cursor, checked, it.radioLabel))
		}
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
				if IsExecutable(info) {
					execs = append(execs, entry.Name())
				}
			}
		}
	}
	return execs, nil
}

func applyChanges(destDir string, items []item) error {
	absDestDir, err := filepath.Abs(destDir)
	if err != nil {
		return err
	}

	var errorMessages []string

	for _, it := range items {
		if it.isNormal {
			if it.state == it.initialState {
				continue // No change
			}
			if it.state == StateInvalid { // Cannot intentionally select an invalid state
				continue
			}

			destPath := getDestPath(absDestDir, it.name)
			sourcePath := filepath.Join(it.sourceDir, it.name)

			if it.state == StateLinked {
				// Newly selected: create link
				err := checkLinkConflict(destPath)
				if err != nil {
					errorMessages = append(errorMessages, fmt.Sprintf("Warning: '%s' %v", destPath, err))
					continue
				}
				os.Remove(destPath)

				err = createLink(sourcePath, destPath)
				if err != nil {
					errorMessages = append(errorMessages, fmt.Sprintf("Error creating link for '%s': %v", it.name, err))
				} else {
					fmt.Printf("Created link: %s\n", destPath)
				}
			} else if it.state == StateNone {
				// Newly unselected: remove link ONLY if it's managed by us
				state, _ := getLinkState(absDestDir, it.sourceDir, it.name)
				if state == StateLinked || state == StateInvalid {
					err := removeLink(destPath)
					if err != nil {
						errorMessages = append(errorMessages, fmt.Sprintf("Error removing link for '%s': %v", it.name, err))
					} else {
						fmt.Printf("Removed link: %s\n", destPath)
					}
				}
			}
		} else if it.isRadio {
			if it.selected && !it.initialSelected {
				destPath := filepath.Join(absDestDir, it.radioGroup)

				if it.radioValue == "none" {
					err := removeLink(destPath)
					if err != nil {
						errorMessages = append(errorMessages, fmt.Sprintf("Error removing link for '%s': %v", it.radioGroup, err))
					} else {
						fmt.Printf("Removed link: %s\n", destPath)
					}
				} else {
					err := checkLinkConflict(destPath)
					if err != nil {
						errorMessages = append(errorMessages, fmt.Sprintf("Warning: '%s' %v", destPath, err))
						continue
					}
					os.Remove(destPath)

					err = createLink(it.radioValue, destPath)
					if err != nil {
						errorMessages = append(errorMessages, fmt.Sprintf("Error creating link for '%s': %v", it.radioGroup, err))
					} else {
						fmt.Printf("Created link: %s\n", destPath)
					}
				}
			}
		}
	}

	if len(errorMessages) > 0 {
		fmt.Println(strings.Join(errorMessages, "\n"))
	}

	return nil
}

func main() {
	var sources StringSliceFlag
	var dest string
	var versionFlag bool

	flag.Var(&sources, "source", "Source directory (can also use -s)")
	flag.Var(&sources, "s", "Source directory (shorthand for --source)")
	flag.StringVar(&dest, "destination", "", "Destination directory (can also use -d)")
	flag.StringVar(&dest, "d", "", "Destination directory (shorthand for --destination)")
	flag.BoolVar(&versionFlag, "version", false, "Print version information (can also use -v)")
	flag.BoolVar(&versionFlag, "v", false, "Print version information (shorthand for --version)")

	// Go's flag package handles -h and --help automatically.

	// Change custom usage description
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "A CLI tool to interactively manage links between directories.\n\n")
		fmt.Fprintf(os.Stderr, "Usage: lks -s <src1> -s <src2> -d <dest>\n\n")
		fmt.Fprintf(os.Stderr, "Options:\n")
		fmt.Fprintf(os.Stderr, "  -s, --source <path>\n\tSource directory containing executables (can be specified multiple times)\n")
		fmt.Fprintf(os.Stderr, "  -d, --destination <path>\n\tDestination directory for links\n")
		fmt.Fprintf(os.Stderr, "  -v, --version\n\tShow version information\n")
		fmt.Fprintf(os.Stderr, "  -h, --help\n\tShow help message\n")
	}

	flag.Parse()

	if versionFlag {
		fmt.Printf("lks version %s\n", getVersion())
		os.Exit(0)
	}

	if len(sources) == 0 || dest == "" {
		fmt.Println("Error: Both source (-s, --source) and destination (-d, --destination) are required.")
		flag.Usage()
		os.Exit(1)
	}

	var absSourceDirs []string
	for _, s := range sources {
		absDir, err := filepath.Abs(s)
		if err != nil {
			fmt.Printf("Error resolving source directory '%s': %v\n", s, err)
			os.Exit(1)
		}
		absSourceDirs = append(absSourceDirs, absDir)
	}

	absDestDir, err := filepath.Abs(dest)
	if err != nil {
		fmt.Printf("Error resolving destination directory: %v\n", err)
		os.Exit(1)
	}

	fileSources := make(map[string][]string)

	for _, absSourceDir := range absSourceDirs {
		executables, err := getExecutables(absSourceDir)
		if err != nil {
			fmt.Printf("Error reading source directory '%s': %v\n", absSourceDir, err)
			os.Exit(1)
		}

		for _, exe := range executables {
			fileSources[exe] = append(fileSources[exe], absSourceDir)
		}
	}

	var items []item

	for _, absSourceDir := range absSourceDirs {
		executables, _ := getExecutables(absSourceDir)
		for _, exe := range executables {
			if len(fileSources[exe]) == 1 {
				state, target := getLinkState(absDestDir, absSourceDir, exe)
				items = append(items, item{
					name:         exe,
					isNormal:     true,
					sourceDir:    absSourceDir,
					state:        state,
					initialState: state,
					target:       target,
				})
			}
		}
	}

	var duplicates []string
	for exe, srcs := range fileSources {
		if len(srcs) > 1 {
			duplicates = append(duplicates, exe)
		}
	}
	sort.Strings(duplicates)

	for _, exe := range duplicates {
		srcs := fileSources[exe]
		destPath := filepath.Join(absDestDir, exe)
		target, err := os.Readlink(destPath)

		existingTarget := ""
		if err == nil {
			existingTarget = target
		}

		hasExistingTargetOption := false
		if existingTarget != "" {
			matchedSource := false
			for _, s := range srcs {
				if existingTarget == filepath.Join(s, exe) {
					matchedSource = true
					break
				}
			}
			if !matchedSource {
				hasExistingTargetOption = true
			}
		}

		if hasExistingTargetOption {
			items = append(items, item{
				isRadio:         true,
				radioGroup:      exe,
				radioValue:      existingTarget,
				radioLabel:      existingTarget,
				selected:        true,
				initialSelected: true,
			})
		}

		for _, s := range srcs {
			val := filepath.Join(s, exe)
			selected := (val == existingTarget)
			items = append(items, item{
				isRadio:         true,
				radioGroup:      exe,
				radioValue:      val,
				radioLabel:      val,
				selected:        selected,
				initialSelected: selected,
			})
		}

		selectedNone := (existingTarget == "")
		items = append(items, item{
			isRadio:         true,
			radioGroup:      exe,
			radioValue:      "none",
			radioLabel:      "none",
			selected:        selectedNone,
			initialSelected: selectedNone,
		})
	}

	p := tea.NewProgram(initialModel(absSourceDirs, absDestDir, items))
	m, err := p.Run()
	if err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}

	if finalModel, ok := m.(model); ok {
		if finalModel.confirmed {
			// Print a newline to separate TUI from execution logs
			fmt.Println("\nApplying changes...")
			err := applyChanges(absDestDir, finalModel.items)
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
