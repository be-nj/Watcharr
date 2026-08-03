# About this fork

This is a personal fork of [sbondCo/Watcharr](https://github.com/sbondCo/Watcharr),
maintained by [@be-nj](https://github.com/be-nj) for their own instance.

## Transparency: AI-assisted development

**The changes in this fork are largely written by AI** (Claude by Anthropic), supervised,
reviewed and directed by @be-nj. In the spirit of upstream's CONTRIBUTING.md (which asks
for AI disclosure in PRs), this fork discloses AI usage as visibly as possible:

- Every AI-authored commit carries a `Co-Authored-By: Claude` trailer
- Issues and PRs note AI authorship explicitly
- Nothing from this fork is submitted upstream without full disclosure

## What this fork adds

See issues for details and status:

- **Fix**: watch providers deduplication (duplicate provider names crash detail pages)
- **Per-watch metadata** (#1): where/how you watched (watch sources incl. cinemas with
  screens and optional ratings), audio language & subtitles, per-watch tags and notes
- **Letterboxd import** (#2): diary/ratings/watched CSVs with per-watch tags and
  sensible merging into existing data
- ~~OIDC login~~ (#3): not needed — solved with upstream's trusted header auth
  behind an Authentik proxy provider instead
- Maybe later (#4): a map page for visited cinemas (Leaflet + OSM)

Everything is built to stay close to upstream conventions so individual features could
be offered upstream later (with AI disclosure), and to keep the fork easily rebasable.
