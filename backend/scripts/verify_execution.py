#!/usr/bin/env python3
"""Exercise a real API process against a disposable, fully migrated test database."""
import argparse
import json
import os
from pathlib import Path
import statistics
import subprocess
import tempfile
import time
import urllib.error
import urllib.request
import uuid

parser = argparse.ArgumentParser(description=__doc__)
parser.add_argument('--binary', required=True, help='Path to a built Automata server')
parser.add_argument('--port', type=int, default=18080)
parser.add_argument('--samples', type=int, default=20, choices=range(1, 41), metavar='1..40')
parser.add_argument('--output', type=Path)
args = parser.parse_args()
database = os.environ.get('TEST_DATABASE_URL')
if not database:
    parser.error('TEST_DATABASE_URL must point to a disposable, migrated database')
binary = str(Path(args.binary).resolve())
base = f'http://127.0.0.1:{args.port}'
token = None
process = None


def request(path, data=None):
    headers = {'Content-Type': 'application/json'}
    if token:
        headers['Authorization'] = f'Bearer {token}'
    req = urllib.request.Request(base + path, data=None if data is None else json.dumps(data).encode(), headers=headers)
    with urllib.request.urlopen(req, timeout=10) as response:
        return json.load(response)


def wait_for(check, timeout=90):
    deadline = time.monotonic() + timeout
    while time.monotonic() < deadline:
        value = check()
        if value:
            return value
        time.sleep(0.1)
    raise TimeoutError('Timed out waiting for the expected execution state')


def run_finished(run_id):
    run = request('/runs/' + run_id)
    return run if run['status'] in ('succeeded', 'failed') else None


with tempfile.TemporaryDirectory(prefix='automata-verification-') as directory:
    env = dict(os.environ, DATABASE_URL=database, HTTP_PORT=str(args.port), ENV='test',
               JWT_SECRET='verification-only-secret-at-least-32-characters',
               CLERK_SECRET_KEY='', TEST_DISABLE_RATE_LIMIT='1')
    log_path = Path(directory) / 'server.log'

    def start():
        global process
        # Refuse to use an unrelated server already listening on the selected port.
        import socket
        with socket.socket() as probe:
            if probe.connect_ex(('127.0.0.1', args.port)) == 0:
                raise RuntimeError(f'Port {args.port} is already occupied')
        with log_path.open('ab') as log:
            process = subprocess.Popen([binary], cwd=directory, env=env, stdout=log, stderr=log)
        def ready():
            if process.poll() is not None:
                raise RuntimeError('API startup failed; check database migrations and configuration')
            try:
                return request('/healthz')
            except (urllib.error.URLError, ConnectionError):
                return None
        wait_for(ready, 15)

    try:
        start()
        identifier = uuid.uuid4().hex
        account = request('/auth/register', {'username': 'verify_' + identifier[:8],
                          'email': identifier + '@example.com', 'password': 'Verify-' + identifier + '!'})
        token = account['access_token']
        graph = {
            'nodes': [
                {'id': 'start', 'type': 'manual_trigger', 'config': {}},
                {'id': 'pause', 'type': 'delay', 'config': {'seconds': 5}, 'retry': {'max_attempts': 2, 'backoff_seconds': 0}},
                {'id': 'finish', 'type': 'logger', 'config': {'message': 'Recovery completed'}},
            ],
            'edges': [{'from_node_id': 'start', 'to_node_id': 'pause'}, {'from_node_id': 'pause', 'to_node_id': 'finish'}],
        }
        workflow = request('/workflows', {'name': 'Worker restart verification', 'graph': graph})['Workflow']
        run_id = request('/workflows/' + workflow['id'] + '/run', {})['id']
        def interrupted_state():
            run = request('/runs/' + run_id)
            steps = {step['node_id']: step for step in run.get('nodes', [])}
            return steps if steps.get('start', {}).get('status') == 'succeeded' and steps.get('pause', {}).get('status') == 'running' else None
        before = wait_for(interrupted_state)
        process.kill()  # Only the child process created by this script.
        process.wait(timeout=10)
        restarted = time.monotonic()
        start()
        recovered = wait_for(lambda: run_finished(run_id))
        recovery_seconds = time.monotonic() - restarted
        steps = {step['node_id']: step for step in recovered['nodes']}
        assert recovered['status'] == 'succeeded', recovered
        assert steps['start']['attempts'] == 1, steps
        assert steps['start']['output'] == before['start']['output'], steps
        assert steps['pause']['attempts'] == 2, steps
        assert steps['finish']['attempts'] == 1, steps

        simple = request('/workflows', {'name': 'Single-node latency baseline', 'graph': {
            'nodes': [{'id': 'start', 'type': 'manual_trigger', 'config': {}}], 'edges': []}})['Workflow']
        admissions, completions = [], []
        for _ in range(args.samples):
            started = time.monotonic()
            run = request('/workflows/' + simple['id'] + '/run', {})
            admissions.append((time.monotonic() - started) * 1000)
            completed = wait_for(lambda: run_finished(run['id']))
            assert completed['status'] == 'succeeded', completed
            completions.append((time.monotonic() - started) * 1000)
        def summary(values):
            ordered = sorted(values)
            import math
            return {'median_ms': round(statistics.median(values), 2), 'p95_ms': round(ordered[math.ceil(len(ordered) * .95) - 1], 2)}
        result = {'samples': args.samples, 'concurrency': 1, 'poll_interval_ms': 100,
                  'recovery': {'status': recovered['status'], 'seconds_after_restart': round(recovery_seconds, 2),
                               'node_attempts': {key: value['attempts'] for key, value in steps.items()}},
                  'admission': summary(admissions), 'observed_completion': summary(completions),
                  'scope': 'Local sequential single-node baseline, not a throughput or external-effect exactly-once guarantee.'}
        if args.output:
            args.output.write_text(json.dumps(result, indent=2) + '\n')
        print(json.dumps(result, indent=2))
    finally:
        if process and process.poll() is None:
            process.terminate()
            try:
                process.wait(timeout=15)
            except subprocess.TimeoutExpired:
                process.kill()
                process.wait(timeout=5)
