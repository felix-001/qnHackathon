#!/bin/bash

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BIN_MANIFESTS="${BIN_MANIFESTS:-$SCRIPT_DIR/bin-manifests.json}"
BIN_MANAGER_API="${BIN_MANAGER_API:-http://localhost:8080/api/v1}"
BIN_DIR="${BIN_DIR:-/usr/local/bin}"
LOG_FILE="${LOG_FILE:-/var/log/bin-proxy.log}"
LOCK_DIR="${LOCK_DIR:-/var/run/bin-proxy}"
LOCK_TIMEOUT="${LOCK_TIMEOUT:-600}"
BIN_PROXY_VERSION="1.2.0"
DOWNLOAD_BASE_URL="${DOWNLOAD_BASE_URL:-${BIN_MANAGER_API}/download}"
DOWNLOAD_TIMEOUT="${DOWNLOAD_TIMEOUT:-300}"

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
        for old_lock in "$old_locks"; do
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

        if [[ $elapsed -lt $LOCK_TIMEOUT ]]; then
            log "Lock exists for $bin_name-$bin_hash (held for ${elapsed}s), skipping"
            return 1
        else
            log "Stale lock detected for $bin_name-$bin_hash (held for ${elapsed}s), removing"
            rm -f "$lock_file"
        fi
    fi

    # Atomic lock acquisition to prevent race condition
    local current_time
    current_time=$(date +%s)
    if (set -o noclobber; echo "$current_time" > "$lock_file") 2>/dev/null; then
        log "Acquired lock for $bin_name-$bin_hash"
        return 0
    else
        log "Failed to acquire lock for $bin_name-$bin_hash"
        return 1
    fi
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
    pids=$(pgrep -f "curl.*${DOWNLOAD_BASE_URL}/${bin_name}$" | grep -v "$$" || true)

    if [[ -n "$pids" ]]; then
        log "Killing old download processes for $bin_name: $pids"
        echo "$pids" | xargs kill -9 2>/dev/null || true
    fi
}

