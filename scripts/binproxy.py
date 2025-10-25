#!/usr/bin/python3

import os
import sys
import json
import hashlib
import time
import subprocess
import platform
import socket
import shutil
import logging
import fcntl
import tempfile
from typing import Dict, List, Optional, Tuple
from datetime import datetime
from pathlib import Path

try:
    import requests
except ImportError:
    print("ERROR: requests library is required. Install with: pip install requests")
    sys.exit(1)

BIN_MANIFESTS = os.getenv(
    "BIN_MANIFESTS", os.path.join(os.path.dirname(__file__), "bin-manifests.json")
)
BIN_MANAGER_API = os.getenv("BIN_MANAGER_API", "http://localhost:8080/api/v1")
BIN_DIR = os.getenv("BIN_DIR", "/usr/local/bin")
LOG_FILE = os.getenv("LOG_FILE", "/var/log/bin-proxy.log")
LOCK_DIR = os.getenv("LOCK_DIR", "/var/run/bin-proxy")
LOCK_TIMEOUT = int(os.getenv("LOCK_TIMEOUT", "600"))
BIN_PROXY_VERSION = "1.2.0"
DOWNLOAD_BASE_URL = os.getenv("DOWNLOAD_BASE_URL", f"{BIN_MANAGER_API}/download")
DOWNLOAD_TIMEOUT = int(os.getenv("DOWNLOAD_TIMEOUT", "300"))

logging.basicConfig(
    level=logging.INFO,
    format="[%(asctime)s] %(message)s",
    datefmt="%Y-%m-%d %H:%M:%S",
    handlers=[logging.FileHandler(LOG_FILE), logging.StreamHandler()],
)
logger = logging.getLogger(__name__)

os.makedirs(LOCK_DIR, exist_ok=True)


def log(message: str):
    logger.info(message)


def error(message: str):
    logger.error(f"ERROR: {message}")


def get_node_info() -> Dict[str, str]:
    cpu_arch = platform.machine()

    os_release = platform.system()
    if os.path.exists("/etc/os-release"):
        try:
            with open("/etc/os-release", "r") as f:
                for line in f:
                    if line.startswith("PRETTY_NAME="):
                        os_release = line.split("=", 1)[1].strip().strip('"')
                        break
        except Exception:
            pass

    node_name = socket.gethostname()

    return {
        "cpuArch": cpu_arch,
        "osRelease": os_release,
        "nodeName": node_name,
        "binProxyVersion": BIN_PROXY_VERSION,
    }


def keepalive_check():
    node_info = get_node_info()
    node_id = socket.gethostname()
    node_info["node_id"] = node_id
    
    try:
        response = requests.get(
            f"{BIN_MANAGER_API}/keepalive",
            params={"node_id": node_id},
            timeout=10,
        )
        if response.status_code != 200:
            raise Exception("Not registered")
        log("Keepalive check successful")
    except Exception:
        log("Node not registered, posting node info")
        try:
            requests.post(
                f"{BIN_MANAGER_API}/keepalive",
                json=node_info,
                headers={"Content-Type": "application/json"},
                timeout=10,
            )
        except Exception as e:
            error(f"Failed to post keepalive: {e}")


def acquire_lock(bin_name: str, bin_hash: str) -> bool:
    lock_file = Path(LOCK_DIR) / f"{bin_name}-{bin_hash}.lock"

    old_locks = list(Path(LOCK_DIR).glob(f"{bin_name}-*.lock"))
    for old_lock in old_locks:
        if old_lock != lock_file:
            log(f"Removing old lock file: {old_lock}")
            old_lock.unlink(missing_ok=True)

    if lock_file.exists():
        try:
            lock_time = int(lock_file.read_text().strip())
            current_time = int(time.time())
            elapsed = current_time - lock_time

            if elapsed < LOCK_TIMEOUT:
                log(
                    f"Lock exists for {bin_name}-{bin_hash} (held for {elapsed}s), skipping"
                )
                return False
            else:
                log(
                    f"Stale lock detected for {bin_name}-{bin_hash} (held for {elapsed}s), removing"
                )
                lock_file.unlink(missing_ok=True)
        except Exception as e:
            error(f"Error checking lock: {e}")
            lock_file.unlink(missing_ok=True)

    try:
        current_time = int(time.time())
        with open(lock_file, "x") as f:
            fcntl.flock(f.fileno(), fcntl.LOCK_EX | fcntl.LOCK_NB)
            f.write(str(current_time))
            fcntl.flock(f.fileno(), fcntl.LOCK_UN)
        log(f"Acquired lock for {bin_name}-{bin_hash}")
        return True
    except FileExistsError:
        log(f"Failed to acquire lock for {bin_name}-{bin_hash}")
        return False
    except Exception as e:
        error(f"Error acquiring lock: {e}")
        return False


