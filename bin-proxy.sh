#!/bin/bash

PROXY_VERSION="1.0.0"
MANIFEST_FILE="/etc/bin-proxy/bin-manifests.json"
API_BASE_URL="${BIN_MANAGER_API:-http://localhost:8081/api/v1}"
DOWNLOAD_URL="${BIN_DOWNLOAD_URL:-http://localhost:8081/api/v1/download}"
LOG_FILE="/var/log/bin-proxy.log"

log() {
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] $1" | tee -a "$LOG_FILE"
}

error_exit() {
    log "ERROR: $1"
    exit 1
}

get_node_info() {
    local node_name=$(hostname)
    local cpu_arch=$(uname -m)
    local os_release=$(cat /etc/os-release | grep "^PRETTY_NAME" | cut -d'"' -f2)
    
    echo "{\"name\":\"$node_name\",\"cpuArch\":\"$cpu_arch\",\"osRelease\":\"$os_release\",\"proxyVersion\":\"$PROXY_VERSION\"}"
}

load_manifest() {
    if [ ! -f "$MANIFEST_FILE" ]; then
        error_exit "Manifest file not found: $MANIFEST_FILE"
    fi
    
    cat "$MANIFEST_FILE"
}

calculate_hash() {
    local file="$1"
    if [ ! -f "$file" ]; then
        echo ""
        return
    fi
    md5sum "$file" | awk '{print $1}'
}

keepalive() {
    local node_name=$(hostname)
    local response=$(curl -s -w "\n%{http_code}" "$API_BASE_URL/keepalive?node=$node_name")
    local http_code=$(echo "$response" | tail -n1)
    local body=$(echo "$response" | sed '$d')
    
    if [ "$http_code" == "404" ]; then
        log "Node not registered, registering..."
        local node_info=$(get_node_info)
        local manifest=$(load_manifest)
        local binaries=$(echo "$manifest" | jq -c '.binaries')
        
        local full_node_info=$(echo "$node_info" | jq --argjson bins "$binaries" '. + {binaries: $bins, lastSeen: (now | strftime("%Y-%m-%dT%H:%M:%SZ"))}')
        
        response=$(curl -s -X POST -H "Content-Type: application/json" \
            -d "$full_node_info" \
            -w "\n%{http_code}" \
            "$API_BASE_URL/keepalive?node=$node_name")
        
        http_code=$(echo "$response" | tail -n1)
        if [ "$http_code" != "200" ]; then
            log "Failed to register node. HTTP code: $http_code"
            return 1
        fi
        log "Node registered successfully"
    fi
    
    return 0
}

check_and_upgrade() {
    local bin_name="$1"
    local current_hash="$2"
    local bin_path="$3"
    
    log "Checking $bin_name for updates..."
    
    local response=$(curl -s -w "\n%{http_code}" "$API_BASE_URL/bins/$bin_name")
    local http_code=$(echo "$response" | tail -n1)
    local body=$(echo "$response" | sed '$d')
    
    if [ "$http_code" != "200" ]; then
        log "Failed to get hash for $bin_name. HTTP code: $http_code"
        return 1
    fi
    
    local remote_hash=$(echo "$body" | jq -r '.data.hash')
    
    if [ "$current_hash" == "$remote_hash" ]; then
        log "$bin_name is up to date (hash: $current_hash)"
        return 0
    fi
    
    log "$bin_name needs upgrade. Current: $current_hash, Remote: $remote_hash"
    
    local temp_file="/tmp/${bin_name}.tmp"
    log "Downloading $bin_name from $DOWNLOAD_URL/$bin_name"
    
    if ! curl -s -f -o "$temp_file" "$DOWNLOAD_URL/$bin_name"; then
        log "Failed to download $bin_name"
        rm -f "$temp_file"
        return 1
    fi
    
    local downloaded_hash=$(calculate_hash "$temp_file")
    
    if [ "$downloaded_hash" != "$remote_hash" ]; then
        log "Hash mismatch for downloaded $bin_name. Expected: $remote_hash, Got: $downloaded_hash"
        rm -f "$temp_file"
        return 1
    fi
    
    if [ -f "$bin_path" ]; then
        cp "$bin_path" "${bin_path}.backup"
    fi
    
    mv "$temp_file" "$bin_path"
    chmod +x "$bin_path"
    
    log "$bin_name upgraded successfully to hash: $downloaded_hash"
    
    if command -v supervisorctl &> /dev/null; then
        log "Restarting service: $bin_name via supervisor"
        supervisorctl restart "$bin_name" || log "Failed to restart $bin_name"
    else
        log "Supervisor not found, skipping service restart"
    fi
    
    response=$(curl -s -X POST -H "Content-Type: application/json" \
        -d "{\"hash\":\"$downloaded_hash\"}" \
        -w "\n%{http_code}" \
        "$API_BASE_URL/bins/$bin_name")
    
    http_code=$(echo "$response" | tail -n1)
    if [ "$http_code" == "200" ]; then
        log "Updated hash reported to bin-manager"
    fi
    
    local temp_manifest=$(mktemp)
    jq --arg name "$bin_name" --arg hash "$downloaded_hash" \
        '(.binaries[] | select(.name == $name) | .hash) = $hash' \
        "$MANIFEST_FILE" > "$temp_manifest"
    mv "$temp_manifest" "$MANIFEST_FILE"
    
    return 0
}

main() {
    log "=== bin-proxy started (version: $PROXY_VERSION) ==="
    
    if ! command -v jq &> /dev/null; then
        error_exit "jq is required but not installed"
    fi
    
    keepalive
    
    local manifest=$(load_manifest)
    
    echo "$manifest" | jq -c '.binaries[]' | while read -r bin; do
        local name=$(echo "$bin" | jq -r '.name')
        local hash=$(echo "$bin" | jq -r '.hash')
        local path=$(echo "$bin" | jq -r '.path // "/usr/local/bin/\($name)"')
        
        check_and_upgrade "$name" "$hash" "$path"
    done
    
    log "=== bin-proxy completed ==="
}

main "$@"
