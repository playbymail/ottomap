#!/bin/bash

set -euo pipefail

RSYNC_PROGRESS=--progress
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Cleanup function
cleanup() {
    local exit_code=$?
    if [ $exit_code -ne 0 ]; then
        log_error "Deployment failed with exit code $exit_code"
        log_info "Cleaning up temporary files..."
        rm -f "${REPO_ROOT}/build/ottomap"
        rm -f "${REPO_ROOT}/build/ottomap-windows-"*.exe
        rm -f "${REPO_ROOT}/build/ottomap-"*.tar.gz
    fi
    exit $exit_code
}

trap cleanup EXIT

log_info "Starting deployment process..."

# Confirm that we're running from the root of the repository
cd "${REPO_ROOT}"
if [ ! -d build ]; then
    log_error "Must run from the root of the repository (build directory not found)"
    exit 2
fi

log_info "Working directory: ${REPO_ROOT}"

# Build the local executable to get version number
LOCAL_EXE=build/ottomap
log_info "Building local executable to determine version..."
if ! go build -o "${LOCAL_EXE}"; then
    log_error "Unable to build local executable"
    exit 2
fi

OTTOVER=$("${LOCAL_EXE}" version)
if [ -z "${OTTOVER}" ]; then
    log_error "'${LOCAL_EXE} version' failed or returned empty version"
    exit 2
fi
log_success "Version detected: ${OTTOVER}"

# Clean up the local executable
rm -f "${LOCAL_EXE}"

# Copy the installation script into the build directory
INSTALL_SCRIPT="bin/install.sh"
log_info "Copying install script to build directory..."
if ! cp -p "${INSTALL_SCRIPT}" build/; then
    log_error "Failed to copy install script to build directory"
    exit 2
fi

log_info "Building executables for version '${OTTOVER}'"

# Build Linux executable
LINUX_EXE=build/ottomap
log_info "Building Linux executable..."
if ! GOOS=linux GOARCH=amd64 go build -o "${LINUX_EXE}-${OTTOVER}"; then
    log_error "Failed to build Linux executable"
    exit 2
fi
log_success "Linux executable built: ${LINUX_EXE}-${OTTOVER}"

# Build Windows executable
WINDOWS_EXE="build/ottomap-windows-${OTTOVER}.exe"
log_info "Building Windows executable..."
if ! GOOS=windows GOARCH=amd64 go build -o "${WINDOWS_EXE}"; then
    log_error "Failed to build Windows executable"
    exit 2
fi
log_success "Windows executable built: ${WINDOWS_EXE}"

# Create deployment tarball
TARBALL="build/ottomap-${OTTOVER}.tar"
log_info "Creating deployment tarball: ${TARBALL}"
if ! tar -c --no-xattrs --no-mac-metadata -f "${TARBALL}" build/install.sh "${LINUX_EXE}-${OTTOVER}" "${WINDOWS_EXE}"; then
    log_error "Failed to create tarball"
    exit 2
fi
log_success "Tarball created: ${TARBALL}"
tar tvf "${TARBALL}"

# Stop services on production server
log_info "Stopping ottomap services on production server..."
if ! ssh tribenet "systemctl stop ottomap.timer"; then
    log_warning "Failed to stop ottomap.timer (may not be running)"
fi
sleep 2

if ! ssh tribenet "systemctl stop ottomap.service"; then
    log_warning "Failed to stop ottomap.service (may not be running)"
fi
sleep 2

# Push tarball to production server
REMOTE_TARBALL="/tmp/ottomap-${OTTOVER}.tar"
log_info "Pushing tarball to production server..."
log_info "Local: ${TARBALL}"
log_info "Remote: ${REMOTE_TARBALL}"

RSYNC_SSH_OPTIONS="-o ServerAliveInterval=30 -o ServerAliveCountMax=5 -o TCPKeepAlive=yes -o IPQoS=throughput"
# rsync -e "ssh $RSYNC_SSH_OPTIONS" ...

if ! rsync -e "ssh $RSYNC_SSH_OPTIONS" -av ${RSYNC_PROGRESS} "${TARBALL}" "tribenet:${REMOTE_TARBALL}"; then
    log_error "Failed to copy tarball to production server"
    exit 2
fi
log_success "Tarball pushed to production server"

# Execute install script on production server
log_info "Executing install script on production server..."
if ! ssh mdhender@tribenet "cd /tmp && tar -xf ottomap-${OTTOVER}.tar && chmod +x build/install.sh && cd /tmp/build && ./install.sh ${OTTOVER}"; then
    log_error "Failed to execute install script on production server"
    exit 2
fi
log_success "Installation completed on production server"

# Start services on production server
log_info "Starting ottomap services on production server..."
if ! ssh tribenet "systemctl start ottomap.service"; then
    log_error "Failed to start ottomap.service"
    exit 2
fi

if ! ssh tribenet "systemctl start ottomap.timer"; then
    log_error "Failed to start ottomap.timer"
    exit 2
fi
log_success "Services started on production server"

# Clean up local build files
log_info "Cleaning up local build files..."
rm -f "${LINUX_EXE}" "${WINDOWS_EXE}" "${TARBALL}"
log_success "Local cleanup completed"

log_success "Deployment to production server completed successfully!"
exit 0