def release_lock(bin_name: str, bin_hash: str):
    lock_file = Path(LOCK_DIR) / f"{bin_name}-{bin_hash}.lock"
    lock_file.unlink(missing_ok=True)
    log(f"Released lock for {bin_name}-{bin_hash}")


def report_progress(bin_name: str, bin_hash: str):
    lock_file = Path(LOCK_DIR) / f"{bin_name}-{bin_hash}.lock"

    if not lock_file.exists():
        return

    try:
        lock_time = int(lock_file.read_text().strip())
        current_time = int(time.time())
        elapsed = current_time - lock_time

        node_name = socket.gethostname()

        payload = {
            "nodeName": node_name,
            "binName": bin_name,
            "targetHash": bin_hash,
            "processingTime": elapsed,
            "status": "in_progress",
        }

        requests.post(
            f"{BIN_MANAGER_API}/bins/{bin_name}/progress",
            json=payload,
            headers={"Content-Type": "application/json"},
            timeout=10,
        )
    except Exception:
        pass


def report_completion(bin_name: str, bin_hash: str, status: str):
    lock_file = Path(LOCK_DIR) / f"{bin_name}-{bin_hash}.lock"

    if not lock_file.exists():
        return

    try:
        lock_time = int(lock_file.read_text().strip())
        current_time = int(time.time())
        elapsed = current_time - lock_time

        node_name = socket.gethostname()

        payload = {
            "nodeName": node_name,
            "binName": bin_name,
            "targetHash": bin_hash,
            "processingTime": elapsed,
            "status": status,
        }

        requests.post(
            f"{BIN_MANAGER_API}/bins/{bin_name}/progress",
            json=payload,
            headers={"Content-Type": "application/json"},
            timeout=10,
        )

        log(
            f"Reported completion for {bin_name}-{bin_hash}: {status} (took {elapsed}s)"
        )
    except Exception as e:
        error(f"Failed to report completion: {e}")


def kill_old_downloads(bin_name: str):
    try:
        result = subprocess.run(
            ["pgrep", "-f", f"curl.*{DOWNLOAD_BASE_URL}/{bin_name}$"],
            capture_output=True,
            text=True,
        )
        if result.returncode == 0:
            pids = result.stdout.strip().split("\n")
            current_pid = str(os.getpid())
            pids = [pid for pid in pids if pid and pid != current_pid]

            if pids:
                log(f"Killing old download processes for {bin_name}: {', '.join(pids)}")
                for pid in pids:
                    try:
                        subprocess.run(["kill", "-9", pid], check=False)
                    except Exception:
                        pass
    except Exception:
        pass


def post_update_status(bin_name: str, new_sha256: str) -> bool:
    node_id = socket.gethostname()

    payload = {
        "node_id": node_id,
        "sha256sum": new_sha256,
    }

    try:
        response = requests.post(
            f"{BIN_MANAGER_API}/bins/{bin_name}",
            json=payload,
            headers={"Content-Type": "application/json"},
            timeout=10,
        )
        if response.status_code in (200, 201):
            log(f"Posted update status for {bin_name} to API")
            return True
        else:
            error(
                f"Failed to post update status for {bin_name}: {response.status_code}"
            )
            return False
    except Exception as e:
        error(f"Failed to post update status for {bin_name}: {e}")
        return False


