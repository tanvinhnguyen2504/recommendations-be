# Changelog

## [0.1.2](https://github.com/tanvinhnguyen2504/recommendations-be/releases/tag/0.1.2) - 2026-11-03

- feat(swagger): add more field data `image` when check health

## [0.1.1](https://github.com/tanvinhnguyen2504/recommendations-be/releases/tag/0.1.1) - 2026-11-03
- [#PR1](https://github.com/tanvinhnguyen2504/recommendations-be/pull/1)
  - feat(ci): add CI and Artifact Registry release workflows

## 2026-10-25

 - docs(swagger): init swagger
 - chore: add psql seed targets and placeholder files
 - fix(config): honour `DATABASE_URL` and reject malformed env integers
 - feat(places): add places API, router and cursor pagination
 - feat(secret): add AES-256-GCM box for sealed values
 - feat(db): add pgx connection pool
 - feat(handler): add error envelope and request decoding
 - feat(config): add embedded layered YAML config
 - chore: add Go module, Makefile and container build