# How to Build OttoMap from Scratch

OttoMap is a project written in Go, a statically typed, compiled programming language designed for simplicity and efficiency. Building the project is straightforward due to Go's cross-platform nature and simple build process.

The basic steps are:
1. Install the required tools (Git and Go).
2. Clone the source code using Git.
3. Build the executable using Go.

This document provides step-by-step instructions for Windows, macOS, and Linux. If you encounter any issues or have questions, please join our Discord server for support.

## Definitions

- **Go**: A programming language.
- **Repository**: A storage location for software packages.
- **Clone**: To make a copy of a repository from a remote server to your local machine.
- **Build tool**: A program that automates the process of building software.

## Prerequisites

Ensure you have the following tools installed before starting:

- **Git**: Used to download and update the source code.
- **Go**: Used to build the project.

You can install these tools from their official websites or use a package manager for easier updates. Popular package managers include `winget` (Windows), `snap` (Linux), and `brew` (macOS).

- For more information about `winget`, visit the [Windows Package Manager](https://learn.microsoft.com/en-us/windows/package-manager/) site.
- For more information about `snap`, visit the [Snapcraft](https://snapcraft.io/docs/installing-snapd) site.
- For more information about `brew`, visit the [Homebrew](https://brew.sh/) site.

## Installing Git

Git is essential for downloading and updating the source code from GitHub. Follow the instructions on the [Install Git](https://github.com/git-guides/install-git) page for your operating system to install Git.

## Installing Go

Go is necessary for building the OttoMap executable. The installation steps vary based on your operating system and package manager preference.

### Windows (Using winget)

1. Open a command prompt.
2. Run the following command:

   ```sh
   winget install Go
   ```

### Linux (Using snap)

1. Open a terminal.
2. Run the following command:

   ```sh
   sudo snap install go --classic
   ```

### macOS (Using brew)

1. Open a terminal.
2. Run the following command:

   ```sh
   brew install go
   ```

### Official Installer

Alternatively, visit the [Download and install](https://go.dev/doc/install) page and follow the instructions for your operating system.

## Cloning the Repository

To build the project, you need a copy of the source code. You can clone the repository using Git (recommended) or download a ZIP file from GitHub.

### Using Git

1. Open your terminal or command prompt.
2. Navigate to the directory where you want to clone the repository.
3. Run the following command:

   ```sh
   git clone https://github.com/playbymail/ottomap.git
   ```

### Using a ZIP File

Download the ZIP file from the official OttoMap repository at [https://github.com/playbymail/ottomap](https://github.com/playbymail/ottomap).

## Building the Project

The build process is consistent across all operating systems:

1. Open your terminal or command prompt.
2. Navigate to the directory containing the source code.
3. Run the following command to build the executable:

   ```sh
   go build
   ```

This command creates an executable named `ottomap` (for macOS and Linux) or `ottomap.exe` (for Windows) in the current directory.

For additional help or if you encounter issues, please join our Discord server. The Go team also provides excellent documentation, which you can find at [https://go.dev/doc/tutorial/compile-install](https://go.dev/doc/tutorial/compile-install).

Please see the `OTTOMAP.md` file for more information about using OttoMap.

# Cross Compiling
Experimental!

```shell
GOOS=windows GOARCH=amd64 go build -o ottomap.exe
scp ottomap.exe shadowcairn:/var/www/ottosvc/assets/ottomap-0.12.16.exe
```