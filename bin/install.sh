#!/bin/bash
# assumes that no services are running

set -euo pipefail

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

# Check if version argument is provided
if [ $# -ne 1 ]; then
    log_error "Usage: $0 <version>"
    log_error "Example: $0 0.52.2"
    exit 1
fi

VERSION="$1"
log_info "Installing ottomap version: ${VERSION}"

# Verify we're running as the correct user (must be mdhender)
CURRENT_USER=$(whoami)
log_info "Running as user: ${CURRENT_USER}"
if [ "${CURRENT_USER}" != "mdhender" ]; then
    log_error "install must run as mdhender on the server"
    exit 1
fi

LINUX_EXE="ottomap-${VERSION}"
WINDOWS_EXE="ottomap-windows-${VERSION}.exe"

# Check if we're in the extraction directory by verifying that the executables are present
if [ ! -f "${LINUX_EXE}" ]; then
    log_error "Linux executable not found in current directory: ${LINUX_EXE}"
    log_error "Make sure you extracted the tarball and are running from the extraction directory"
    exit 1
fi

# Check if windows executable exists
if [ ! -f "${WINDOWS_EXE}" ]; then
    log_error "Windows executable not found in current directory: ${WINDOWS_EXE}"
    log_error "Make sure you extracted the tarball and are running from the extraction directory"
    exit 1
fi

log_info "Found required files:"
log_info "  - ${LINUX_EXE} (Linux binary)"
log_info "  - ${WINDOWS_EXE} (Windows binary)"

# Create backup of current installation
BACKUP_DIR="/home/mdhender/bin/backup-$(date +%Y%m%d-%H%M%S)"
if [ -f "/home/mdhender/bin/ottomap" ]; then
    log_info "Creating backup of current installation..."
    mkdir -p "${BACKUP_DIR}"
    if ! cp -p "/home/mdhender/bin/ottomap" "${BACKUP_DIR}/ottomap.backup"; then
        log_warning "Failed to create backup (continuing anyway)"
    else
        log_success "Backup created: ${BACKUP_DIR}/ottomap.backup"
    fi
fi

# Ensure directories exist
log_info "Ensuring required directories exist..."
mkdir -p "/home/mdhender/bin"
mkdir -p "/var/www/ottomap.mdhenderson.com/bin"
mkdir -p "/var/www/ottomap.mdhenderson.com/assets/uploads"

# Install versioned Linux executable to user bin
VERSIONED_USER_BIN="/home/mdhender/bin/ottomap.${VERSION}"
log_info "Installing versioned Linux executable to user bin..."
log_info "Target: ${VERSIONED_USER_BIN}"
if ! cp -p "${LINUX_EXE}" "${VERSIONED_USER_BIN}"; then
    log_error "Failed to install versioned Linux executable to user bin"
    exit 1
fi
chmod +x "${VERSIONED_USER_BIN}"
log_success "Versioned executable installed to user bin"

# Install versioned Linux executable to web bin
VERSIONED_WEB_BIN="/var/www/ottomap.mdhenderson.com/bin/ottomap.${VERSION}"
log_info "Installing versioned Linux executable to web bin..."
log_info "Target: ${VERSIONED_WEB_BIN}"
if ! cp -p "${LINUX_EXE}" "${VERSIONED_WEB_BIN}"; then
    log_error "Failed to install versioned Linux executable to web bin"
    exit 1
fi
chmod +x "${VERSIONED_WEB_BIN}"
log_success "Versioned executable installed to web bin"

# Install Windows executable to web uploads
WEB_WINDOWS_EXE="/var/www/ottomap.mdhenderson.com/assets/uploads/ottomap-windows-${VERSION}.exe"
log_info "Installing Windows executable to web uploads..."
log_info "Target: ${WEB_WINDOWS_EXE}"
if ! cp -p "${WINDOWS_EXE}" "${WEB_WINDOWS_EXE}"; then
    log_error "Failed to install Windows executable to web uploads"
    exit 1
fi
chmod +r "${WEB_WINDOWS_EXE}"
log_success "Windows executable installed to web uploads"

# Update current Linux executable in user bin
CURRENT_USER_BIN="/home/mdhender/bin/ottomap"
log_info "Updating current Linux executable in user bin..."
log_info "Target: ${CURRENT_USER_BIN}"
if ! cp -p "${LINUX_EXE}" "${CURRENT_USER_BIN}"; then
    log_error "Failed to update current Linux executable in user bin"
    exit 1
fi
chmod +x "${CURRENT_USER_BIN}"
log_success "Current executable updated in user bin"

# Update current Linux executable in web bin
CURRENT_WEB_BIN="/var/www/ottomap.mdhenderson.com/bin/ottomap"
log_info "Updating current Linux executable in web bin..."
log_info "Target: ${CURRENT_WEB_BIN}"
if ! cp -p "${LINUX_EXE}" "${CURRENT_WEB_BIN}"; then
    log_error "Failed to update current Linux executable in web bin"
    exit 1
fi
chmod +x "${CURRENT_WEB_BIN}"
log_success "Current executable updated in web bin"

# Verify installations
log_info "Verifying installations..."

# Test user bin executable
if [ -x "${CURRENT_USER_BIN}" ]; then
    USER_VERSION=$("${CURRENT_USER_BIN}" version 2>/dev/null || echo "unknown")
    log_info "User bin executable version: ${USER_VERSION}"
    if [ "${USER_VERSION}" = "${VERSION}" ]; then
        log_success "User bin executable verification passed"
    else
        log_warning "User bin executable version mismatch (expected: ${VERSION}, got: ${USER_VERSION})"
    fi
else
    log_error "User bin executable is not executable"
    exit 1
fi

# Test web bin executable
if [ -x "${CURRENT_WEB_BIN}" ]; then
    WEB_VERSION=$("${CURRENT_WEB_BIN}" version 2>/dev/null || echo "unknown")
    log_info "Web bin executable version: ${WEB_VERSION}"
    if [ "${WEB_VERSION}" = "${VERSION}" ]; then
        log_success "Web bin executable verification passed"
    else
        log_warning "Web bin executable version mismatch (expected: ${VERSION}, got: ${WEB_VERSION})"
    fi
else
    log_error "Web bin executable is not executable"
    exit 1
fi

# Check file permissions and ownership
log_info "Checking file permissions and ownership..."
ls -la "${CURRENT_USER_BIN}" "${CURRENT_WEB_BIN}" "${WEB_WINDOWS_EXE}" 2>/dev/null || true

# Clean up extraction directory
log_info "Cleaning up temporary files..."
cd /
echo rm -f "/tmp/ottomap-${VERSION}.tar"
rm -rf "/tmp/build" 2>/dev/null || true

log_success "Installation completed successfully!"
log_info "Summary:"
log_info "  - Version: ${VERSION}"
log_info "  - User binary: ${CURRENT_USER_BIN}"
log_info "  - Web binary: ${CURRENT_WEB_BIN}"
log_info "  - Windows binary: ${WEB_WINDOWS_EXE}"
log_info "  - Versioned user binary: ${VERSIONED_USER_BIN}"
log_info "  - Versioned web binary: ${VERSIONED_WEB_BIN}"

if [ -n "${BACKUP_DIR:-}" ] && [ -d "${BACKUP_DIR}" ]; then
    log_info "  - Backup: ${BACKUP_DIR}"
fi

exit 0
