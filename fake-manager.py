#!/usr/bin/env python3

from flask import Flask, request, jsonify
import hashlib
import random

app = Flask(__name__)

nodes = {}
bin_hashes = {}

@app.route('/api/v1/keepalive', methods=['GET', 'POST'])
def keepalive():
    if request.method == 'GET':
        node_id = request.args.get('node_id')
        if not node_id:
            return jsonify({'error': 'node_id parameter required'}), 400
        
        if node_id in nodes:
            return jsonify(nodes[node_id]), 200
        else:
            return jsonify({'error': 'node not found'}), 404
    
    elif request.method == 'POST':
        data = request.get_json()
        if not data or 'node_id' not in data:
            return jsonify({'error': 'node_id required in request body'}), 400
        
        node_id = data['node_id']
        nodes[node_id] = {
            'node_id': node_id,
            'cpu_arch': data.get('cpu_arch', 'x86_64'),
            'os_release': data.get('os_release', 'unknown'),
            'node_name': data.get('node_name', node_id),
            'bin_proxy_version': data.get('bin_proxy_version', '1.0.0')
        }
        
        return jsonify({
            'message': 'node registered successfully',
            'node': nodes[node_id]
        }), 201

@app.route('/api/v1/bins/<bin_name>', methods=['GET', 'POST'])
def bins(bin_name):
    if request.method == 'GET':
        if bin_name not in bin_hashes:
            bin_hashes[bin_name] = {
                'bin_name': bin_name,
                'sha256sum': hashlib.sha256(f'{bin_name}-v{random.randint(1, 100)}'.encode()).hexdigest(),
                'version': 'latest'
            }
        
        return jsonify(bin_hashes[bin_name]), 200
    
    elif request.method == 'POST':
        data = request.get_json()
        if not data or 'sha256sum' not in data:
            return jsonify({'error': 'sha256sum required in request body'}), 400
        
        node_id = data.get('node_id')
        sha256sum = data['sha256sum']
        
        if bin_name not in bin_hashes:
            bin_hashes[bin_name] = {
                'bin_name': bin_name,
                'sha256sum': sha256sum,
                'version': 'latest',
                'nodes': {}
            }
        
        if 'nodes' not in bin_hashes[bin_name]:
            bin_hashes[bin_name]['nodes'] = {}
        
        if node_id:
            bin_hashes[bin_name]['nodes'][node_id] = {
                'sha256sum': sha256sum,
                'updated_at': 'now'
            }
        
        return jsonify({
            'message': f'bin {bin_name} hash updated successfully',
            'bin_name': bin_name,
            'sha256sum': sha256sum
        }), 200

if __name__ == '__main__':
    app.run(host='0.0.0.0', port=8080, debug=True)
