# Error Codes

The gateway returns stable error codes in JSON error responses and records the full message in traces. Provider credentials are never included in error messages.

## Gateway Errors

| Code | Meaning |
| --- | --- |
| `invalid_request` | Request body or query parameters are invalid. |
| `unauthorized` | The app token is missing, unknown, or lacks the required grant. |
| `policy_denied` | A policy rule rejected the request before routing. |
| `route_failed` | No enabled and routable Provider supports the request under the current policy. |
| `provider_failed` | The Provider failed with an unclassified error. |

## Provider Errors

| Code | Meaning |
| --- | --- |
| `provider_missing_credential` | Required `api_key_env` is configured but the environment variable is missing or empty. |
| `provider_connection_failed` | The gateway could not connect to the upstream Provider. |
| `provider_timeout` | The upstream Provider request timed out. |
| `provider_unauthorized` | The upstream Provider returned `401` or `403`. |
| `provider_rate_limited` | The upstream Provider returned `429`. |
| `provider_upstream_status` | The upstream Provider returned another non-2xx status. |
| `provider_invalid_response` | The upstream Provider returned malformed or incompatible JSON. |
| `provider_empty_response` | The upstream Provider response contained no choices. |

## Management API Errors

| Code | Meaning |
| --- | --- |
| `reload_unavailable` | Runtime manager is not available in the current handler mode. |
| `reload_failed` | Config reload failed and the previous runtime snapshot remains active. |
| `runtime_unavailable` | Runtime manager is not available for Provider management or tool invocation. |
| `provider_update_failed` | Provider enable/disable config update failed. |
| `provider_probe_failed` | Provider probe failed; inspect Provider health for the classified reason. |
| `audit_unavailable` | Audit store is not configured. |
| `tool_denied` | Tool invocation was blocked by the current read-only MVP safety boundary. |
| `tool_scope_denied` | The app token is valid but does not include all required tool scopes. |
| `tool_unavailable` | Tool registry is not configured, or the selected tool dependency is unavailable. |
| `tool_failed` | Tool adapter execution failed with a stable tool SDK error or an unclassified error. |

Tool invocation errors include a `trace_id` when the request reached the tool invocation handler.
