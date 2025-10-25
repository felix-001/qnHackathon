#!/usr/bin/python3
from flask import Flask, request, jsonify
from datetime import datetime, UTC

app = Flask(__name__)

nodes = {}
bins = {
    "bin1": {
        "sha256sum": "abcd1234567890abcd1234567890abcd1234567890abcd1234567890abcd1234",
        "version": "latest",
    },
    "bin2": {
        "sha256sum": "1234abcd5678901234abcd5678901234abcd5678901234abcd5678901234abcd",
        "version": "1.0.0",
    },
    "bin3": {
        "sha256sum": "1234abcd5678901234abcd5678901234abcd5678901234abcd5678901234abc5",
        "version": "latest",
    },
}
node_bins = {}
progress_data = {}


@app.route("/api/v1/keepalive", methods=["GET", "POST"])
def keepalive():
    if request.method == "GET":
        node_id = request.args.get("node_id")
        if not node_id:
            return jsonify({"error": "node_id parameter required"}), 400

        if node_id in nodes:
            return jsonify(nodes[node_id]), 200
        else:
            return jsonify({"error": "node not found"}), 404

    elif request.method == "POST":
        data = request.get_json()
        if not data or "node_id" not in data:
            return jsonify({"error": "node_id required in request body"}), 400

        node_id = data["node_id"]
        nodes[node_id] = {
            "node_id": node_id,
            "cpu_arch": data.get("cpu_arch", ""),
            "os_release": data.get("os_release", ""),
            "node_name": data.get("node_name", ""),
            "bin_proxy_version": data.get("bin_proxy_version", ""),
            "last_seen": datetime.now(UTC).isoformat(),
        }

        return jsonify({"message": "node registered", "node": nodes[node_id]}), 201


@app.route("/api/v1/bins/<bin_name>", methods=["GET", "POST"])
def bins_handler(bin_name):
    if request.method == "GET":
        if bin_name in bins:
            return (
                jsonify(
                    {
                        "bin_name": bin_name,
                        "sha256sum": bins[bin_name]["sha256sum"],
                        "version": bins[bin_name]["version"],
                    }
                ),
                200,
            )
        else:
            return jsonify({"error": "binary not found"}), 404

    elif request.method == "POST":
        data = request.get_json()
        if not data or "sha256sum" not in data or "node_id" not in data:
            return jsonify({"error": "sha256sum and node_id required"}), 400

        node_id = data["node_id"]
        sha256sum = data["sha256sum"]

        if node_id not in node_bins:
            node_bins[node_id] = {}

        node_bins[node_id][bin_name] = {
            "sha256sum": sha256sum,
            "updated_at": datetime.now(UTC).isoformat(),
        }

        return (
            jsonify(
                {
                    "message": "binary version updated for node",
                    "node_id": node_id,
                    "bin_name": bin_name,
                    "sha256sum": sha256sum,
                }
            ),
            200,
        )


@app.route("/api/v1/bins/<bin_name>/progress", methods=["POST"])
def bin_progress(bin_name):
    data = request.get_json()
    if not data:
        return jsonify({"error": "request body required"}), 400

    node_name = data.get("nodeName")
    target_hash = data.get("targetHash")
    processing_time = data.get("processingTime")
    status = data.get("status")

    if not all([node_name, target_hash, status]):
        return (
            jsonify({"error": "nodeName, targetHash, and status are required"}),
            400,
        )

    progress_key = f"{node_name}:{bin_name}:{target_hash}"
    progress_data[progress_key] = {
        "nodeName": node_name,
        "binName": bin_name,
        "targetHash": target_hash,
        "processingTime": processing_time,
        "status": status,
        "updated_at": datetime.utcnow().isoformat(),
    }

    return (
        jsonify(
            {
                "message": "progress recorded",
                "nodeName": node_name,
                "binName": bin_name,
                "status": status,
            }
        ),
        200,
    )


@app.route("/api/v1/download/<bin_file_name>", methods=["GET"])
def download(bin_file_name):
    return (
        jsonify(
            {
                "message": "download endpoint - implement actual file serving as needed",
                "bin_file_name": bin_file_name,
                "note": "This is a mock endpoint. In production, serve actual binary files here.",
            }
        ),
        200,
    )


@app.route("/health", methods=["GET"])
def health():
    return (
        jsonify(
            {"status": "healthy", "nodes_count": len(nodes), "bins_count": len(bins)}
        ),
        200,
    )


if __name__ == "__main__":
    print("Starting fake-manager.py mock server...")
    print("Server running on http://0.0.0.0:8080")
    print("\nAvailable endpoints:")
    print("  GET  /api/v1/keepalive?node_id=<id>")
    print("  POST /api/v1/keepalive")
    print("  GET  /api/v1/bins/<bin-name>")
    print("  POST /api/v1/bins/<bin-name>")
    print("  POST /api/v1/bins/<bin-name>/progress")
    print("  GET  /api/v1/download/<bin-file-name>")
    print("  GET  /health")
    print("\nNote: Requires Flask (pip install flask)")
    app.run(host="0.0.0.0", port=8080, debug=True)
