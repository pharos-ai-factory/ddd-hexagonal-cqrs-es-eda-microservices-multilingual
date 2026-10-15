# 0025 — Repository documentation site

Status: accepted, 15 October 2026. Extends decision 0018.

## Decision

Use Rspress **2.0.23**, pinned in the root development dependencies, for the
local documentation site. Existing Markdown files remain the authoritative
sources at their current paths. The site includes the root guides, contract
guides, context READMEs, decisions and verification evidence. A dedicated Education
section explains the concepts with café examples and a sequential learning path;
the existing architecture guides and decisions define the implementation rules.

Use the stable `rspress-plugin-mermaid` **1.0.3** community plugin, listed by
Rspress and declaring Rspress 2 compatibility. It renders Mermaid fences as SVGs
in the browser and follows the selected light/dark theme. Its browser renderer
avoids a Chromium dependency during builds. Configure Mermaid's strict security
mode. The pinned plugin currently uses Mermaid 10; review diagram syntax and
browser behaviour when upgrading it.

Use Rspress's built-in search, theme and Markdown exports. Additional terminology
and document-export plugins can be evaluated when there is a concrete need.

## Source links and navigation

The root configuration selects Markdown routes and excludes dependency, generated
and local artefact directories. Navigation uses the existing context catalogue
and discovers decision titles from their source files.

A build transform validates relative links against the checkout. Markdown pages
and directories with a README resolve within the site. Other source files and
directories link to their GitHub locations on `main`. The transform runs before
Rspress's link checks and applies to HTML and Markdown exports. Repository files
retain their original relative links for ordinary Markdown readers.

## Verification and limits

`pnpm docs:check` checks documentation tooling types, lint, link regression tests
and the static build. The complete `pnpm check` includes it, so the existing
verification job checks documentation on pull requests. Browser inspection proves
that Mermaid actually renders; its runtime diagrams are absent from static HTML
until JavaScript runs. Markdown exports retain the original Mermaid fences.

Builds write to ignored `.local/docs/generated/`; Rspress cache files are ignored.
Development and preview bind to loopback on port 3001. This decision adds local
documentation commands and verification, with no deployment workflow.