def get_sha256sum(file_path: str) -> str:
    if not os.path.exists(file_path):
        return ""

    try:
        sha256_hash = hashlib.sha256()
        with open(file_path, "rb") as f:
            for byte_block in iter(lambda: f.read(4096), b""):
                sha256_hash.update(byte_block)
        return sha256_hash.hexdigest()
    except Exception as e:
        error(f"Error calculating SHA256 for {file_path}: {e}")
        return ""


def query_latest_sha256(bin_name: str) -> Optional[str]:
    if not bin_name:
        error("bin_name is required for query_latest_sha256")
        return None

    url = f"{BIN_MANAGER_API}/bins/{bin_name}"

    try:
        response = requests.get(url, timeout=10)
        if response.status_code == 200:
            data = response.json()
            if "sha256sum" in data:
                return data["sha256sum"]
            elif "sha256" in data:
                return data["sha256"]

        error(f"Failed to query latest SHA256 for {bin_name}")
        return None
    except Exception as e:
        error(f"Failed to query latest SHA256 for {bin_name}: {e}")
        return None


def download_binary(bin_name: str, temp_file: str) -> bool:
    if not bin_name or not temp_file:
        error("bin_name and temp_file are required for download_binary")
        return False

    required_space = 102400
    stat = shutil.disk_usage("/tmp")
    available = stat.free // 1024
    if available < required_space:
        error(
            f"Insufficient disk space in /tmp (available: {available}KB, required: {required_space}KB)"
        )
        return False

    url = f"{DOWNLOAD_BASE_URL}/{bin_name}"

    log(f"Downloading {bin_name} from {url}")

    try:
        response = requests.get(url, timeout=DOWNLOAD_TIMEOUT, stream=True)
        if response.status_code == 200:
            with open(temp_file, "wb") as f:
                for chunk in response.iter_content(chunk_size=8192):
                    f.write(chunk)
            os.chmod(temp_file, 0o755)
            return True
        else:
            error(f"Failed to download {bin_name}: HTTP {response.status_code}")
            return False
    except Exception as e:
        error(f"Failed to download {bin_name}: {e}")
        return False


