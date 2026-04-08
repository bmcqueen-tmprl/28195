# Notes — Ticket 28195

_Created 2026-04-08_
Versions:
- SDK Go 1.37
- Server: OSS 1.30 UI 2.42.1 (Originally found in Cloud UI 2.47)

Temporal CLI doesn't support header propagation at all - See [CLI/983](https://github.com/temporalio/cli/issues/938)  
Additionally, existing schedule headers are truncated on Schedule Update via CLI

On any edit to a schedule via the UI (Workflow Type, task queue, schedule spec), the encoded headers are
iteratively encoded, leading to multiple levels of encoding.

##### Initially Set Schedule with Headers
```bash
❯ tmprl schedule describe -s schedule -o json
{
  # ...
    "action": {
      "startWorkflow": {
        "workflowId": "2d925f8f-6572-4e88-b2b1-c8fccb155620",
        "workflowType": {
          "name": "SomeWorkflow"
        },
        "taskQueue": {
          "name": "SomeTaskQueue",
          "kind": "TASK_QUEUE_KIND_NORMAL"
        },
        "input": {},
        "workflowExecutionTimeout": "0s",
        "workflowRunTimeout": "0s",
        "workflowTaskTimeout": "0s",
        "header": {
          "fields": {
            "custom-header": {
              "metadata": {
                "encoding": "anNvbi9wbGFpbg=="
              },
              "data": "eyJrZXkiOiJURVNULUhFQURFUiIsInZhbHVlIjoiVEVTVC1WQUxVRSJ9"
              # Decodes to:
              # {"key":"TEST-HEADER","value":"TEST-VALUE"}
            }
          }
        }
      }
    },
    # ...
}

```

#### Schedule after editing via UI
```bash
# ...
"action": {
      "startWorkflow": {
        "workflowId": "2d925f8f-6572-4e88-b2b1-c8fccb155620",
        "workflowType": {
          "name": "SomeWorkflow"
        },
        "taskQueue": {
          "name": "SomeTaskQueue",
          "kind": "TASK_QUEUE_KIND_NORMAL"
        },
        "input": {},
        "workflowExecutionTimeout": "0s",
        "workflowRunTimeout": "0s",
        "workflowTaskTimeout": "0s",
        "header": {
          "fields": {
            "custom-header": {
              "metadata": {
                "encoding": "anNvbi9wbGFpbg=="
              },
              "data": "eyJtZXRhZGF0YSI6eyJlbmNvZGluZyI6ImFuTnZiaTl3YkdGcGJnPT0ifSwiZGF0YSI6ImV5SnJaWGtpT2lKVVJWTlVMVWhGUVVSRlVpSXNJblpoYkhWbElqb2lWRVZUVkMxV1FVeFZSU0o5In0="
            }
          }
          # Decodes to:
          # {"metadata":{"encoding":"anNvbi9wbGFpbg=="},"data":"eyJrZXkiOiJURVNULUhFQURFUiIsInZhbHVlIjoiVEVTVC1WQUxVRSJ9"}
          # Note that this is the exact same information in the header pre-UIEdit
        }
      }
    },
# ...
```