#!/bin/bash

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BIN_MANIFESTS="${BIN_MANIFESTS:-$SCRIPT_DIR/bin-manifests.json}"
LOG_FILE="${LOG_FILE:-/var/log/bin-inspector.log}"
CHECK_INTERVAL="${CHECK_INTERVAL:-60}"
MAX_FAILURES="${MAX_FAILURES:-3}"
FAILURE_STATE_FILE="${FAILURE_STATE_FILE:-/var/run/bin-inspector-failures.json}"
PROMETHEUS_URL="${PROMETHEUS_URL:-http://localhost:9090}"

log() {
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] $*" | tee -a "$LOG_FILE"
}

error() {
    log "ERROR: $*" >&2
}

check_process_running() {
    local bin_name="$1"

    if pgrep -x "$bin_name" > /dev/null 2>&1; then
        return 0
    fi

    if pgrep -f "$bin_name" > /dev/null 2>&1; then
        return 0
    fi

    if command -v supervisorctl &> /dev/null; then
        if supervisorctl status "$bin_name" 2>/dev/null | grep -q RUNNING; then
            return 0
        fi
    fi

    return 1
}

check_service_port() {
    local bin_name="$1"
    local port="$2"

    if [[ -z "$port" ]]; then
        log "No port configured for $bin_name, skipping port check"
        return 0
    fi

    if command -v nc &> /dev/null; then
        if nc -z localhost "$port" 2>/dev/null; then
            return 0
        fi
    elif command -v timeout &> /dev/null; then
        if timeout 1 bash -c "cat < /dev/null > /dev/tcp/localhost/$port" 2>/dev/null; then
            return 0
        fi
    fi

    return 1
}

check_prometheus_alerts() {
    local bin_name="$1"

    if [[ -z "$PROMETHEUS_URL" ]]; then
        log "Prometheus URL not configured, skipping alert check"
        return 0
    fi

    local query="ALERTS{alertname=~\".*${bin_name}.*\",alertstate=\"firing\"}"
    local url="${PROMETHEUS_URL}/api/v1/query?query=$(echo "$query" | jq -sRr @uri)"

    local response
    if ! response=$(curl -s -f --max-time 5 "$url" 2>/dev/null); then
        log "Failed to query Prometheus, skipping alert check for $bin_name"
        return 0
    fi

    local alert_count
    alert_count=$(echo "$response" | jq -r '.data.result | length' 2>/dev/null || echo "0")

    if [[ "$alert_count" -gt 0 ]]; then
        error "Prometheus has $alert_count firing alerts for $bin_name"
        return 1
    fi

    return 0
}

get_failure_count() {
    local bin_name="$1"

    if [[ ! -f "$FAILURE_STATE_FILE" ]]; then
        echo "0"
        return
    fi

    if command -v jq &> /dev/null; then
        jq -r --arg name "$bin_name" '.[$name].count // 0' "$FAILURE_STATE_FILE" 2>/dev/null || echo "0"
    else
        echo "0"
    fi
}

increment_failure_count() {
    local bin_name="$1"

    if [[ ! -f "$FAILURE_STATE_FILE" ]]; then
        echo '{}' > "$FAILURE_STATE_FILE"
    fi

    if command -v jq &> /dev/null; then
        local temp_file="/tmp/failures.tmp.$$"
        local timestamp
        timestamp=$(date -u +"%Y-%m-%dT%H:%M:%SZ")

        jq --arg name "$bin_name" --arg ts "$timestamp" \
            '.[$name].count = ((.[$name].count // 0) + 1) | .[$name].lastFailure = $ts' \
            "$FAILURE_STATE_FILE" > "$temp_file"
        mv "$temp_file" "$FAILURE_STATE_FILE"
    fi
}

reset_failure_count() {
    local bin_name="$1"

    if [[ ! -f "$FAILURE_STATE_FILE" ]]; then
        return
    fi

    if command -v jq &> /dev/null; then
        local temp_file="/tmp/failures.tmp.$$"
        jq --arg name "$bin_name" 'del(.[$name])' "$FAILURE_STATE_FILE" > "$temp_file"
        mv "$temp_file" "$FAILURE_STATE_FILE"
    fi
}