def change_version(
    bin_name: str, target_sha256: str, operation: str = "upgrade"
) -> bool:
    if not bin_name or not target_sha256:
        error("bin_name and target_sha256 are required for change_version")
        return False

    bin_path = Path(BIN_DIR) / bin_name
    bin_archive_dir = Path(BIN_DIR) / ".archive" / bin_name
    bin_archive_dir.mkdir(parents=True, exist_ok=True)

    log(f"[{operation}] Changing {bin_name} to version (SHA256: {target_sha256})")

    source_file = ""

    if operation == "rollback":
        source_file = str(bin_archive_dir / target_sha256)
        if not os.path.exists(source_file):
            error(
                f"Rollback failed: archived binary not found for SHA256: {target_sha256}"
            )
            return False
        log(f"Using archived binary for rollback: {source_file}")
    else:
        kill_old_downloads(bin_name)

        with tempfile.NamedTemporaryFile(
            mode="wb", delete=False, prefix=f"{bin_name}.tmp.", dir="/tmp"
        ) as tmp:
            source_file = tmp.name

        report_progress(bin_name, target_sha256)

        if not download_binary(bin_name, source_file):
            report_completion(bin_name, target_sha256, "failed")
            return False

        downloaded_sha256 = get_sha256sum(source_file)

        if downloaded_sha256 != target_sha256:
            error(
                f"SHA256 mismatch for downloaded {bin_name} (expected: {target_sha256}, got: {downloaded_sha256})"
            )
            os.unlink(source_file)
            report_completion(bin_name, target_sha256, "failed")
            return False

    current_sha256 = ""
    if bin_path.exists():
        current_sha256 = get_sha256sum(str(bin_path))
        if current_sha256:
            archive_path = bin_archive_dir / current_sha256
            if not archive_path.exists():
                shutil.copy2(str(bin_path), str(archive_path))
                os.chmod(str(archive_path), 0o755)
                log(f"Archived current binary: {archive_path}")

    if operation == "upgrade":
        shutil.copy2(source_file, str(bin_path))
        os.unlink(source_file)
    else:
        shutil.copy2(source_file, str(bin_path))

    os.chmod(str(bin_path), 0o755)

    log("Binary file replaced successfully")

    if shutil.which("supervisorctl"):
        log(f"Restarting service: {bin_name} via supervisor")
        try:
            result = subprocess.run(
                ["supervisorctl", "restart", bin_name],
                capture_output=True,
                text=True,
                timeout=30,
            )
            log(result.stdout)
            if result.stderr:
                log(result.stderr)

            if result.returncode == 0:
                log(f"Service {bin_name} restarted successfully")
                time.sleep(2)

                status_result = subprocess.run(
                    ["supervisorctl", "status", bin_name],
                    capture_output=True,
                    text=True,
                    timeout=10,
                )

                if "RUNNING" in status_result.stdout:
                    log(f"Service {bin_name} verified running after restart")
                    if operation == "upgrade":
                        report_completion(bin_name, target_sha256, "success")
                        post_update_status(bin_name, target_sha256)
                    record_version_change(
                        bin_name, current_sha256, target_sha256, operation, "success"
                    )
                    return True
                else:
                    error(f"Service {bin_name} not running after restart")
                    if current_sha256:
                        log("Auto-rollback: restoring previous version")
                        rollback_source = bin_archive_dir / current_sha256
                        if rollback_source.exists():
                            shutil.copy2(str(rollback_source), str(bin_path))
                            rollback_result = subprocess.run(
                                ["supervisorctl", "restart", bin_name],
                                capture_output=True,
                                text=True,
                                timeout=30,
                            )
                            if rollback_result.returncode == 0:
                                log("Auto-rollback successful, service restarted")
                                record_version_change(
                                    bin_name,
                                    target_sha256,
                                    current_sha256,
                                    "auto-rollback",
                                    "success",
                                )
                            else:
                                error("Auto-rollback failed - service may be down")
                                record_version_change(
                                    bin_name,
                                    target_sha256,
                                    current_sha256,
                                    "auto-rollback",
                                    "failed",
                                )

                    if operation == "upgrade":
                        report_completion(bin_name, target_sha256, "failed")
                    record_version_change(
                        bin_name, current_sha256, target_sha256, operation, "failed"
                    )
                    return False
            else:
                error(f"Failed to restart service {bin_name}")
                if current_sha256:
                    log("Auto-rollback: restoring previous version")
                    rollback_source = bin_archive_dir / current_sha256
                    if rollback_source.exists():
                        shutil.copy2(str(rollback_source), str(bin_path))
                        rollback_result = subprocess.run(
                            ["supervisorctl", "restart", bin_name],
                            capture_output=True,
                            text=True,
                            timeout=30,
                        )
                        if rollback_result.returncode == 0:
                            log("Auto-rollback successful, service restarted")
                            record_version_change(
                                bin_name,
                                target_sha256,
                                current_sha256,
                                "auto-rollback",
                                "success",
                            )
                        else:
                            error("Auto-rollback failed - service may be down")
                            record_version_change(
                                bin_name,
                                target_sha256,
                                current_sha256,
                                "auto-rollback",
                                "failed",
                            )

                if operation == "upgrade":
                    report_completion(bin_name, target_sha256, "failed")
                record_version_change(
                    bin_name, current_sha256, target_sha256, operation, "failed"
                )
                return False
        except subprocess.TimeoutExpired:
            error(f"Timeout restarting service {bin_name}")
            return False
        except Exception as e:
            error(f"Error restarting service {bin_name}: {e}")
            return False
    else:
        log("supervisorctl not found, skipping service restart")
        if operation == "upgrade":
            report_completion(bin_name, target_sha256, "success")
            post_update_status(bin_name, target_sha256)
        record_version_change(
            bin_name, current_sha256, target_sha256, operation, "success"
        )

    return True


