# go-pg-extras

PostgreSQL database performance insights for Go. Locks, index usage, buffer cache hit ratios, vacuum stats, and more.

A Go port of [rails-pg-extras](https://github.com/pawurb/rails-pg-extras), providing the same powerful PostgreSQL diagnostic queries as a Go library and standalone CLI binary. This project is **not** officially affiliated with or endorsed by any of the upstream projects it draws from.

## Overview

`go-pg-extras` gives you easy access to a collection of useful PostgreSQL metadata queries that help diagnose performance issues, identify missing or unused indexes, check cache efficiency, inspect locks, and more. It can be used in two ways:

- **As a Go library** — import into your Go applications for programmatic access to PostgreSQL diagnostics.
- **As a standalone CLI binary** — run queries directly from the command line against any PostgreSQL database.

## Query Sources

The SQL queries used in this project originate from several sources:

- [heroku-pg-extras](https://github.com/heroku/heroku-pg-extras)
- [rails-pg-extras](https://github.com/pawurb/rails-pg-extras) (and its core dependency [ruby-pg-extras](https://github.com/pawurb/ruby-pg-extras))
- [PostgreSQL Unused Index Size](https://hakibenita.com/postgresql-unused-index-size) by Haki Benita
- [Useful SQLs to Check Contents of PostgreSQL shared_buffers](https://sites.google.com/site/itmyshare/database-tips-and-examples/postgres/useful-sqls-to-check-contents-of-postgresql-shared_buffer)
- [Index Maintenance](https://wiki.postgresql.org/wiki/Index_Maintenance) — PostgreSQL Wiki

## Disclaimer

This project is an independent Go port. It is not affiliated with, endorsed by, or officially connected to Heroku, the rails-pg-extras project, or any of the other sources listed above. All credit for the original queries and concepts belongs to their respective authors.

## License

MIT
