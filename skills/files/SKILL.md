---
name: files
description: Browse OneDrive, get item metadata, download to a named path, and dry-run upload.
---

# files

  files list
  files get ITEM_ID
  files download ITEM_ID --out /path/to/dest
  files upload --file /path/to/local --dry-run

Download writes only to --out. File bytes MUST NOT appear in MCP results.