get_previous_version() {
    local bin_name="$1"

    if [[ ! -f "$BIN_MANIFESTS" ]]; then
        return 1
    fi

    if command -v jq &> /dev/null; then
        local history
        history=$(jq -r --arg name "$bin_name" \
            '.binaries[] | select(.binaryName == $name) | .versionHistory // []' \
            "$BIN_MANIFESTS" 2>/dev/null)

        if [[ -z "$history" ]] || [[ "$history" == "[]" ]]; then
            error "No version history found for $bin_name"
            return 1
        fi

        local prev_sha256
        prev_sha256=$(echo "$history" | jq -r '.[-2].to // .[-1].from // empty' 2>/dev/null)

        if [[ -z "$prev_sha256" ]] || [[ "$prev_sha256" == "null" ]]; then
            error "Cannot determine previous version for $bin_name"
            return 1
        fi

        echo "$prev_sha256"
        return 0
    fi

    return 1
}

trigger_rollback() {
    local bin_name="$1"

    log "Triggering rollback for $bin_name"

    local prev_version
    if ! prev_version=$(get_previous_version "$bin_name"); then
        error "Failed to get previous version for $bin_name, cannot rollback"
        return 1
    fi

    log "Rolling back $bin_name to previous version: $prev_version"

    if [[ -f "${SCRIPT_DIR}/bin-proxy.sh" ]]; then
        if bash -c "source ${SCRIPT_DIR}/bin-proxy.sh && rollback_binary \"$bin_name\" \"$prev_version\""; then
            log "Rollback successful for $bin_name"
            reset_failure_count "$bin_name"
            return 0
        else
            error "Rollback failed for $bin_name"
            return 1
        fi
    else
        error "bin-proxy.sh not found, cannot perform rollback"
        return 1
    fi
}

inspect_service() {
    local bin_name="$1"
    local port="$2"

    log "Inspecting service: $bin_name"

    local checks_passed=true

    if ! check_process_running "$bin_name"; then
        error "Service $bin_name is not running"
        checks_passed=false
    else
        log "✓ Process check passed for $bin_name"
    fi

    if ! check_service_port "$bin_name" "$port"; then
        error "Service $bin_name port $port is not accessible"
        checks_passed=false
    else
        log "✓ Port check passed for $bin_name"
    fi

    if ! check_prometheus_alerts "$bin_name"; then
        error "Service $bin_name has Prometheus alerts"
        checks_passed=false
    else
        log "✓ Prometheus check passed for $bin_name"
    fi

    if [[ "$checks_passed" == false ]]; then
        increment_failure_count "$bin_name"
        local failure_count
        failure_count=$(get_failure_count "$bin_name")

        error "Service $bin_name failed health checks (failure count: $failure_count/$MAX_FAILURES)"

        if [[ $failure_count -ge $MAX_FAILURES ]]; then
            error "Service $bin_name has failed $failure_count times, triggering rollback"
            if trigger_rollback "$bin_name"; then
                log "Rollback completed for $bin_name"
            else
                error "Rollback failed for $bin_name"
            fi
        fi

        return 1
    else
        reset_failure_count "$bin_name"
        log "Service $bin_name is healthy"
        return 0
    fi
}

main() {
    log "=== Starting bin-inspector ==="

    if [[ ! -f "$BIN_MANIFESTS" ]]; then
        error "Manifests file not found: $BIN_MANIFESTS"
        exit 1
    fi

    if ! command -v jq &> /dev/null; then
        error "jq is required but not installed"
        exit 1
    fi

    local services
    services=$(jq -r '.binaries[] | "\(.binaryName):\(.port // "")"' "$BIN_MANIFESTS" 2>/dev/null)

    if [[ -z "$services" ]]; then
        error "No services found in manifests"
        exit 1
    fi

    while IFS=: read -r bin_name port; do
        if [[ -n "$bin_name" ]]; then
            inspect_service "$bin_name" "$port" || true
        fi
    done <<< "$services"

    log "=== bin-inspector completed ==="
}

if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    main "$@"
fi
