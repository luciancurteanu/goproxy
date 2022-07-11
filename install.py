#!/usr/bin/env python3
import os
import subprocess as sp
import sys
import argparse


parser = argparse.ArgumentParser(
    description='A hacky script to install the go server'
)

parser.add_argument('--path', required=True, help='path to deploy to')
parser.add_argument('--user', required=True, help='user to deploy as')
parser.add_argument('--name', default='goproxy', help='name of the built binary')
parser.add_argument('--no-service', action='store_true', default=False)

args = parser.parse_args()
args.path = os.path.abspath(args.path)

# Check if we're root
if os.getuid():
    print('this script must be run as root', file=sys.stderr)
    exit(1)

# FS manipulation
root = os.path.abspath(os.path.dirname(__file__))
in_root = lambda *p: os.path.join(root, *p)

if not os.path.exists(os.path.dirname(args.path)):
    print('path', os.path.dirname(args.path), 'does not exist')
    exit(1)

if not os.path.exists(args.path):
    os.mkdir(args.path)

in_path = lambda *p: os.path.join(args.path, *p)

# Build program
if sp.call(['go', 'build', '-o', in_path(args.name), in_root('main.go')]):
    print('error building program', file=sys.stderr)
    exit(1)

# Copy config
if sp.call(['cp', in_root('config.json'), in_path('config.json')]):
    print('error copying config', file=sys.stderr)
    exit(1)

# Set ownership
if sp.call(['chown', '-R', f'{args.user}:{args.user}', args.path]):
    print('error changing ownership for', args.path,
          'to', args.user, file=sys.stderr)
    exit(1)

# Set mode
if sp.call(['chmod', '0750', args.path]):
    print('error changing mode for', args.path, file=sys.stderr)

if args.no_service:
    exit(0)


with open(os.path.join('/etc/systemd/system/', args.name + '.service'), 'w') as f:
    f.write(f'''
[Unit]
Description={args.name} service
After=network.target
StartLimitIntervalSec=0

[Service]
Type=simple
Restart=always
RestartSec=1
User={args.user}
WorkingDirectory={args.path}
ExecStart={in_path(args.name)} -config={in_path('config.json')}

[Install]
WantedBy=multi-user.target
    ''')
