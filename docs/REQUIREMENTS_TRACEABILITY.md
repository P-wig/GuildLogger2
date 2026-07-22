# Requirements Traceability Matrix

This matrix maps GuildLogger2 requirements to design and implementation artifacts.

| Requirement | Description | Design Reference | Implementation Status |
|---|---|---|---|
| R-01 | Discord OAuth login | SOFTWARE_DESIGN_DOCUMENT.md (Key Workflows) | Implemented |
| R-02 | Guild connect and bot install | SOFTWARE_DESIGN_DOCUMENT.md (Key Workflows) | Implemented |
| R-03 | Member sync with role checks | SOFTWARE_DESIGN_DOCUMENT.md (Member Lifecycle Fields) | Implemented |
| R-04 | Event create/register/unregister | SRS.md (FR-04) | Implemented |
| R-05 | Hosted/participation stats | SRS.md (FR-05) | Implemented |
| R-06 | Anniversary notifications | SRS.md (FR-06) | Implemented |
| R-07 | Health endpoint and service readiness | API_REFERENCE.md (Implemented Endpoints) | Implemented |
| R-08 | Go backend runtime migration | SOFTWARE_DESIGN_DOCUMENT.md (Migration Notes) | Implemented |
| R-09 | Manual event log CRUD | API_REFERENCE.md (Guild Event Logs) | Implemented |

## Notes

- This matrix is intentionally concise during migration.
- Expand artifact links and test case IDs as endpoints are implemented.
