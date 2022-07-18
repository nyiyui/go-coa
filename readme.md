TODO: impl JIT (estimate if running parallelly will make it faster)
TODO: think of ways to effectively use multiple cores
TODO: impl @catch
TODO: make @map concurrent
TODO: name debugger "Bifrost"
TODO: support list, string with some arithmetic operators (depends on dispatch)
TODO: add dispatching support to callables

## Debugger (Bifrost)

- Breakpoints (set using comments or flags)
  - `-b=filename[:line[:column]]-breakpoint_name`
  - `#coa:breakpoint breakpoint_name`
- Snapshots (automatically take and continue or pause)
  - Contents of all scopes
  - CPU usage, etc?
