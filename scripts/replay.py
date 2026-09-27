#!/usr/bin/env python3
"""Replay one original message using only its consumer context's credential."""
import argparse
import os
from pathlib import Path
import subprocess
import sys
from dev import ROOT, OWNERS, load

parser = argparse.ArgumentParser(description=__doc__)
parser.add_argument("consumer")
parser.add_argument("--env-file", type=Path, default=ROOT/".local/dev.env")
args = parser.parse_args()
owner = args.consumer.split(".", 1)[0]
if owner not in OWNERS:
    parser.error("consumer must belong to one of the six contexts")
values = load(args.env_file)
broker = f'amqp://cafe_{owner}:{values[owner.upper()+"_BROKER_PASSWORD"]}@127.0.0.1:{values["AMQP_PORT"]}/reference'
environment = dict(os.environ, APP_ENV="development", BROKER_URL=broker)
subprocess.run([sys.executable, "scripts/go.py", "run", "./apps/replay", args.consumer],
               cwd=ROOT, env=environment, check=True)
