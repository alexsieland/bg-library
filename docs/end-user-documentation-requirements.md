End-User Documentation Requirements

Purpose
This document defines the audience, tone, and update rules for end-user facing documentation for the Board Game Library application. It is intended for authors and maintainers who will update `MANUAL.md` and other end-user docs.

Audience and Tone
- Audience: non-technical staff and volunteers who will operate the library desk at an event. Readers are comfortable using a web browser but may not be familiar with technical terms.
- Reading level: keep language simple (approximate 5th–8th grade reading level).
- Tone: clear, respectful, and concise. Do not be patronizing. Assume competence but avoid unexplained jargon.
- UI references: describe common icons and their locations (for example: "search box near the top with a magnifying glass icon"), avoid using internal component names or developer-only terms.

Content Rules
- Keep content focused on the 99% cases. Do not attempt to document every rare edge case.
- Use numbered steps for workflows. Each step should be 1–2 short sentences where possible.
- Include a short introductory sentence for each workflow describing the goal.
- Add a small "Quick tips" section with 2–4 bullets for common small issues or time-savers.
- Include a short Troubleshooting & Contacts block listing the most common problems and the local escalation path (for example: "ask the desk lead" or follow the event's staff policy). Do not assume a specific organizational hierarchy — wording should allow events to adapt instructions to their local structure.

Update & Maintenance Rules
- The `MANUAL.md` file at the repository root is the canonical quick reference for front-desk workflows and must be kept up to date as the frontend UI changes.
- Whenever a new frontend feature affects the user workflows (for example: new scan behavior, new buttons or labels, new loan/return rules), update `MANUAL.md` to reflect how volunteers should use the feature.
- If the OpenAPI or frontend type generation changes labels or endpoints that are surfaced to end users, coordinate with the documentation owner to update `MANUAL.md`.
- Mark any in-progress or incomplete features in the manual clearly as "TBD (work in progress)" so desk staff know they may not be available yet.
- Keep the manual concise enough to print front-and-back on a single sheet. If the workflow grows beyond the single-sheet scope, create a longer guide but keep the quick reference current.

Contribution & Review
- Who can edit: any repository contributor may propose updates via a pull request. Use clear commit messages describing the workflow change and link to related frontend changes when possible.
- Review: at least one other contributor should review the change and confirm the steps match the UI before merging.
- Testing: where possible, test the workflow in the running frontend (local or event deployment) to ensure labels and button names match the manual.

File Maintenance Checklist (for PRs that change user-facing behavior)
1. Update `MANUAL.md` with new or changed workflow steps.
2. Mark incomplete features as "TBD (work in progress)".
3. Add or update examples or screenshots only when they are stable.
4. Add a brief note in the PR description pointing to the frontend change and the person who tested it.
5. Assign a reviewer and confirm the PR passes normal project checks.

Imperative
- It is imperative that `MANUAL.md` is kept current with frontend user-facing changes. Out-of-date quick-reference cards cause confusion at busy events.

Notes
- This document itself should be reviewed periodically (for example: yearly or when major UI changes occur) to ensure it still reflects writing guidance and update rules.