post_update_status() {
    local bin_name="$1"
    local new_sha256="$2"

    local node_name
    node_name=$(hostname)

    local payload
    payload="{\"nodeName\":\"$node_name\",\"binName\":\"$bin_name\",\"sha256\":\"$new_sha256\",\"version\":\"latest\"}"

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

get_sha256sum() {
    local file="$1"
    if [[ ! -f "$file" ]]; then
        echo ""
        return
    fi
    sha256sum "$file" | awk '{print $1}'
}

query_latest_sha256() {
    local bin_name="$1"

    if [[ -z "$bin_name" ]]; then
        error "bin_name is required for query_latest_sha256"
        return 1
    fi

    local url="${BIN_MANAGER_API}/bins/${bin_name}"

    local response
    response=$(curl -s -f "$url" 2>/dev/null)

    if [[ $? -ne 0 ]]; then
        log "Trying alternative endpoint for $bin_name"
        url="${BIN_MANAGER_API}/releases/latest/${bin_name}/sha256"
        response=$(curl -s -f "$url" 2>/dev/null)
        if [[ $? -ne 0 ]]; then
            error "Failed to query latest SHA256 for $bin_name"
            echo ""
            return 1
        fi
    fi

    # Validate JSON response
    if ! echo "$response" | jq -e '.sha256' > /dev/null 2>&1; then
        error "Invalid API response format for $bin_name (sha256 field missing)"
        return 1
    fi
    
    echo "$response" | jq -r '.sha256'
}

download_binary() {
    local bin_name="$1"
    local temp_file="$2"

    if [[ -z "$bin_name" ]] || [[ -z "$temp_file" ]]; then
        error "bin_name and temp_file are required for download_binary"
        return 1
    fi

    # Check disk space before download (require at least 100MB free)
    local required_space=102400  # 100MB in KB
    local available
    available=$(df /tmp | tail -1 | awk '{print $4}')
    if [[ $available -lt $required_space ]]; then
        error "Insufficient disk space in /tmp (available: ${available}KB, required: ${required_space}KB)"
        return 1
    fi

    local url="${DOWNLOAD_BASE_URL}/${bin_name}"

    log "Downloading $bin_name from $url"

    if ! curl -s -f --max-time "$DOWNLOAD_TIMEOUT" -o "$temp_file" "$url"; then
        log "Trying alternative download endpoint"
        url="${BIN_MANAGER_API}/releases/latest/${bin_name}/download"
        if ! curl -s -f --max-time "$DOWNLOAD_TIMEOUT" -o "$temp_file" "$url"; then
            error "Failed to download $bin_name"
            return 1
        fi
    fi

    chmod +x "$temp_file"
    return 0
}

change_version() {
    local bin_name="$1"
    local target_sha256="$2"
    local operation="${3:-upgrade}"

    if [[ -z "$bin_name" ]] || [[ -z "$target_sha256" ]]; then
        error "bin_name and target_sha256 are required for change_version"
        return 1
    fi

    local bin_path="${BIN_DIR}/${bin_name}"
    local bin_archive_dir="${BIN_DIR}/.archive/${bin_name}"
    mkdir -p "$bin_archive_dir"

    log "[$operation] Changing $bin_name to version (SHA256: $target_sha256)"

    local source_file=""

    if [[ "$operation" == "rollback" ]]; then
        source_file="${bin_archive_dir}/${target_sha256}"
        if [[ ! -f "$source_file" ]]; then
            error "Rollback failed: archived binary not found for SHA256: $target_sha256"
            return 1
        fi
        log "Using archived binary for rollback: $source_file"
    else
        kill_old_downloads "$bin_name"
        source_file="/tmp/${bin_name}.tmp.$$"
        report_progress "$bin_name" "$target_sha256"

        if ! download_binary "$bin_name" "$source_file"; then
            report_completion "$bin_name" "$target_sha256" "failed"
            return 1
        fi

        local downloaded_sha256
        downloaded_sha256=$(get_sha256sum "$source_file")

        if [[ "$downloaded_sha256" != "$target_sha256" ]]; then
            error "SHA256 mismatch for downloaded $bin_name (expected: $target_sha256, got: $downloaded_sha256)"
            rm -f "$source_file"
            report_completion "$bin_name" "$target_sha256" "failed"
            return 1
        fi
    fi

    local current_sha256=""
    if [[ -f "$bin_path" ]]; then
        current_sha256=$(get_sha256sum "$bin_path")
        if [[ -n "$current_sha256" ]]; then
            local archive_path="${bin_archive_dir}/${current_sha256}"
            if [[ ! -f "$archive_path" ]]; then
                cp "$bin_path" "$archive_path"
                chmod +x "$archive_path"
                log "Archived current binary: $archive_path"
            fi
        fi
    fi

    if [[ "$operation" == "upgrade" ]]; then
        cp "$source_file" "$bin_path"
        rm -f "$source_file"
    else
        cp "$source_file" "$bin_path"
    fi
    chmod +x "$bin_path"

    log "Binary file replaced successfully"

    if command -v supervisorctl &> /dev/null; then
        log "Restarting service: $bin_name via supervisor"
        if supervisorctl restart "$bin_name" 2>&1 | tee -a "$LOG_FILE"; then
            log "Service $bin_name restarted successfully"
            sleep 2
            if supervisorctl status "$bin_name" | grep -q RUNNING; then
                log "Service $bin_name verified running after restart"
                if [[ "$operation" == "upgrade" ]]; then
                    report_completion "$bin_name" "$target_sha256" "success"
                    post_update_status "$bin_name" "$target_sha256"
                fi
                record_version_change "$bin_name" "$current_sha256" "$target_sha256" "$operation" "success"
                return 0
            else
                error "Service $bin_name not running after restart"
                if [[ -n "$current_sha256" ]]; then
                    log "Auto-rollback: restoring previous version"
                    local rollback_source="${bin_archive_dir}/${current_sha256}"
                    if [[ -f "$rollback_source" ]]; then
                        cp "$rollback_source" "$bin_path"
                        if supervisorctl restart "$bin_name" 2>&1 | tee -a "$LOG_FILE"; then
                            log "Auto-rollback successful, service restarted"
                            record_version_change "$bin_name" "$target_sha256" "$current_sha256" "auto-rollback" "success"
                        else
                            error "Auto-rollback failed - service may be down"
                            record_version_change "$bin_name" "$target_sha256" "$current_sha256" "auto-rollback" "failed"
                        fi
                    fi
                fi
                if [[ "$operation" == "upgrade" ]]; then
                    report_completion "$bin_name" "$target_sha256" "failed"
                fi
                record_version_change "$bin_name" "$current_sha256" "$target_sha256" "$operation" "failed"
                return 1
            fi
        else
            error "Failed to restart service $bin_name"
            if [[ -n "$current_sha256" ]]; then
                log "Auto-rollback: restoring previous version"
                local rollback_source="${bin_archive_dir}/${current_sha256}"
                if [[ -f "$rollback_source" ]]; then
                    cp "$rollback_source" "$bin_path"
                    if supervisorctl restart "$bin_name" 2>&1 | tee -a "$LOG_FILE"; then
                        log "Auto-rollback successful, service restarted"
                        record_version_change "$bin_name" "$target_sha256" "$current_sha256" "auto-rollback" "success"
                    else
                        error "Auto-rollback failed - service may be down"
                        record_version_change "$bin_name" "$target_sha256" "$current_sha256" "auto-rollback" "failed"
                    fi
                fi
            fi
            if [[ "$operation" == "upgrade" ]]; then
                report_completion "$bin_name" "$target_sha256" "failed"
            fi
            record_version_change "$bin_name" "$current_sha256" "$target_sha256" "$operation" "failed"
            return 1
        fi
    else
        log "supervisorctl not found, skipping service restart"
        if [[ "$operation" == "upgrade" ]]; then
            report_completion "$bin_name" "$target_sha256" "success"
            post_update_status "$bin_name" "$target_sha256"
        fi
        record_version_change "$bin_name" "$current_sha256" "$target_sha256" "$operation" "success"
    fi

    return 0
}

record_version_change() {
    local bin_name="$1"
    local from_sha256="$2"
    local to_sha256="$3"
    local operation="$4"
    local result="$5"

    if [[ ! -f "$BIN_MANIFESTS" ]]; then
        return 1
    fi

    if command -v jq &> /dev/null; then
        local temp_file="/tmp/bin-manifests.tmp.$$"
        local timestamp
        timestamp=$(date -u +"%Y-%m-%dT%H:%M:%SZ")

        jq --arg name "$bin_name" \
           --arg from "$from_sha256" \
           --arg to "$to_sha256" \
           --arg op "$operation" \
           --arg res "$result" \
           --arg ts "$timestamp" '
            (.binaries[] | select(.name == $name)) |= (
                .currentSha256 = $to |
                .versionHistory //= [] |
                .versionHistory += [{
                    "from": $from,
                    "to": $to,
                    "operation": $op,
                    "result": $res,
                    "timestamp": $ts
                }]
            )' "$BIN_MANIFESTS" > "$temp_file"
        mv "$temp_file" "$BIN_MANIFESTS"
        log "Recorded version change: $bin_name ($operation: $from_sha256 -> $to_sha256, result: $result)"
    fi
}

update_binary() {
    local bin_name="$1"
    local current_sha256="$2"
    local latest_sha256="$3"

    if [[ -z "$bin_name" ]]; then
        error "bin_name is required for update_binary"
        return 1
    fi

    if [[ "$current_sha256" == "$latest_sha256" ]] && [[ -n "$current_sha256" ]]; then
        log "$bin_name is already up to date (SHA256: $current_sha256)"
        return 0
    fi

    log "Updating $bin_name (current: $current_sha256, latest: $latest_sha256)"

    change_version "$bin_name" "$latest_sha256" "upgrade"
    return $?
}

rollback_binary() {
    local bin_name="$1"
    local target_sha256="$2"

    if [[ -z "$bin_name" ]] || [[ -z "$target_sha256" ]]; then
        error "bin_name and target_sha256 are required for rollback_binary"
        return 1
    fi

    log "Rolling back $bin_name to SHA256: $target_sha256"

    change_version "$bin_name" "$target_sha256" "rollback"
    return $?
}

update_manifest() {
    local bin_name="$1"
    local new_sha256="$2"

    if [[ ! -f "$BIN_MANIFESTS" ]]; then
        error "Manifests file not found: $BIN_MANIFESTS"
        return 1
    fi

    if command -v jq &> /dev/null; then
        local temp_file="/tmp/bin-manifests.tmp.$$"
        jq --arg name "$bin_name" --arg sha256 "$new_sha256" \
            '(.binaries[] | select(.name == $name) | .currentSha256) = $sha256' \
            "$BIN_MANIFESTS" > "$temp_file"
        mv "$temp_file" "$BIN_MANIFESTS"
    else
        log "jq not found, skipping manifest update"
    fi
}

process_binary() {
    local bin_name="$1"
    local current_sha256="$2"

    if [[ -z "$bin_name" ]]; then
        error "bin_name is required for process_binary"
        return 1
    fi

    log "Processing binary: $bin_name"

    local latest_sha256
    latest_sha256=$(query_latest_sha256 "$bin_name")

    if [[ -z "$latest_sha256" ]]; then
        error "Failed to get latest SHA256 for $bin_name"
        return 1
    fi

    if ! acquire_lock "$bin_name" "$latest_sha256"; then
        return 1
    fi

    if update_binary "$bin_name" "$current_sha256" "$latest_sha256"; then
        update_manifest "$bin_name" "$latest_sha256"
        release_lock "$bin_name" "$latest_sha256"
        return 0
    fi

    release_lock "$bin_name" "$latest_sha256"
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
    binaries=$(jq -r '.binaries[] | "\(.name):\(.currentSha256 // "")"' "$BIN_MANIFESTS")

    while IFS=: read -r bin_name current_sha256; do
        if [[ -n "$bin_name" ]]; then
            process_binary "$bin_name" "$current_sha256" || true
        fi
    done <<< "$binaries"

    log "=== bin-proxy completed ==="
}

main "$@"
