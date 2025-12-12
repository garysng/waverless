#!/bin/bash

# Waverless simplified build script
# Auto-increment version and build API Server or Web UI

set -e

VERSION_FILE_API=".version"
VERSION_FILE_WEB=".version-web"
REGISTRY="${IMAGE_REGISTRY:-docker.io/wavespeed}"
USE_PROXY=false

# Default proxy settings
PROXY_HOST="127.0.0.1"
PROXY_PORT="7890"

# Color output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# Get current version
get_current_version() {
    local component=$1
    local version_file

    if [ "$component" = "web" ]; then
        version_file="$VERSION_FILE_WEB"
    else
        version_file="$VERSION_FILE_API"
    fi

    if [ -f "$version_file" ]; then
        cat "$version_file"
    else
        echo "1.0.0"
    fi
}

# Increment version (patch number +1)
increment_version() {
    local version=$1
    local major minor patch

    IFS='.' read -r major minor patch <<< "$version"
    patch=$((patch + 1))
    echo "${major}.${minor}.${patch}"
}

# Save version
save_version() {
    local component=$1
    local version=$2
    local version_file

    if [ "$component" = "web" ]; then
        version_file="$VERSION_FILE_WEB"
    else
        version_file="$VERSION_FILE_API"
    fi

    echo "$version" > "$version_file"
}

# Setup proxy
setup_proxy() {
    if [ "$USE_PROXY" = true ]; then
        export https_proxy="http://${PROXY_HOST}:${PROXY_PORT}"
        export http_proxy="http://${PROXY_HOST}:${PROXY_PORT}"
        export all_proxy="socks5://${PROXY_HOST}:${PROXY_PORT}"
        echo -e "${YELLOW}Proxy enabled: ${PROXY_HOST}:${PROXY_PORT}${NC}"
    fi
}

# Show help
show_help() {
    cat << EOF
Usage: $0 [COMMAND] [OPTIONS]

Commands:
  api       Build and push API Server image
  web       Build and push Web UI image
  version   Show current version and next version
  help      Show this help message

Options:
  -p        Enable proxy (default: 127.0.0.1:7890)

Examples:
  $0 api              # Build API Server (auto-increment version)
  $0 api -p           # Build API Server with proxy
  $0 web -p           # Build Web UI with proxy
  $0 version          # Check version

Environment Variables:
  IMAGE_REGISTRY      Image registry (default: docker.io/wavespeed)
  PROXY_HOST          Proxy host (default: 127.0.0.1)
  PROXY_PORT          Proxy port (default: 7890)
EOF
}

# Build image
build_image() {
    local component=$1
    local current_version=$(get_current_version "$component")
    local next_version=$(increment_version "$current_version")

    echo -e "${BLUE}================================${NC}"
    echo -e "${BLUE}Waverless Build Tool${NC}"
    echo -e "${BLUE}================================${NC}"
    echo -e "Component:       ${component}"
    echo -e "Current version: ${YELLOW}${current_version}${NC}"
    echo -e "Next version:    ${GREEN}${next_version}${NC}"
    echo -e "Registry:        ${REGISTRY}"

    # Setup proxy if enabled
    setup_proxy

    echo ""

    # Confirm
    read -p "Continue building? (y/n): " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        echo "Cancelled"
        exit 0
    fi

    # Call deploy.sh to build
    if [ "$component" = "api" ]; then
        IMAGE_REGISTRY=${REGISTRY} IMAGE_TAG=${next_version} ./deploy.sh -t ${next_version} build
    else
        IMAGE_REGISTRY=${REGISTRY} IMAGE_TAG=${next_version} ./deploy.sh -t ${next_version} build-web
    fi

    # Save new version
    save_version "$component" "$next_version"

    echo ""
    echo -e "${GREEN}================================${NC}"
    echo -e "${GREEN}✓ Build completed!${NC}"
    echo -e "${GREEN}================================${NC}"
    echo -e "Image: ${REGISTRY}/waverless$([ "$component" = "web" ] && echo "-web"):${next_version}"
    echo -e "Version updated: ${current_version} → ${GREEN}${next_version}${NC}"
    echo ""
}

# Show version
show_version() {
    local current_version_api=$(get_current_version "api")
    local next_version_api=$(increment_version "$current_version_api")
    local current_version_web=$(get_current_version "web")
    local next_version_web=$(increment_version "$current_version_web")

    echo -e "${BLUE}API Server:${NC}"
    echo -e "  Current version: ${YELLOW}${current_version_api}${NC}"
    echo -e "  Next version:    ${GREEN}${next_version_api}${NC}"
    echo ""
    echo -e "${BLUE}Web UI:${NC}"
    echo -e "  Current version: ${YELLOW}${current_version_web}${NC}"
    echo -e "  Next version:    ${GREEN}${next_version_web}${NC}"
}

# Main function
main() {
    local command=""

    # Parse arguments
    while [[ $# -gt 0 ]]; do
        case $1 in
            -p|--proxy)
                USE_PROXY=true
                shift
                ;;
            api|web|version|help)
                command=$1
                shift
                ;;
            --help|-h)
                command="help"
                shift
                ;;
            *)
                echo "Error: Unknown option '$1'"
                echo ""
                show_help
                exit 1
                ;;
        esac
    done

    # Default to help if no command
    if [ -z "$command" ]; then
        command="help"
    fi

    # Execute command
    case "$command" in
        api)
            build_image "api"
            ;;
        web)
            build_image "web"
            ;;
        version)
            show_version
            ;;
        help)
            show_help
            ;;
        *)
            echo "Error: Unknown command '$command'"
            echo ""
            show_help
            exit 1
            ;;
    esac
}

main "$@"
