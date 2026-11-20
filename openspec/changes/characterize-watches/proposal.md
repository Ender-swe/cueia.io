# Cueia recurring page watches

## Why

An owner can watch a public page and receive an in-app alert when a condition becomes true. A repeated observation or interrupted delivery must leave enough evidence to explain each alert.

## What Changes

### Observe and evaluate

Create an owner-bound watch with target URL, condition and value, interval, linked task ID, lifecycle state, and next due time. Use the browser page reader for both scheduled and manual checks. Save an observation with stable ID, actual URL, title, observed time, bounded text, and enough matching evidence to explain the result. Persist the previous baseline, last match, check count, consecutive failures, and delivery state.

| Condition | Evaluation | When to notify |
| --- | --- | --- |
| Page text changed | Compare normalized visible lines. Collapse whitespace, ignore relative-time-only changes, and summarize added, removed, and updated lines. | Use the first successful read as the baseline. Notify on each subsequent material difference. |
| Phrase appears | Trim and collapse whitespace in phrase and page text, and compare case-insensitively. | Notify on first observed presence and again after absence followed by presence. |
| Price below threshold | Recognize a USD amount with a dollar sign or USD marker, allowing intervening whitespace. Compare numeric amounts using a strict less-than threshold. | Notify on the first matching observation in each uninterrupted sequence of matches. Equality does not qualify. |

### Scheduling and delivery

Use the durable task scheduler to claim each due run. Save the new baseline and a recoverable alert intent for the same observation together. Identify delivery by watch and observation. A worker restarting after the baseline advances can finish publication. A replay finds the existing inbox item. Link the item to the observation, target, and watch. Persist the next eligible time so a restart cannot produce an extra scheduled check.

Support check now, pause, resume, and stop from Tracking or Activity. A manual check shares the watch’s claim and must preserve a newer baseline if another run completes first. On failure retain the last successful baseline and visible error, increment consecutive failures, and delay the next attempt with increasing backoff. Pause after five consecutive failures and record the pause notification once for that streak. Resume clears the pause and permits a new streak. Stop is terminal. Check the current lifecycle version when applying late results so an overlapping check or resume cannot reactivate a stopped watch.

### Acceptance

1. Serve a fictional price page through the fixture reader. Use a threshold of 50 and successive prices of 80, 40, 40, 80, and 40. Expect alert counts of 0, 1, 1, 1, and 2. Interrupt after saving the second observation and before publishing it. Restart, replay the run twice, and verify the two alerts link to the second and fifth observations.
2. With a fake clock, advance to each scheduled retry during five consecutive read failures. Verify increasing delays, a visible error, one pause notification, and zero scheduled reads while paused. Resume and complete one successful read. Pause and resume a second time and verify the baseline survives.
3. Hold a page read in flight, stop the watch, then release the read and restart. Expect stopped state and no new alert or eligible run. Test a first read already containing the phrase, a repeated match with mixed case and whitespace, and a price equal to the limit. Verify initial, repeated, and boundary behavior independently.

## Capabilities

### New Capabilities

- `page-watches` covers recurring observations, condition matching, recovery, and delivery of in-app alerts.

### Modified Capabilities

## Impact

Depends on durable scheduling, browser observations, and the personal workspace inbox. The watch owns matching and deduplication. The inbox owns read state and navigation from alert to observation.
