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
				"install":    {"install"},
				"remove":     {"uninstall"},
				"update":     {"update"},
				"upgrade":    {"upgrade"},
				"search":     {"search"},
				"info":       {"info"},
				"autoremove": {"autoremove"},
				"clean":      {"cleanup"},
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
				"install":    {"install"},
				"remove":     {"remove", "--purge"},
				"update":     {"update"},
				"upgrade":    {"upgrade"},
				"search":     {"search"},
				"info":       {"show"},
				"autoremove": {"autoremove", "--purge"},
				"clean":      {"clean"},
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
				"install":    {"install"},
				"remove":     {"remove"},
				"update":     {"check-update"},
				"upgrade":    {"upgrade"},
				"search":     {"search"},
				"info":       {"info"},
				"autoremove": {"autoremove"},
				"clean":      {"clean", "all"},
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
				"install":    {"install"},
				"remove":     {"remove"},
				"update":     {"check-update"},
				"upgrade":    {"upgrade"},
				"search":     {"search"},
				"info":       {"info"},
				"autoremove": {"autoremove"},
				"clean":      {"clean", "all"},
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
				"install":    {"add"},
				"remove":     {"del"},
				"update":     {"update"},
				"upgrade":    {"upgrade"},
				"search":     {"search"},
				"info":       {"info"},
				"autoremove": {"autoremove"},
				"clean":      {"cache", "clean"},
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
				"install":    {"-S"},
				"remove":     {"-Rns"},
				"update":     {"-Sy"},
				"upgrade":    {"-Syu"},
				"search":     {"-Ss"},
				"info":       {"-Si"},
				"autoremove": {"-Rns", "$(pacman -Qdtq)"},
				"clean":      {"-Sc"},
			},
		}
	}

	// Default to nil if no package manager is found
	return nil
}
