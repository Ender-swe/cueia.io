# Cueia ideas, goals, memory, and saved results

## Why

A person can find unfinished work, see why the agent suggested it, and return to its evidence and results after closing the app. A completed job should update its linked goal without losing progress the person recorded manually.

## What Changes

### Suggestions, goals, and preferences

Create suggestions from identifiable sources, including actionable incoming mail and unfinished goals. Save source IDs, explanation, editable request, and disposition. Omit outgoing replies and requests already completed. Accept, edit, or dismiss a suggestion from Ideas. If an active task already handles that message, open it. Keep an existing goal association, or create the goal and task needed for a newly accepted idea. Resolve competing decisions atomically. Dismissal must leave no partially created task or goal.

Store owner-scoped goals, milestones, completion state, task associations, and evidence. Apply each successful task outcome once to its linked milestone while preserving manual progress. Failed, cancelled, or waiting tasks leave completion unchanged. Persist display name, avatar and color, tone, editable memories, and background-update preference. Forgetting a memory removes it from future context and subsequent reads.

Persist in-app notifications with owner, source event, task or observation ID, timestamp, and read state. Opening an item leads to its source. A restarted producer can deliver the same source event without adding another item. List searchable Apps by capability and actual status so the owner can distinguish available, disconnected, disabled, and planned integrations.

### Results and transaction imports

Save plans, reports, comparisons, and finance summaries as typed artifacts with stable IDs, source evidence, generating task, and creation time. Reopen the same artifact from Activity, a goal, or a historical chat card. Finance detail includes the date range, transaction count, income, spending, remaining balance, category totals, and rows supporting the calculation.

Import UTF-8 CSV up to 500,000 bytes inclusive. Accept an initial byte-order mark, reordered columns, CR/LF/CRLF line endings, quoted fields with commas or line breaks, and doubled quotes. Match required headers `date`, `description`, `amount`, and `category` after trimming and case normalization. Reject missing or duplicate required headers, malformed quoting, invalid dates, or ambiguous amounts before saving a report. Ignore unrelated columns. Parse signed decimal amounts into integer cents and validate real calendar dates. Count positive values as spending and the absolute value of negative amounts as income. Compute remaining as income minus spending and determine the period from valid transaction dates.

### Acceptance

1. Race two accepts of the same idea from separate requests. Read one accepted suggestion, one linked task, and one appropriate goal. For a suggestion about a message already being processed, both requests must return the existing task. Race dismissal with acceptance and verify only the winning disposition’s records exist.
2. Mark one milestone complete manually, complete a task on another milestone, and replay the outcome. Restart and reopen the goal and associated artifact from Activity. Both milestones retain their progress. Change the tone, forget a memory, and mark an alert read. Verify these changes after relaunch.
3. Import a three-row Cueia fixture with amounts `-100.00`, `10.01`, and `0.02`. Expect income 100.00, spending 10.03, and remaining 89.97. Reorder columns and quote a description containing a comma. Totals must agree. Add a second normalized amount header and require a validation error with no artifact. Exercise a valid 500,000-byte file and reject one byte more.

## Capabilities

### New Capabilities

- `personal-context` covers identity, memory, and evidence-backed suggestions.
- `goals-and-activity` covers milestones, linked work, notifications, and read state.
- `task-artifacts` covers saved reports, plans, comparisons, and imported finance results.

### Modified Capabilities

## Impact

Depends on workspace storage, mail evidence, and durable task outcomes. Supplies the notification inbox to watches and saved artifact details to the web client. Finance uses user-supplied files. Bank feeds and device push are future extensions.
