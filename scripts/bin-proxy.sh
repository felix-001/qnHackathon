#!/bin/bash

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BIN_MANIFESTS="${BIN_MANIFESTS:-$SCRIPT_DIR/bin-manifests.json}"
BIN_MANAGER_API="${BIN_MANAGER_API:-http://localhost:8080/api/v1}"
BIN_DIR="${BIN_DIR:-/usr/local/bin}"
LOG_FILE="${LOG_FILE:-/var/log/bin-proxy.log}"
LOCK_DIR="${LOCK_DIR:-/var/run/bin-proxy}"
BIN_PROXY_VERSION="1.1.0"
DOWNLOAD_BASE_URL="${DOWNLOAD_BASE_URL:-${BIN_MANAGER_API}/download}"

mkdir -p "$LOCK_DIR"

log() {
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] $*" | tee -a "$LOG_FILE"
}

error() {
    log "ERROR: $*" >&2
}

get_node_info() {
    local cpu_arch
    cpu_arch=$(uname -m)
    
    local os_release
    if [[ -f /etc/os-release ]]; then
        os_release=$(grep '^PRETTY_NAME=' /etc/os-release | cut -d'=' -f2 | tr -d '"')
    else
        os_release=$(uname -s)
    fi
    
    local node_name
    node_name=$(hostname)
    
    echo "{\"cpuArch\":\"$cpu_arch\",\"osRelease\":\"$os_release\",\"nodeName\":\"$node_name\",\"binProxyVersion\":\"$BIN_PROXY_VERSION\"}"
}

keepalive_check() {
    local node_info
    node_info=$(get_node_info)
    
    local response
    response=$(curl -s -f "${BIN_MANAGER_API}/keepalive" 2>/dev/null)
    
    if [[ $? -ne 0 ]] || [[ -z "$response" ]]; then
        log "Node not registered, posting node info"
        curl -s -X POST -H "Content-Type: application/json" \
            -d "$node_info" \
            "${BIN_MANAGER_API}/keepalive" 2>/dev/null || true
    else
        log "Keepalive check successful"
    fi
}

acquire_lock() {
    local bin_name="$1"
    local bin_hash="$2"
    local lock_file="${LOCK_DIR}/${bin_name}-${bin_hash}.lock"
    
    local old_locks
    old_locks=$(ls "${LOCK_DIR}/${bin_name}"-*.lock 2>/dev/null || true)
    if [[ -n "$old_locks" ]]; then
        for old_lock in $old_locks; do
            if [[ "$old_lock" != "$lock_file" ]]; then
                log "Removing old lock file: $old_lock"
                rm -f "$old_lock"
            fi
        done
    fi
    
    if [[ -f "$lock_file" ]]; then
        local lock_time
        lock_time=$(cat "$lock_file")
        local current_time
        current_time=$(date +%s)
        local elapsed=$((current_time - lock_time))
        
        if [[ $elapsed -lt 600 ]]; then
            log "Lock exists for $bin_name-$bin_hash (held for ${elapsed}s), skipping"
            return 1
        else
            log "Stale lock detected for $bin_name-$bin_hash (held for ${elapsed}s), removing"
            rm -f "$lock_file"
        fi
    fi
    
    date +%s > "$lock_file"
    log "Acquired lock for $bin_name-$bin_hash"
    return 0
}

release_lock() {
    local bin_name="$1"
    local bin_hash="$2"
    local lock_file="${LOCK_DIR}/${bin_name}-${bin_hash}.lock"
    rm -f "$lock_file"
    log "Released lock for $bin_name-$bin_hash"
}

report_progress() {
    local bin_name="$1"
    local bin_hash="$2"
    local lock_file="${LOCK_DIR}/${bin_name}-${bin_hash}.lock"
    
    if [[ ! -f "$lock_file" ]]; then
        return
    fi
    
    local lock_time
    lock_time=$(cat "$lock_file")
    local current_time
    current_time=$(date +%s)
    local elapsed=$((current_time - lock_time))
    
    local node_name
    node_name=$(hostname)
    
    local payload
    payload="{\"nodeName\":\"$node_name\",\"binName\":\"$bin_name\",\"targetHash\":\"$bin_hash\",\"processingTime\":$elapsed,\"status\":\"in_progress\"}"
    
    curl -s -X POST -H "Content-Type: application/json" \
        -d "$payload" \
        "${BIN_MANAGER_API}/bins/${bin_name}/progress" 2>/dev/null || true
}

report_completion() {
    local bin_name="$1"
    local bin_hash="$2"
    local status="$3"
    local lock_file="${LOCK_DIR}/${bin_name}-${bin_hash}.lock"
    
    if [[ ! -f "$lock_file" ]]; then
        return
    fi
    
    local lock_time
    lock_time=$(cat "$lock_file")
    local current_time
    current_time=$(date +%s)
    local elapsed=$((current_time - lock_time))
    
    local node_name
    node_name=$(hostname)
    
    local payload
    payload="{\"nodeName\":\"$node_name\",\"binName\":\"$bin_name\",\"targetHash\":\"$bin_hash\",\"processingTime\":$elapsed,\"status\":\"$status\"}"
    
    curl -s -X POST -H "Content-Type: application/json" \
        -d "$payload" \
        "${BIN_MANAGER_API}/bins/${bin_name}/progress" 2>/dev/null || true
    
    log "Reported completion for $bin_name-$bin_hash: $status (took ${elapsed}s)"
}

