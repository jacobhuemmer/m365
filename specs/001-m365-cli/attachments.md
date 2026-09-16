# m365 Attachments

This file owns outbound attach, inbound metadata, and save-to-path behavior
for both mail and Teams. Attachments are not a third workload namespace.
Mail parent commands are owned by [mail.md](mail.md). Teams parent commands
are owned by [teams.md](teams.md). CLI streams and exit classes are owned by
[cli-contract.md](cli-contract.md).

## User Scenarios & Testing *(mandatory)*

### User Story 6 — Attach local documents on send or reply (Priority: P2)

The user attaches one or more local documents when sending or replying (mail
and Teams). Dry-run lists the files that would be attached (name, size) and
does not upload. Missing, unreadable, empty, oversize, and too-many files
are validation failures, not service errors.

**Why this priority**: Documents are part of real mail and chat, but attach
depends on send/reply already existing.

**Independent Test**: Dry-run send with two synthetic local files; assert
names and sizes appear and no upload occurs. Repeat without dry-run and
assert the sent message carries those attachments. Missing path, empty file,
file over 10 MiB, and an 11th file each exit `3` with no send.

**Acceptance Scenarios**:

1. **Given** one or more existing local files within size and count caps,
   **When** the user dry-runs mail send, mail reply, or Teams send with those
   paths, **Then** stdout lists each file's name and size, exits `0`, and
   does not upload.
2. **Given** the same inputs without `--dry-run`, **When** the user sends,
   **Then** the created message includes those attachments.
3. **Given** a missing path or unreadable file, **When** the user sends with
   that attach flag, **Then** the command exits `3` and does not send.
4. **Given** a zero-byte file, a file larger than 10 MiB, or more than 10
   files, **When** the user sends, **Then** the command exits `3` before
   send, names the cap that was exceeded, and does not silently skip a file.
5. **Given** a remote URL instead of a local path, **When** the user attaches
   it, **Then** the command exits `3` and does not fetch the URL.

### User Story 7 — Save a named attachment to a local path (Priority: P2)

The user names the message or chat message, the attachment, and a destination
path. Success reports the path written. Omitting the path is a usage error.
An existing local file is refused unless overwrite is explicit.

**Why this priority**: Saving a document is the safe counterpart of attach,
and it must not surprise-write a file.

**Independent Test**: Save a synthetic attachment to a new path; assert the
file exists and stdout reports that path. Omit the path (exit `3`, no write).
Save to an existing path without overwrite (exit `3`, file unchanged). Save
with overwrite (file replaced). Unknown attachment exits `6`.

**Acceptance Scenarios**:

1. **Given** a known mail or Teams attachment and a new destination path,
   **When** the user saves, **Then** the file is written only there, stdout
   reports that path, and file bytes do not appear on stdout.
2. **Given** no destination path, **When** the user saves, **Then** the
   command exits `3` and writes no file.
3. **Given** a destination that already exists and no overwrite flag,
   **When** the user saves, **Then** the command exits `3` and the existing
   file is unchanged.
4. **Given** `--overwrite` and an existing destination, **When** the user
   saves, **Then** the destination is replaced and stdout reports that path.
5. **Given** an unknown message or unknown attachment, **When** the user
   saves, **Then** the command exits `6` and writes no file.
6. **Given** two inbound attachments with the same display name, **When** the
   user saves by name, **Then** the command exits `3` and asks for the
   attachment id. Saving by id succeeds.

## Functional Requirements

- **FR-048**: Mail send, mail reply, and Teams send MUST accept one or more
  local filesystem paths as attachments. Attachments belong to that send;
  they are not a separate namespace.
- **FR-049**: `--dry-run` with attach flags MUST list each file's display
  name and size and MUST NOT upload file bytes.
- **FR-050**: Missing path, unreadable file, zero-byte file, remote URL, and
  non-file path MUST fail with exit `3` before send. These MUST NOT be
  reported as Microsoft 365 service errors.
- **FR-051**: Each outbound file MUST be at most 10 MiB. Each send or reply
  MUST attach at most 10 files. Exceeding either cap MUST fail with exit `3`
  before send and MUST name the cap. `--help` for send, reply, and attach
  MUST show both caps.
- **FR-052**: The tool MUST NOT silently skip, truncate, or drop an
  attachment. Two outbound files with the same base name MAY both be
  attached; they remain distinct paths.
- **FR-053**: List, get, thread, and messages MUST include attachment
  metadata: id, display name, size, and content type. Zero attachments MUST
  be an empty list with exit `0`, not an error.
- **FR-054**: Save MUST require the parent message identity (mail id, or
  Teams chat id plus message id), the attachment id or unique display name,
  and a destination path. Success MUST write the file only to that path and
  MUST report that path on stdout. Save MUST NOT print file bytes on stdout
  or stderr.
- **FR-055**: If the destination exists, save MUST refuse with exit `3`
  unless `--overwrite` is present. `--overwrite` MUST replace the file.
- **FR-056**: Save MUST use these failure classes: unknown parent or
  attachment → `6`; missing path, unwritable path, existing file without
  overwrite, or ambiguous display name → `3`; usable session rejected by
  Microsoft 365 → `5`; missing session or missing namespace consent → `4`.
- **FR-057**: Attachment bytes MUST NOT be written to stdout in list, get,
  thread, messages, watch, send, or dry-run. The tool MUST NOT browse
  OneDrive or SharePoint as a file picker. The tool MUST NOT attach from a
  remote URL. Local paths only.
- **FR-058**: Attachment operations MUST stay in the parent namespace. Mail
  save and mail attachments-list MUST NOT call Teams. Teams save and Teams
  attachments-list MUST NOT call mail. A parent identity from the other
  namespace MUST NOT be treated as a successful cross-namespace save.
- **FR-059**: Tests and fixtures MUST use synthetic files and metadata. Live
  mailbox or chat documents MUST NOT be stored in the repository.
