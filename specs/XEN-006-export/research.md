# Research: Export (Stage 6)

## R1. Postman export = inverse mapping
- **Decision**: one mapping table (XEN-003 postman-mapping.md) drives both directions; the Go
  writer DTOs in `webservice/postman/dto` gain `New{X}(entity, passthrough)` constructors next to
  the existing `ToEntity`. Passthrough merge: Xenon-generated JSON first, then passthrough keys
  not already present; Xenon-owned arrays never merged.
- **Validation**: generated JSON validated against the official v2.1.0 JSON schema (vendored in
  test resources) in tests.

## R2. Encrypted secrets
- **Decision**: age scrypt recipient (`filippo.io/age`, already a dependency since XEN-002),
  armored output, logN 18 (BDD lowers via `XENON_SCRYPT_LOG_N`). Self-contained: no keychain or
  workspace vault key involved, so any Xenon on any machine can import with the passphrase.

## R3. Zip writing
- **Decision**: `archive/zip` streaming to a temp file next to the target, deflate, UTF-8 names
  with forward slashes (zip flag 0x800), then atomic rename; cancel removes the temp file.

## R4. Fidelity testing
- **Decision**: the stage-3 Postman fixture corpus is round-tripped in BDD and compared with the
  normaliser (export-formats §4). **Newman** (`newman` npm, pinned) runs exported collections
  against ApiMock in CI to prove Postman accepts and sends them (SC-002); Node is already in CI.
- **Alternatives**: comparing against Postman desktop (not automatable).

## R5. Xenon YAML import
- **Decision**: new readers in `webservice/xenonyaml/` producing `ImportItem`s for the XEN-003
  pipeline (bundle, zip manifest, single files); workspace zips are a special path: unpack into
  the chosen folder (must be empty) and register via `WorkspaceService.Open` semantics.
