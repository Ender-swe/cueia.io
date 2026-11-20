# Cueia workspace design

## 1. Visual theme & atmosphere

Cueia is a dark, conversation-first workspace. The navigation and task lists are compact. The middle pane gives messages room to breathe, with open space between speakers and a clear path from each reply to its supporting email or calendar entry. Adjacent panes sit on the same plane. A slight change in charcoal and a hairline divider mark each boundary.

Two marks identify Cueia. Three open apricot rings form the small mark at the top of the navigation rail. A folded terracotta paper loop appears in a circular crop beside Cueia’s replies (21 px) and above its name in the desktop Tasks pane (118 px). [Artwork provenance](PROVENANCE.md)

## 2. Color palette & roles

| Color                                                                        | Where it appears                                                                                                              |
| ---------------------------------------------------------------------------- | ----------------------------------------------------------------------------------------------------------------------------- |
| Rail charcoal (#191918), canvas charcoal (#1B1B1A), panel charcoal (#1E1E1D) | The rail is darkest. The conversation occupies the canvas. Both side panes use the slightly lighter panel shade.              |
| Graphite (#383734)                                                           | One-pixel rules divide panes, the conversation header, and task sections.                                                     |
| Warm ivory (#F0EDE8) and soft stone (#AAA7A1)                                | Ivory carries headings and body text. Stone carries timestamps, placeholders, secondary icons, and empty states.              |
| Ash (#292927) and umber (#503B30)                                            | Cueia’s replies and email threads use ash. The user’s messages use umber with cream text (#F6EEE7).                           |
| Apricot (#E9AA83)                                                            | The brand mark, active Chat control, selected thread icon, attachment names, task link, and keyboard focus share this accent. |
| Pale sage (#B9DCA2)                                                          | A small dot beside “Connected” in the sample workspace. The private workspace pairs a stone dot with “Agent disconnected.”    |

The shared palette and the workspace-specific shades are defined in the [theme](packages/ui/tailwind.css) and [workspace stylesheet](apps/dashboard/app/workspace.css).

## 3. Typography rules

The workspace uses **DM Sans**. Standard interface text is 15 px with a line height of 1.5. Message text is slightly larger (16 px, line height 1.55). Navigation captions and timestamps are 11 px. Byline names are 13 px semibold, and timestamps use equal-width digits so times align.

Headings stay close to reading size. The conversation title is 17 px semibold with slightly tightened spacing (−0.025 em). Cueia’s profile name is the largest label (23 px, semibold, −0.04 em). Thread and task names use a medium weight (550), leaving the heavier weight (600) for headings and emphasis.

## 4. Component stylings

- **Navigation and buttons.** The rail stacks an icon above each label. An open destination gets a rounded charcoal tile (#33312E). The selected thread gets a full-width row (#33322F). Small add and close controls have transparent resting surfaces. Keyboard focus on buttons draws a 2 px apricot outline with a 2 px gap.
- **Messages and records.** Cueia’s turns align left and the user’s align right, with the byline directly above the bubble. The upper corner nearest the speaker is nearly square (3 px), while the other corners are 11 px. A calendar entry has a thin border, 8 px corners, and a date tile with a small month above a large day. An email thread groups its subject, sender, body, and attachments on one ash surface.
- **Search and composer.** Search is a compact field (36 px minimum height) with an inset magnifier and 8 px corners. The composer is a wider, warmer bar (#302F2C) with a thin border, 10 px corners, and a minimum height of 52 px. An attachment button sits to the left of the draft and a dimmed microphone to the right. Focus turns the border apricot. A centered caption beneath the bar explains the draft state.
- **Tasks and dialogs.** The Tasks pane centers the portrait, name, and connection status above left-aligned task and activity rows. Horizontal rules separate the sections. Preferences and scheduled-task details open in small dialogs. A PDF preview uses an extra-large dialog with the filename as its heading.

## 5. Layout principles

**Desktop.** From left to right, the workspace opens with a 72 px navigation rail, a resizable Chats pane, the conversation, and a resizable Tasks pane. Chat, Find, Files, Ideas, and Tasks sit in the middle of the rail, with Menu at the bottom. Find, Files, and Ideas replace the contents of the left pane. Tasks opens independently on the right.

**Reading area.** The 62 px header and bottom composer stay in place while the feed scrolls. Messages occupy a centered column up to 720 px wide. Each turn takes up to 90% of that width, capped at 580 px. The composer follows the same centerline and can reach 744 px.

**Compact screens.** At 760 px and below, the rail narrows to 54 px, the header to 50 px, and both side panes start closed. Opening Chat, Find, Files, or Ideas places a height-adjustable pane above the conversation. Tasks still opens on the right. Only one side pane is open at a time. When the conversation itself reaches 420 px or less, turns can use 96% of the reading column and the calendar entry drops its trailing icon. [Workspace layout](apps/dashboard/app/pages/chat.tsx) · [Pane sizing](apps/dashboard/app/resizable-sidebar.tsx)
