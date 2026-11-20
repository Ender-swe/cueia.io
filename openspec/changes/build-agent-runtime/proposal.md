# Cueia agent runtime and interactive choices

## Why

An owner should see which tool supplied an answer, whether work continues after chat stops, and which decisions still need attention. A model’s statement that it finished a job is insufficient evidence of completion.

## What Changes

### Conversation and task execution

Serve an authenticated CopilotKit runtime under Bun. Stream AG-UI turn start, assistant text, tool calls with stable call IDs, corresponding results, and completion or failure. In scripted mode, return repeatable turns over the fictional workspace. In model mode, assemble the owner’s messages and available tools for OpenAI, Anthropic, Gemini, or a configured compatible gateway. Route remote AG-UI conversations to their configured server with server-held credentials. Report which backend handled the turn.

Preserve provider message roles, tool arguments, and call/result pairing across follow-up requests. Support the configured OpenAI-compatible Responses or Chat Completions wire format. Limit steps and attempts. Retry transient failures that occur before output within that budget. After text has reached the client or a tool has produced an effect, preserve that progress and surface an interruption without replaying delivered work. Tell the user when a step limit ends a turn.

Delegate by creating a durable task and returning its ID in the tool result. Persist its plan, step outcomes, evidence, previously supplied answers, and remaining work. Pause for missing input or a reviewed action. Resume from the recorded answer or receipt. Only mark success when the expected artifact, receipt, or other observable result exists. Propagate cancellation to in-flight provider and tool requests.

### Tools and evidence

Expose mail and calendar reads, page reads, files, desktop, search, delegation, and reviewed actions according to backend and connection capabilities. Route writes through the action ledger. Bind each result to its call, source, owner, thread or task, and turn. Record a page title, URL and bounded excerpt with the observation. Preserve task and file IDs so history can open durable results even after temporary access links expire.

Use Parallel MCP for configurable public search. Apply `WEB_SEARCH_ENABLED` on both API and worker. Validate the query, bound response bytes, excerpt size and result count, and accept only valid HTTP(S) source links. Reuse a search session within a chat or task. Explain that a query and supplied context go to the search service. Distinguish an empty valid result from malformed data, cancellation, deadline, rate limit, or provider failure.

For model-driven chat, allow reads from the connected primary calendar with explicit start, end, and time zone. Ask for missing details, limit the range to 366 days, return at most 20 events, and identify partial results. For search, reject a service response exceeding 1 MiB and truncate an individual source excerpt at 30,000 characters while preserving its URL and truncation flag.

### Choice panels and optional OpenBot adapter

Support Jev off, scripted, and live modes. Off makes no choice-service request. Scripted decisions identify themselves in the UI. Live decisions use a server-side TypeSafe key and may use the latest message, agent context, and candidate options. Offer clarification, comparison, and refinement from the current turn’s evidence. Save the panel version, options, evidence IDs, selected option, and owner/thread/turn binding. An accepted selection is single-use. Superseded controls remain readable in history. Discard a late inference if selection or a newer turn has overtaken it.

Keep OpenBot disabled by default. An enabled contract adapter accepts an injected authenticated transport. Discovery must verify identity and granted capabilities before reporting availability. Return an Intelligence runtime descriptor for a compatible client. Preserve distinct channel, thread, agent, computer Bot, and action IDs. Support channel creation, computer status, snapshots, navigation, and request/take/release of human control through the authorized gateway. Preserve policy refusals and mark ambiguous mutation responses uncertain. Defer any claim of live integration until measured.

### Acceptance

1. Have a scripted provider read one fictional mail thread, emit a tool result, and delegate a reply. Close chat, supply task input, approve once, and resume. Inspect the stream, saved task, and receipt. Each must identify the same job and attachment.
2. Fail a provider before first output and verify bounded retry. Fail it after partial output or a successful tool call and verify no duplicate text or tool effect. Stop a search in flight and verify the originating operation observes cancellation. Run the same search disabled and expect zero transport calls.
3. Generate two sourced options, refine, select, and replay after restart. Only the current owner and thread may select a current option. Earlier controls remain spent after replay. Exercise a stale inference returning after selection.
4. Count zero adapter calls while disabled. Through a transport double, return distinct channel and thread IDs, deny navigation during human control, then interrupt a mutation response. Check that the refusal is visible and the ambiguous mutation is issued once.

## Capabilities

### New Capabilities

- `agent-runtime` covers configured backends, delegated model execution, and tool results.
- `source-search` covers cancellable searches and source evidence.
- `conversation-choices` covers evidence-backed panels and persisted selections.
- `external-agent-adapter` covers optional transport discovery and handover.

### Modified Capabilities

## Impact

Depends on workspace access and durable tasks. Conversation history stores tool results and selection identities. The clients open these results and the action ledger authorizes external changes.
