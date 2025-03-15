# pkgs - Unified Package Manager Interface

`pkgs` is a command-line tool that provides a unified interface for package management across different operating
systems. It wraps around native package managers like `apt`, `dnf`, `yum`, `apk`, `pacman`, and `brew`, allowing you
to use the same commands regardless of the underlying system.

## Features

- Unified commands across different operating systems (Linux distributions and macOS)
- Automatic detection of the system's package manager
- Support for common package management operations
- Intelligent privilege handling:
  - Automatic sudo elevation on Linux when required
  - No sudo usage on macOS with Homebrew (as recommended)
- Intelligent handling of package manager-specific behaviors

## Supported Package Managers

- `brew` (macOS)
- `apt` (Debian, Ubuntu)
- `dnf` (Fedora, RHEL 8+)
- `yum` (CentOS, RHEL 7 and earlier)
- `apk` (Alpine)
- `pacman` (Arch)

## Installation

### From Source

```bash
git clone https://github.com/mobydeck/pkgs.git
cd pkgs
just build
sudo just install
```

### Using Go

```bash
git clone https://github.com/mobydeck/pkgs.git
cd pkgs
go build
sudo mv pkgs /usr/local/bin/
```

### Just Commands

The project includes a justfile with the following commands:

```bash
# Build the application
just build

# Install the application to /usr/local/bin
just install

# Clean build artifacts
just clean

# Show the current version
just version

# Build for specific platforms
just build-linux-amd64
just build-macos-arm64
just build-windows-amd64

# Build for all platforms
just build-all

# Create release packages
just package
```

## Usage

### Basic Commands

```bash
# Install packages
pkgs install nginx
pkgs i vim git curl

# Remove packages
pkgs remove nginx
pkgs rm vim git curl

# Search for packages
pkgs search nginx
pkgs s python

# Show package information
pkgs info nginx
pkgs show vim

# Update package lists
pkgs update
pkgs up

# Upgrade installed packages
pkgs upgrade
pkgs ug

# Remove unused packages
pkgs autoremove

# Clean package cache
pkgs clean

# Show which package manager is being used
pkgs which

# Show only the package manager name (useful for scripting)
pkgs which -s
```

For example, a script could now do something like:
```bash
PM=$(pkgs which -s)
if [ "$PM" = "apt" ]; then
    # Do something specific for apt systems
elif [ "$PM" = "brew" ]; then
    # Do something specific for Homebrew systems
fi
```

### Help

```bash
# Show general help
pkgs --help

# Show command-specific help
pkgs install --help
```

### Version

```bash
# Show version information
pkgs --version

# Run commands non-interactively (useful for CI/CD and automation)
pkgs --yes install nginx
pkgs -y update
```

### Non-Interactive Mode

For CI/CD pipelines and automation scripts, you can use the `--yes` or `-y` flag to run commands non-interactively:

```bash
# Install packages without prompting
pkgs -y install nginx

# Update and upgrade without prompting
pkgs -y update && pkgs -y upgrade

# Remove packages without prompting
pkgs -y remove nginx
```

This flag automatically adds the appropriate non-interactive flag to the underlying package manager:
- `-y` for apt, dnf, and yum
- `--noconfirm` for pacman
- No additional flag for brew and apk (as they're already non-interactive by default)

Alternatively, you can set the `PKGS_YES` environment variable to achieve the same effect:

```bash
# Set the environment variable for the current session
export PKGS_YES=true

# Now all commands will run in non-interactive mode
pkgs install nginx
pkgs update
pkgs remove nginx

# You can also set it for a single command
PKGS_YES=1 pkgs install nginx
```

The `PKGS_YES` environment variable accepts the following values (case-insensitive):
- `true`, `yes`, `1`, `y`: Enable non-interactive mode
- Any other value or unset: Use the default interactive mode


## Package Manager Specifics

### Homebrew (macOS)

On macOS systems, `pkgs` automatically detects and uses Homebrew. Some key differences when using Homebrew:

- Commands never use sudo (as recommended by Homebrew)
- `autoremove` runs both `brew autoremove` to remove unused dependencies and `brew cleanup` to remove old versions
- `remove` uses `brew uninstall` instead of remove/purge

### Linux Package Managers

Each Linux package manager has its own specific implementation:

- `apt` (Debian/Ubuntu): Uses `--purge` for thorough removal
- `dnf`/`yum` (RedHat): Uses `check-update` for the update command
- `apk` (Alpine): Uses `add` and `del` instead of install/remove
- `pacman` (Arch): Uses special flags like `-S`, `-Rns`, etc.

#### Privilege Elevation on Linux

On Linux systems, package management operations typically require root privileges. The `pkgs` tool will:

1. Check if the current user has root privileges
2. If not, automatically use `sudo` to elevate privileges for commands that require it
3. If `sudo` is not available, provide a clear error message

Commands that require privilege elevation:
- install
- remove
- update
- upgrade
- autoremove
- clean

## License

MIT License

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request. 