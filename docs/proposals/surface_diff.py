#!/usr/bin/env python3
"""Spec-vs-tree CLI surface diff for go-c8y-cli.

Compares the frozen v1 yaml/json specs (api/spec/json) against the live Cobra
command tree (cmd/gen-surface JSON dump). Classifies deviations:
  - commands missing from the tree (breaking)
  - flags dropped (breaking)
  - flag type-category changes (slice<->scalar etc.)
Additive differences (extra commands/flags) are counted, not listed.

Rules learned during the 2026-07-01 audit:
  - group-level `skip` (unofficial APIs removed in 2023) excludes the whole spec
  - per-command `skip`/`hidden` excluded (tree walk hides hidden commands)
  - params with `children` contribute their children as flags, NOT the parent
    (v1 codegen expanded computed types like firmwareDetails into child flags)
  - `stringStatic` params are fixed query/body fragments, never flags
  - the tree walk excludes hidden/deprecated flags, so v1-hidden flags appear
    "dropped" — they live in ALLOWLIST instead
"""
import json, glob, sys, collections

SURFACE = sys.argv[1] if len(sys.argv) > 1 else 'surface.json'
SPEC_GLOB = 'api/spec/json/*.json'

# Locally installed extensions pollute the surface; exclude their roots.
EXTENSIONS = {'analytics', 'continuous', 'devmgmt', 'mapper', 'organizer', 'tedge'}

# Verified accepted deviations: flags hidden/deprecated in v1 (the tree walk
# excludes hidden+deprecated flags so they cannot round-trip), or deliberate.
ALLOWLIST_FLAGS = {
    ('microservices loglevels delete', 'logLevel'),           # hidden in v1 (hardcoded body template)
    ('remoteaccess configurations create-telnet', 'credentialsType'),  # hidden in v1
    ('ui plugins list', 'hasVersions'),                        # hidden in v1
    ('inventory find', 'onlyDevices'),                         # hidden in v1 (deprecated)
    ('ui plugins versions create', 'tags'),                    # present as deprecated alias of --tag
    ('ui plugins versions update', 'tags'),                    # present as deprecated alias of --tag
}
# Accepted type-category changes (additive: scalar -> slice accepts more).
ALLOWLIST_TYPES = set()

IGNORE = {'data', 'template', 'templateVars', 'processingMode'}

def spec_cat(t):
    if t in ('boolean', 'booleanDefault', 'optional_fragment'):
        return 'bool'
    if t == 'integer':
        return 'int'
    if t == 'float':
        return 'float'
    if t.endswith('[]'):
        return 'slice'
    if t == 'stringStatic':
        return 'static'
    if t in ('firmwareDetails', 'softwareDetails', 'configurationDetails'):
        # computed types: the v1 codegen expanded these into sibling flags
        # (name/version/url); the param itself never became a flag
        return 'static'
    return 'string'

def tree_cat(t):
    if t == 'bool':
        return 'bool'
    if t in ('int', 'int8', 'int16', 'int32', 'int64', 'count'):
        return 'int'
    if t in ('float', 'float32', 'float64'):
        return 'float'
    if 'Slice' in t or 'Array' in t or t == 'stringToString':
        return 'slice'
    return 'string'

def add_param(flags, p):
    children = p.get('children') or []
    if children:
        # computed type: the children become flags, the parent does not
        for ch in children:
            add_param(flags, ch)
        return
    cat = spec_cat(p.get('type', 'string'))
    if cat == 'static':
        return
    nm = p.get('name')
    if not nm or nm in IGNORE:
        return
    flags[nm] = cat

spec_cmds = {}
for f in sorted(glob.glob(SPEC_GLOB)):
    d = json.load(open(f))
    if d.get('group', {}).get('skip'):
        continue
    gname = d['group']['name'].lower().replace('/', ' ')
    for c in d.get('commands', []):
        if c.get('skip') or c.get('hidden'):
            continue
        alias = (c.get('alias') or {}).get('go') or c['name']
        key = f"{gname} {alias}"
        flags = {}
        pipe = None
        for sect in ('queryParameters', 'pathParameters', 'headerParameters', 'body'):
            for p in c.get(sect) or []:
                add_param(flags, p)
                if p.get('pipeline'):
                    pipe = p.get('name')
        spec_cmds[key] = dict(flags=flags, pipe=pipe)

tree_cmds = {}
for c in json.load(open(SURFACE)):
    if c['path'][0] in EXTENSIONS:
        continue
    key = ' '.join(c['path'])
    flags = {}
    pipe = None
    for fl in c.get('flags') or []:
        if fl['name'] in IGNORE:
            continue
        flags[fl['name']] = tree_cat(fl.get('type', 'string'))
        if fl.get('pipeline'):
            pipe = fl['name']
    tree_cmds[key] = dict(flags=flags, pipe=pipe)

missing_cmds, flag_missing, type_changed, pipe_changed = [], [], [], []
extra_flag_count = matched = 0
for key, sc in sorted(spec_cmds.items()):
    tc = tree_cmds.get(key)
    if tc is None:
        missing_cmds.append(key)
        continue
    matched += 1
    for fname, fcat in sc['flags'].items():
        if (key, fname) in ALLOWLIST_FLAGS:
            continue
        if fname not in tc['flags']:
            flag_missing.append((key, fname, fcat))
        elif tc['flags'][fname] != fcat:
            if (key, fname) in ALLOWLIST_TYPES:
                continue
            # scalar->slice is additive (accepts more input); slice->scalar breaks
            if fcat == 'string' and tc['flags'][fname] == 'slice':
                continue
            type_changed.append((key, fname, fcat, tc['flags'][fname]))
    extra_flag_count += len(set(tc['flags']) - set(sc['flags']))
    if sc['pipe'] and tc['pipe'] and sc['pipe'] != tc['pipe']:
        pipe_changed.append((key, sc['pipe'], tc['pipe']))

extra_cmds = sorted(set(tree_cmds) - set(spec_cmds))

print(f"spec commands: {len(spec_cmds)}  core tree commands: {len(tree_cmds)}  matched: {matched}")
print(f"\n== SPEC COMMANDS MISSING FROM TREE: {len(missing_cmds)} ==")
for k in missing_cmds:
    print(' ', k)
print(f"\n== FLAGS DROPPED: {len(flag_missing)} ==")
for k, fl, c in flag_missing:
    print(f"  {k}: --{fl} ({c})")
print(f"\n== BREAKING TYPE CHANGES: {len(type_changed)} ==")
for k, fl, a, b in type_changed:
    print(f"  {k}: --{fl} {a} -> {b}")
print(f"\n== PIPELINE DRIVER CHANGED: {len(pipe_changed)} ==")
for k, a, b in pipe_changed:
    print(f"  {k}: {a} -> {b}")
print(f"\n== additive: {len(extra_cmds)} extra commands, {extra_flag_count} extra flags ==")
sys.exit(1 if (missing_cmds or flag_missing or type_changed or pipe_changed) else 0)