kill_old_downloads() {
    local bin_name="$1"
    
    local pids
    pids=$(pgrep -f "curl.*${bin_name}" | grep -v "$$" || true)
    
    if [[ -n "$pids" ]]; then
        log "Killing old download processes for $bin_name: $pids"
        echo "$pids" | xargs kill -9 2>/dev/null || true
    fi
}

post_update_status() {
    local bin_name="$1"
    local new_md5="$2"
    
    local node_name
    node_name=$(hostname)
    
    local payload
    payload="{\"nodeName\":\"$node_name\",\"binName\":\"$bin_name\",\"md5\":\"$new_md5\",\"version\":\"latest\"}"
    
    if curl -s -X POST -H "Content-Type: application/json" \
        -d "$payload" \
        "${BIN_MANAGER_API}/bins/${bin_name}" 2>/dev/null; then
        log "Posted update status for $bin_name to API"
        return 0
    else
        error "Failed to post update status for $bin_name"
        return 1
    fi
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
    local url="${BIN_MANAGER_API}/bins/${bin_name}"
    
    local response
    response=$(curl -s -f "$url" 2>/dev/null)
    
    if [[ $? -ne 0 ]]; then
        log "Trying alternative endpoint for $bin_name"
        url="${BIN_MANAGER_API}/releases/latest/${bin_name}/md5"
        response=$(curl -s -f "$url" 2>/dev/null)
        if [[ $? -ne 0 ]]; then
            error "Failed to query latest MD5 for $bin_name"
            echo ""
            return 1
        fi
    fi
    
    echo "$response" | grep -o '"md5":"[^"]*"' | cut -d'"' -f4
}

download_binary() {
    local bin_name="$1"
    local temp_file="$2"
    local url="${DOWNLOAD_BASE_URL}/${bin_name}"
    
    log "Downloading $bin_name from $url"
    
    if ! curl -s -f -o "$temp_file" "$url"; then
        log "Trying alternative download endpoint"
        url="${BIN_MANAGER_API}/releases/latest/${bin_name}/download"
        if ! curl -s -f -o "$temp_file" "$url"; then
            error "Failed to download $bin_name"
            return 1
        fi
    fi
    
    chmod +x "$temp_file"
    return 0
}

update_binary() {
    local bin_name="$1"
    local current_md5="$2"
    local latest_md5="$3"
    
    if [[ "$current_md5" == "$latest_md5" ]] && [[ -n "$current_md5" ]]; then
        log "$bin_name is already up to date (MD5: $current_md5)"
        return 0
    fi
    
    log "Updating $bin_name (current: $current_md5, latest: $latest_md5)"
    
    kill_old_downloads "$bin_name"
    
    local bin_path="${BIN_DIR}/${bin_name}"
    local temp_file="/tmp/${bin_name}.tmp.$$"
    
    report_progress "$bin_name" "$latest_md5"
    
    if ! download_binary "$bin_name" "$temp_file"; then
        report_completion "$bin_name" "$latest_md5" "failed"
        return 1
    fi
    
    local downloaded_md5
    downloaded_md5=$(get_md5sum "$temp_file")
    
    if [[ "$downloaded_md5" != "$latest_md5" ]]; then
        error "MD5 mismatch for downloaded $bin_name (expected: $latest_md5, got: $downloaded_md5)"
        rm -f "$temp_file"
        report_completion "$bin_name" "$latest_md5" "failed"
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
            report_completion "$bin_name" "$latest_md5" "success"
        else
            error "Failed to restart service $bin_name"
            if [[ -f "${bin_path}.backup" ]]; then
                log "Rolling back to previous version"
                mv "${bin_path}.backup" "$bin_path"
                supervisorctl restart "$bin_name" 2>&1 | tee -a "$LOG_FILE"
            fi
            report_completion "$bin_name" "$latest_md5" "failed"
            return 1
        fi
    else
        log "supervisorctl not found, skipping service restart"
        report_completion "$bin_name" "$latest_md5" "success"
    fi
    
    post_update_status "$bin_name" "$latest_md5"
    
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
    
    if ! acquire_lock "$bin_name" "$latest_md5"; then
        return 1
    fi
    
    if update_binary "$bin_name" "$current_md5" "$latest_md5"; then
        update_manifest "$bin_name" "$latest_md5"
        release_lock "$bin_name" "$latest_md5"
        return 0
    fi
    
    release_lock "$bin_name" "$latest_md5"
    return 1
}

update_node_info() {
    if [[ ! -f "$BIN_MANIFESTS" ]]; then
        return 1
    fi
    
    local node_info
    node_info=$(get_node_info)
    
    if command -v jq &> /dev/null; then
        local temp_file="/tmp/bin-manifests.tmp.$$"
        echo "$node_info" | jq -r '.' > /tmp/node-info.tmp
        jq --slurpfile nodeinfo /tmp/node-info.tmp '.nodeInfo = $nodeinfo[0]' \
            "$BIN_MANIFESTS" > "$temp_file"
        mv "$temp_file" "$BIN_MANIFESTS"
        rm -f /tmp/node-info.tmp
    fi
}

main() {
    log "=== Starting bin-proxy v${BIN_PROXY_VERSION} ==="
    
    if [[ ! -f "$BIN_MANIFESTS" ]]; then
        error "Manifests file not found: $BIN_MANIFESTS"
        exit 1
    fi
    
    if ! command -v jq &> /dev/null; then
        error "jq is required but not installed"
        exit 1
    fi
    
    update_node_info
    
    keepalive_check
    
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
