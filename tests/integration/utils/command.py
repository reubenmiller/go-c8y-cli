"""Global fixtures"""
from dataclasses import dataclass
import json
import os
import subprocess
import time
from typing import Any, Dict, List
import logging

LOGGER = logging.getLogger()


@dataclass
class CommandContext:
    command: str = ""
    duration: float = 0
    exit_code: int = 0
    stdout: str = ""
    stderr: str = ""
    json: Dict[str, Any] = None
    jsonlines: List[Dict[str, Any]] = None

def prepare(command: str, **kwargs) -> str:
    for key, value in kwargs.items():
        command = command.replace(key, str(value))
        LOGGER.warning(f"Replacing {key} with {value}")
    
    return command.strip()


def execute(command: str):
    start = time.monotonic()
    proc = subprocess.Popen(
        ["bash", "-c", command],
        stdout=subprocess.PIPE,
        stderr=subprocess.PIPE,
        env=os.environ,
    )
    proc.wait()
    jsonlines = []
    for line in proc.stdout:
        try:
            jsonlines.append(json.loads(line))
        except:
            pass
    
    return CommandContext(command, exit_code=proc.returncode, stdout=proc.stdout, stderr=proc.stdout, duration=(time.monotonic() - start), json=None, jsonlines=jsonlines)
