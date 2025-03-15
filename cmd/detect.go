package cmd

import (
	"os/exec"
)

// PackageManager represents a system package manager
type PackageManager struct {
	Name     string
	Bin      string
	Type     string
	Commands map[string][]string
}

// DetectPackageManager identifies which package manager is available on the system
func DetectPackageManager() *PackageManager {
	// Check for Homebrew (macOS)
	if _, err := exec.LookPath("brew"); err == nil {
		return &PackageManager{
			Name: "brew",
			Bin:  "brew",
			Type: "macos",
			Commands: map[string][]string{
				"install":      {"install"},
				"reinstall":    {"reinstall"},
				"remove":       {"uninstall"},
				"update":       {"update"},
				"upgrade":      {"upgrade"},
				"search":       {"search"},
				"info":         {"info"},
				"autoremove":   {"autoremove"},
				"clean":        {"cleanup"},
				"add-repo":     {"tap"},
				"add-key":      {""},
				"enable-repo":  {""},
				"disable-repo": {""},
			},
		}
	}

	// Check for apt (Debian/Ubuntu)
	if _, err := exec.LookPath("apt"); err == nil {
		return &PackageManager{
			Name: "apt",
			Bin:  "apt",
			Type: "debian",
			Commands: map[string][]string{
				"install":      {"install"},
				"reinstall":    {"install", "--reinstall"},
				"remove":       {"remove", "--purge"},
				"update":       {"update"},
				"upgrade":      {"upgrade"},
				"search":       {"search"},
				"info":         {"show"},
				"autoremove":   {"autoremove", "--purge"},
				"clean":        {"clean"},
				"add-repo":     {""},
				"add-key":      {""},
				"enable-repo":  {""},
				"disable-repo": {""},
			},
		}
	}

	// Check for dnf (Fedora/RHEL 8+)
	if _, err := exec.LookPath("dnf"); err == nil {
		return &PackageManager{
			Name: "dnf",
			Bin:  "dnf",
			Type: "redhat",
			Commands: map[string][]string{
				"install":      {"install"},
				"reinstall":    {"reinstall"},
				"remove":       {"remove"},
				"update":       {"check-update"},
				"upgrade":      {"upgrade"},
				"search":       {"search"},
				"info":         {"info"},
				"autoremove":   {"autoremove"},
				"clean":        {"clean", "all"},
				"add-repo":     {""},
				"add-key":      {""},
				"enable-repo":  {""},
				"disable-repo": {""},
			},
		}
	}

	// Check for yum (CentOS/RHEL 7 and earlier)
	if _, err := exec.LookPath("yum"); err == nil {
		return &PackageManager{
			Name: "yum",
			Bin:  "yum",
			Type: "redhat",
			Commands: map[string][]string{
				"install":      {"install"},
				"reinstall":    {"reinstall"},
				"remove":       {"remove"},
				"update":       {"check-update"},
				"upgrade":      {"upgrade"},
				"search":       {"search"},
				"info":         {"info"},
				"autoremove":   {"autoremove"},
				"clean":        {"clean", "all"},
				"add-repo":     {""},
				"add-key":      {""},
				"enable-repo":  {""},
				"disable-repo": {""},
			},
		}
	}

	// Check for apk (Alpine)
	if _, err := exec.LookPath("apk"); err == nil {
		return &PackageManager{
			Name: "apk",
			Bin:  "apk",
			Type: "alpine",
			Commands: map[string][]string{
				"install":      {"add"},
				"reinstall":    {"add", "--force-overwrite"},
				"remove":       {"del"},
				"update":       {"update"},
				"upgrade":      {"upgrade"},
				"search":       {"search"},
				"info":         {"info"},
				"autoremove":   {"autoremove"},
				"clean":        {"cache", "clean"},
				"add-repo":     {""},
				"add-key":      {""},
				"enable-repo":  {""},
				"disable-repo": {""},
			},
		}
	}

	// Check for pacman (Arch)
	if _, err := exec.LookPath("pacman"); err == nil {
		return &PackageManager{
			Name: "pacman",
			Bin:  "pacman",
			Type: "arch",
			Commands: map[string][]string{
				"install":      {"-S"},
				"reinstall":    {"-S", "--needed"},
				"remove":       {"-Rns"},
				"update":       {"-Sy"},
				"upgrade":      {"-Syu"},
				"search":       {"-Ss"},
				"info":         {"-Si"},
				"autoremove":   {"-Rns", "$(pacman -Qdtq)"},
				"clean":        {"-Sc"},
				"add-repo":     {""},
				"add-key":      {""},
				"enable-repo":  {""},
				"disable-repo": {""},
			},
		}
	}

	// Default to nil if no package manager is found
	return nil
}
