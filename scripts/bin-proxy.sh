#!/bin/bash

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BIN_MANIFESTS="${BIN_MANIFESTS:-$SCRIPT_DIR/bin-manifests.json}"
BIN_MANAGER_API="${BIN_MANAGER_API:-http://localhost:8080/api/v1}"
BIN_DIR="${BIN_DIR:-/usr/local/bin}"
LOG_FILE="${LOG_FILE:-/var/log/bin-proxy.log}"

log() {
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] $*" | tee -a "$LOG_FILE"
}

error() {
    log "ERROR: $*" >&2
}

get_md5sum() {
    local file="$1"
    if [[ ! -f "$file" ]]; then
        echo ""
        return
    fi
    md5sum "$file" | awk '{print $1}'
}

query_latest_md5() {
    local bin_name="$1"
    local url="${BIN_MANAGER_API}/releases/latest/${bin_name}/md5"
    
    local response
    response=$(curl -s -f "$url" 2>/dev/null)
    
    if [[ $? -ne 0 ]]; then
        error "Failed to query latest MD5 for $bin_name from $url"
        echo ""
        return 1
    fi
    
    echo "$response" | grep -o '"md5":"[^"]*"' | cut -d'"' -f4
}

download_binary() {
    local bin_name="$1"
    local temp_file="$2"
    local url="${BIN_MANAGER_API}/releases/latest/${bin_name}/download"
    
    log "Downloading $bin_name from $url"
    
    if ! curl -s -f -o "$temp_file" "$url"; then
        error "Failed to download $bin_name"
        return 1
    fi
    
    chmod +x "$temp_file"
    return 0
}

update_binary() {
    local bin_name="$1"
    local current_md5="$2"
    local latest_md5="$3"
    
    if [[ "$current_md5" == "$latest_md5" ]]; then
        log "$bin_name is already up to date (MD5: $current_md5)"
        return 0
    fi
    
    log "Updating $bin_name (current: $current_md5, latest: $latest_md5)"
    
    local bin_path="${BIN_DIR}/${bin_name}"
    local temp_file="/tmp/${bin_name}.tmp.$$"
    
    if ! download_binary "$bin_name" "$temp_file"; then
        return 1
    fi
    
    local downloaded_md5
    downloaded_md5=$(get_md5sum "$temp_file")
    
    if [[ "$downloaded_md5" != "$latest_md5" ]]; then
        error "MD5 mismatch for downloaded $bin_name (expected: $latest_md5, got: $downloaded_md5)"
        rm -f "$temp_file"
        return 1
    fi
    
    if [[ -f "$bin_path" ]]; then
        cp "$bin_path" "${bin_path}.backup"
    fi
    
    mv "$temp_file" "$bin_path"
    chmod +x "$bin_path"
    
    log "Successfully updated $bin_name"
    
    if command -v supervisorctl &> /dev/null; then
        log "Restarting service: $bin_name via supervisor"
        if supervisorctl restart "$bin_name" 2>&1 | tee -a "$LOG_FILE"; then
            log "Service $bin_name restarted successfully"
        else
            error "Failed to restart service $bin_name"
            if [[ -f "${bin_path}.backup" ]]; then
                log "Rolling back to previous version"
                mv "${bin_path}.backup" "$bin_path"
                supervisorctl restart "$bin_name" 2>&1 | tee -a "$LOG_FILE"
            fi
            return 1
        fi
    else
        log "supervisorctl not found, skipping service restart"
    fi
    
    return 0
}

update_manifest() {
    local bin_name="$1"
    local new_md5="$2"
    
    if [[ ! -f "$BIN_MANIFESTS" ]]; then
        error "Manifests file not found: $BIN_MANIFESTS"
        return 1
    fi
    
    if command -v jq &> /dev/null; then
        local temp_file="/tmp/bin-manifests.tmp.$$"
        jq --arg name "$bin_name" --arg md5 "$new_md5" \
            '(.binaries[] | select(.name == $name) | .currentMd5) = $md5' \
            "$BIN_MANIFESTS" > "$temp_file"
        mv "$temp_file" "$BIN_MANIFESTS"
    else
        log "jq not found, skipping manifest update"
    fi
}

process_binary() {
    local bin_name="$1"
    local current_md5="$2"
    
    log "Processing binary: $bin_name"
    
    local latest_md5
    latest_md5=$(query_latest_md5 "$bin_name")
    
    if [[ -z "$latest_md5" ]]; then
        error "Failed to get latest MD5 for $bin_name"
        return 1
    fi
    
    if update_binary "$bin_name" "$current_md5" "$latest_md5"; then
        update_manifest "$bin_name" "$latest_md5"
        return 0
    fi
    
    return 1
}

main() {
    log "=== Starting bin-proxy ==="
    
    if [[ ! -f "$BIN_MANIFESTS" ]]; then
        error "Manifests file not found: $BIN_MANIFESTS"
        exit 1
    fi
    
    if ! command -v jq &> /dev/null; then
        error "jq is required but not installed"
        exit 1
    fi
    
    local binaries
    binaries=$(jq -r '.binaries[] | "\(.name):\(.currentMd5)"' "$BIN_MANIFESTS")
    
    while IFS=: read -r bin_name current_md5; do
        if [[ -n "$bin_name" ]]; then
            process_binary "$bin_name" "$current_md5" || true
        fi
    done <<< "$binaries"
    
    log "=== bin-proxy completed ==="
}

main "$@"
