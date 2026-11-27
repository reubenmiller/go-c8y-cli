"""Runs a command, sends it SIGINT after a delay and prints how long it took to end.

usage: interrupt.py <delay seconds> <command> [args...]
Prints "<seconds> <exit code>"; the exit code is negative for a command killed by a signal.
"""
import signal
import subprocess
import sys
import time

delay = float(sys.argv[1])
start = time.time()
proc = subprocess.Popen(sys.argv[2:], stdin=subprocess.DEVNULL, stdout=subprocess.DEVNULL, stderr=subprocess.DEVNULL)
time.sleep(delay)
proc.send_signal(signal.SIGINT)
try:
    proc.wait(timeout=60)
except subprocess.TimeoutExpired:
    proc.kill()
    proc.wait()
print("%.1f %d" % (time.time() - start, proc.returncode))
