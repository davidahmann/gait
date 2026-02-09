# SEO and AEO Checklist (OSS Site)

Date: 2026-02-09

This checklist tracks what Gait’s public docs site implements today for search and assistant discovery.

## Baseline Principles

- Keep core SEO fundamentals first (crawlability, internal links, clear text content, canonical metadata).
- Treat AEO as an extension of SEO, not a separate stack.
- Keep all structured data aligned with visible page content.

## Implemented

- Canonical + OpenGraph metadata in site layout (`docs-site/src/app/layout.tsx`).
- Crawl directives and sitemap references (`docs-site/public/robots.txt`).
- XML sitemap for primary public routes (`docs-site/public/sitemap.xml`).
- Assistant-oriented index and context pages (`docs-site/public/llms.txt`, `docs-site/public/llm/*.md`, `docs-site/public/ai-sitemap.xml`).
- FAQ content on homepage plus `FAQPage` JSON-LD (`docs-site/src/app/page.tsx`).
- `SoftwareApplication` JSON-LD on homepage (`docs-site/src/app/page.tsx`).

## Notes on Feature Expectations

- Google AI features do not require special “AI-only” markup beyond normal SEO best practices.
- FAQ rich results are currently limited by Google eligibility rules (not guaranteed for general software sites).
- OpenAI discovery in ChatGPT Search depends on allowing `OAI-SearchBot` crawl access.

## Source References

- Google AI features guidance:
  - https://developers.google.com/search/docs/appearance/ai-features
- Google FAQ structured data:
  - https://developers.google.com/search/docs/appearance/structured-data/faqpage
- Google canonicalization guidance:
  - https://developers.google.com/search/docs/crawling-indexing/canonicalization
- Google sitemap guidance:
  - https://developers.google.com/search/docs/advanced/sitemaps/overview
- OpenAI crawler documentation:
  - https://platform.openai.com/docs/bots
- Schema.org `SoftwareApplication`:
  - https://schema.org/SoftwareApplication
