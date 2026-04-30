# Design System

GuildLogger2 frontend design guidance.

## Design Direction

GuildLogger2 is an operations dashboard for Discord gaming communities.
Visual language should be clear, data-dense, and actionable.

## UI Principles

1. Prioritize clarity over decoration.
2. Expose event status, participation, and reminders prominently.
3. Make role-based actions obvious (member, host, admin).
4. Keep interactions fast and predictable on desktop and mobile.

## Color Roles

- Primary: core actions and active navigation
- Success: completed/healthy operations
- Warning: pending/at-risk states
- Error: failures/invalid actions
- Neutral: informational and structural UI surfaces

## Typography

- Use a readable sans-serif stack.
- Ensure strong hierarchy for page title, section title, and card content.
- Avoid overly compact text in analytics cards and tables.

## Core Components

- App shell with persistent navigation
- Event cards (status, slots, host)
- Member profile summary cards
- Stats tiles (hosted, participated, tenure)
- Notification queue view
- Confirmation dialogs for state-changing actions

## Accessibility

- Keyboard navigation for all interactive controls
- Visible focus states
- WCAG-compliant contrast
- Error messages tied to form fields

## Responsive Behavior

- Mobile-first spacing and stacking
- Keep critical actions reachable without deep scrolling
- Collapse dense tables into summary cards on small screens

## Implementation Notes

- Use shared component primitives where possible.
- Keep page-specific styling minimal and isolated.
- Update this file when introducing new reusable patterns.