def record_version_change(
    bin_name: str, from_sha256: str, to_sha256: str, operation: str, result: str
):
    if not os.path.exists(BIN_MANIFESTS):
        return

    try:
        with open(BIN_MANIFESTS, "r") as f:
            manifests = json.load(f)

        for binary in manifests.get("binaries", []):
            if binary.get("binaryName") == bin_name:
                binary["previousVersion"] = binary.get("version", "")
                binary["version"] = to_sha256
                break

        with open(BIN_MANIFESTS, "w") as f:
            json.dump(manifests, f, indent=2)

        log(
            f"Recorded version change: {bin_name} ({operation}: {from_sha256} -> {to_sha256}, result: {result})"
        )
    except Exception as e:
        error(f"Failed to record version change: {e}")


def update_binary(bin_name: str, current_sha256: str, latest_sha256: str) -> bool:
    if not bin_name:
        error("bin_name is required for update_binary")
        return False

    if current_sha256 == latest_sha256 and current_sha256:
        log(f"{bin_name} is already up to date (SHA256: {current_sha256})")
        return True

    log(f"Updating {bin_name} (current: {current_sha256}, latest: {latest_sha256})")

    return change_version(bin_name, latest_sha256, "upgrade")


def rollback_binary(bin_name: str, target_sha256: str) -> bool:
    if not bin_name or not target_sha256:
        error("bin_name and target_sha256 are required for rollback_binary")
        return False

    log(f"Rolling back {bin_name} to SHA256: {target_sha256}")

    return change_version(bin_name, target_sha256, "rollback")


def update_manifest(bin_name: str, new_sha256: str):
    if not os.path.exists(BIN_MANIFESTS):
        error(f"Manifests file not found: {BIN_MANIFESTS}")
        return

    try:
        with open(BIN_MANIFESTS, "r") as f:
            manifests = json.load(f)

        for binary in manifests.get("binaries", []):
            if binary.get("binaryName") == bin_name:
                binary["version"] = new_sha256
                break

        with open(BIN_MANIFESTS, "w") as f:
            json.dump(manifests, f, indent=2)
    except Exception as e:
        error(f"Failed to update manifest: {e}")


def process_binary(bin_name: str, current_sha256: str) -> bool:
    if not bin_name:
        error("bin_name is required for process_binary")
        return False

    log(f"Processing binary: {bin_name}")

    latest_sha256 = query_latest_sha256(bin_name)

    if not latest_sha256:
        error(f"Failed to get latest SHA256 for {bin_name}")
        return False

    if not acquire_lock(bin_name, latest_sha256):
        return False

    try:
        if update_binary(bin_name, current_sha256, latest_sha256):
            update_manifest(bin_name, latest_sha256)
            release_lock(bin_name, latest_sha256)
            return True
    except Exception as e:
        error(f"Error processing binary {bin_name}: {e}")

    release_lock(bin_name, latest_sha256)
    return False


def update_node_info():
    if not os.path.exists(BIN_MANIFESTS):
        return

    node_info = get_node_info()

    try:
        with open(BIN_MANIFESTS, "r") as f:
            manifests = json.load(f)

        manifests["nodeInfo"] = node_info

        with open(BIN_MANIFESTS, "w") as f:
            json.dump(manifests, f, indent=2)
    except Exception as e:
        error(f"Failed to update node info: {e}")


def main():
    log(f"=== Starting bin-proxy v{BIN_PROXY_VERSION} ===")

    if not os.path.exists(BIN_MANIFESTS):
        error(f"Manifests file not found: {BIN_MANIFESTS}")
        sys.exit(1)

    update_node_info()

    keepalive_check()

    try:
        with open(BIN_MANIFESTS, "r") as f:
            manifests = json.load(f)

        binaries = manifests.get("binaries", [])

        for binary in binaries:
            bin_name = binary.get("binaryName")
            current_sha256 = binary.get("version", "")

            if bin_name:
                try:
                    process_binary(bin_name, current_sha256)
                except Exception as e:
                    error(f"Error processing {bin_name}: {e}")
    except Exception as e:
        error(f"Failed to read manifests: {e}")
        sys.exit(1)

    log("=== bin-proxy completed ===")


if __name__ == "__main__":
    main()
