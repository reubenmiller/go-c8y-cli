#!/usr/bin/env python3
"""Regenerate CLI_MIGRATION_TRACKER.md from the live repo state.

Walks go-c8y-cli/pkg/cmd/*/ for commands still on spec-generated *.auto.go and
cross-references the go-c8y v2 SDK service method surfaces. Run from the go-c8y-cli
repo root:  python3 docs/proposals/gen_tracker.py
"""
import os, re
from collections import Counter

CLI = "pkg/cmd"
SDK = "../go-c8y-v2/pkg/c8y/api"

INFRA = {"activitylog","alias","api","assert","cache","cli","completion","extension","factory",
         "realtime","root","scripts","sessions","settings","subcommand","template","util","version"}

# CLI group -> backing SDK service dir (None = no service exists yet)
GROUP_SVC = {
 "alarms":"alarms","operations":"operations","events":"events","measurements":"measurements",
 "tenants":"tenants","devices":"devices","inventory":"inventory/managedobjects","bulkoperations":"bulkoperations",
 "applications":"applications","microservices":"microservices","users":"users","usergroups":"usergroups",
 "userroles":"userroles","features":"features","notification2":"notification2","remoteaccess":"remoteaccess",
 "binaries":"binaries","devicegroups":"devicegroups",
 "agents":"inventory/managedobjects","deviceprofiles":"inventory/managedobjects","smartgroups":"inventory/managedobjects",
 "firmware":"repository/firmware","software":"repository/software","configuration":"repository/configuration",
 "currenttenant":"tenants/currenttenant","systemoptions":"tenants/systemoptions",
 "tenantoptions":"tenants/tenantoptions","tenantstatistics":"tenants/usagestatistics",
 "devicemanagement":"trustedcertificates","ui":"plugins","userreferences":"users",
 "currentapplication":None,"currentuser":None,"databroker":None,"datahub":None,"deviceregistration":None,
}
CRUD = {"Create","Delete","Get","List","ListAll","Update","ResolveID"}
# services whose method surface isn't a plain `func (s *Service) X(` (composition); assume CRUD
ASSUME_CRUD = {"repository/firmware","repository/software"}
VERIFY = {"plugins","remoteaccess"}  # exists but surface not auto-detected

CAND = {"list":"List","get":"Get","create":"Create","update":"Update","delete":"Delete","count":"Count",
 "deletecollection":"DeleteList","updatecollection":"UpdateList","getseries":"ListSeries",
 "getsupportedmeasurements":"ListSupportedMeasurements","getsupportedseries":"ListSupportedSeries",
 "availability":"GetAvailability","register":"Create","approve":"Update","find":"Find","findbytext":"List",
 "logout":"Logout","resetuserpassword":"ResetPassword","getuserbyname":"GetByUsername",
 "getforcategory":"ListByCategory","updatebulk":"UpdateByCategory","updateedit":"UpdateEditableFlag",
 "getcredentials":"Create","cancel":"Update","getchild":"Get","listchildren":"List","listassets":"List",
 "tfa":"GetTFA","enable":"Enable","disable":"Disable","listreferences":"ListApplicationReferences",
 "subscriptions":"List","listsubscriptions":"List","getinventoryrole":"Get","listinventoryroles":"List",
 "listusermembership":"ListGroupsWithUser","revoketotpsecret":"Update"}

def svcname(s):
    if s is None: return None
    return {"inventory/managedobjects":"ManagedObjects","tenants/currenttenant":"Tenants.Current",
            "tenants/systemoptions":"Tenants.SystemOptions","tenants/tenantoptions":"Tenants.Options",
            "tenants/usagestatistics":"Tenants.UsageStats","repository/firmware":"Repository.Firmware",
            "repository/software":"Repository.Software","repository/configuration":"Repository.Configuration",
            "trustedcertificates":"TrustedCertificates","plugins":"UIPlugins"}.get(s, s.split("/")[-1].capitalize())

