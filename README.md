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

# Reinstall packages
pkgs reinstall nginx
pkgs ri vim git curl

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

# Add repository keys
pkgs add-key nodesource https://deb.nodesource.com/gpgkey/nodesource.gpg.key

# Add repositories
pkgs add-repo nodesource "deb [signed-by=/etc/apt/keyrings/nodesource.asc] https://deb.nodesource.com/node_20.x nodistro main"

# Enable a repository
pkgs enable-repo docker-ce

# Disable a repository
pkgs disable-repo docker-ce

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

### Repository Management

`pkgs` provides commands to manage package repositories and their signing keys:

```bash
# Add a repository key
pkgs add-key [name] url

# Examples:
# For apt-based systems (Debian/Ubuntu)
pkgs add-key nodesource https://deb.nodesource.com/gpgkey/nodesource.gpg.key

# For Alpine Linux
pkgs add-key alpine-key https://alpine-keys.example.com/key.rsa.pub
# Or without specifying a name (will use the filename from the URL)
pkgs add-key https://alpine-keys.example.com/key.rsa.pub
```

```bash
# Add a repository
pkgs add-repo [name] url

# Examples:
# For apt-based systems (Debian/Ubuntu)
pkgs add-repo nodesource "deb [signed-by=/etc/apt/keyrings/nodesource.asc] https://deb.nodesource.com/node_20.x nodistro main"

# For dnf/yum-based systems (Fedora/RHEL/CentOS)
pkgs add-repo https://download.docker.com/linux/fedora/docker-ce.repo

# For Alpine Linux
pkgs add-repo edge-testing https://dl-cdn.alpinelinux.org/alpine/edge/testing

# For Homebrew
pkgs add-repo homebrew/cask-fonts
```

```bash
# Enable a repository
pkgs enable-repo name

# Examples:
# For apt-based systems (Debian/Ubuntu)
pkgs enable-repo nodesource

# For dnf/yum-based systems
pkgs enable-repo docker-ce

# For Alpine Linux
pkgs enable-repo edge-testing
```

```bash
# Disable a repository
pkgs disable-repo name

# Examples:
# For apt-based systems (Debian/Ubuntu)
pkgs disable-repo nodesource

# For dnf/yum-based systems
pkgs disable-repo docker-ce

# For Alpine Linux
pkgs disable-repo edge-testing
```

These commands handle the package manager-specific details for you, making it easier to manage repositories across different systems.

## Package Manager Specifics

### Homebrew (macOS)

On macOS systems, `pkgs` automatically detects and uses Homebrew. Some key differences when using Homebrew:

- Commands never use sudo (as recommended by Homebrew)
- `autoremove` runs both `brew autoremove` to remove unused dependencies and `brew cleanup` to remove old versions
- `remove` uses `brew uninstall` instead of remove/purge
- `reinstall` uses `brew reinstall` to reinstall packages
- `add-repo` uses `brew tap` to add new taps
- `add-key` is not applicable for Homebrew

### Linux Package Managers

Each Linux package manager has its own specific implementation:

- `apt` (Debian/Ubuntu): 
  - Uses `--purge` for thorough removal
  - Uses `--reinstall` flag for reinstalling packages
  - `add-key` saves keys to `/etc/apt/keyrings/name.asc`
  - `add-repo` creates files in `/etc/apt/sources.list.d/name.list`
  - `enable-repo` uncomments entries in repository files
  - `disable-repo` comments out entries in repository files
- `dnf`/`yum` (RedHat): 
  - Uses `check-update` for the update command
  - Has a dedicated `reinstall` command
  - `add-repo` creates files in `/etc/yum.repos.d/` directory
  - `add-key` provides guidance for using `rpm --import`
  - `enable-repo` sets `enabled=1` in repository files
  - `disable-repo` sets `enabled=0` in repository files
- `apk` (Alpine): 
  - Uses `add` and `del` instead of install/remove
  - Uses `add --force-overwrite` for reinstalling
  - `add-key` adds keys to `/etc/apk/keys/`
  - `add-repo` adds repositories to `/etc/apk/repositories`
  - `enable-repo` uncomments repository entries
  - `disable-repo` comments out repository entries
- `pacman` (Arch): 
  - Uses special flags like `-S`, `-Rns`, etc.
  - Uses `-S --needed` for reinstalling packages
  - `add-key` provides guidance for using `pacman-key --add`
  - `add-repo` provides guidance for manually editing `/etc/pacman.conf`
  - `enable-repo` and `disable-repo` provide guidance for manually editing `/etc/pacman.conf`

#### Privilege Elevation on Linux

On Linux systems, package management operations typically require root privileges. The `pkgs` tool will:

1. Check if the current user has root privileges
2. If not, automatically use `sudo` to elevate privileges for commands that require it
3. If `sudo` is not available, provide a clear error message

Commands that require privilege elevation:
- install
- reinstall
- remove
- update
- upgrade
- autoremove
- clean
- add-key
- add-repo
- enable-repo
- disable-repo

## License

MIT License

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request. 