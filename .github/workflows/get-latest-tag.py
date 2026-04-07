#!/usr/bin/env python3
import json, sys, os

token = os.environ.get("GH_TOKEN", "")
headers = {}
if token:
    headers["Authorization"] = f"Bearer {token}"

# Fetch releases from sing-box_mod
import urllib.request
url = "https://api.github.com/repos/ouyangmland/sing-box_mod/releases?per_page=50"
req = urllib.request.Request(url, headers=headers)
with urllib.request.urlopen(req) as resp:
    releases = json.loads(resp.read())

# Find latest v2bx tag
tags = sorted([
    r["tag_name"] for r in releases
    if "v2bx" in r["tag_name"] and r["tag_name"].startswith("v")
])
print(tags[-1] if tags else "")