def svc_methods(svc):
    if svc in ASSUME_CRUD: return set(CRUD)
    d = os.path.join(SDK, svc); methods=set()
    if os.path.isdir(d):
        for f in os.listdir(d):
            if f.endswith(".go") and not f.endswith("_test.go"):
                methods.update(re.findall(r'func \(s \*Service\) ([A-Z][A-Za-z0-9]*)\(', open(os.path.join(d,f)).read()))
    return methods

method_re = re.compile(r'Method:\s+"([A-Z]+)"'); path_re = re.compile(r'NewStringTemplate\("([^"]+)"\)')
BIN = lambda s: any(k in s for k in ("binary","registerbulk","download","upload"))
SUBRES = lambda s: any(k in s for k in ("child","asset","addition","version","subscription","token","reference",
        "membership","statistic","services","plugins","configurations","certificates","patches","jobs","loglevel","tenants"))

rows=[]
for g in sorted(os.listdir(CLI)):
    gdir=os.path.join(CLI,g)
    if not os.path.isdir(gdir) or g in INFRA: continue
    svc=GROUP_SVC.get(g,"??"); sname=svcname(svc); methods=svc_methods(svc) if svc else set()
    for sub in sorted(os.listdir(gdir)):
        sdir=os.path.join(gdir,sub)
        if not os.path.isdir(sdir): continue
        autos=[f for f in os.listdir(sdir) if f.endswith(".auto.go")]
        if not autos: continue  # migrated
        txt=open(os.path.join(sdir,autos[0])).read()
        m=method_re.search(txt); p=[x for x in path_re.findall(txt) if x]
        method=m.group(1) if m else "—"; path=p[0] if p else "— (dynamic)"
        cand=CAND.get(sub)
        if svc is None: sdk="❌ none"; eff="L" if g=="deviceregistration" else "M"
        elif svc in VERIFY: sdk=f"⚠️ {sname}?"; eff="M"
        elif cand and cand in methods: sdk=f"✅ {sname}.{cand}"; eff="S"
        else: sdk=f"⚠️ {sname} +method"; eff="M"
        if g=="binaries": eff="M"
        if BIN(sub): eff="L"
        if SUBRES(sub) and not sdk.startswith("✅"): eff={"S":"M","M":"M","L":"L"}[eff]
        rows.append(dict(group=g,cmd=sub,method=method,path=path,sdk=sdk,eff=eff))

eo={"S":0,"M":1,"L":2}
rows.sort(key=lambda r:(eo[r['eff']], r['sdk'][0]!="✅", r['group'], r['cmd']))
ce=Counter(r['eff'] for r in rows); cs=Counter(r['sdk'][0] for r in rows)
L=["# CLI → v2 migration tracker (per-command)\n",
   "One row per individual go-c8y-cli command still on a spec-generated `*.auto.go`, cross-referenced against the go-c8y **v2** SDK. Fill the **Pri** (priority) column to plan; sorted by effort then SDK-readiness but meant to be re-sorted/filtered freely. Regenerate with `python3 docs/proposals/gen_tracker.py`.\n",
   f"**Remaining commands: {len(rows)}.**  Effort — S:{ce['S']} · M:{ce['M']} · L:{ce['L']}.  SDK call — ✅ ready:{cs['✅']} · ⚠️ service-exists-method-missing:{cs['⚠']} · ❌ no-service:{cs['❌']}.\n",
   "**Effort:** `S` ≈ ≤½ day (SDK method ready, mechanical) · `M` ≈ ~1 day (new SDK method / resolution / query-build / sub-resource) · `L` ≈ multi-day (new SDK service, or binary/multipart).\n",
   "**SDK call:** ✅ `Service.Method` exists · ⚠️ service exists but this method must be added (`Service?` = surface unverified) · ❌ no service at all.\n",
   "| Pri | Command | HTTP | Endpoint | Effort | SDK call |","|:--:|---|:--:|---|:--:|---|"]
for r in rows:
    L.append(f"|  | `c8y {r['group']} {r['cmd']}` | {r['method']} | `{r['path']}` | {r['eff']} | {r['sdk']} |")
open("docs/proposals/CLI_MIGRATION_TRACKER.md","w").write("\n".join(L)+"\n")
print(f"wrote tracker: {len(rows)} commands; effort {dict(ce)}; sdk {dict(cs)}")
