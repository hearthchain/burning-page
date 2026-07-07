---
purpose: Canonical weekly WAVES price journal (waves.csv) and how it is generated and consumed
---

# Weekly price journal

`waves.csv` is the canonical weekly price journal for WAVES (cmc_id 1274). It has two columns: `week_end`, an ISO date where weeks end Sunday UTC, and `price_avg_usd`, the average of daily closing prices within the week, sourced from CoinMarketCap. It covers the full history since listing, 527 rows.

The file is exported from the cto-agent repo's SQLite DB (`data/cmc_history.db`, table `price_weekly`, built by `scripts/backfill_cmc_prices.py`). Regenerate with `make journal`; the sqlite3 one-liner is in the Makefile.

The credit formula reads MaxSince(layer date) = the maximum `price_avg_usd` over all weeks with `week_end` >= that date.

The CSV decimal strings are the canonical values; parsers must not round-trip through floats.
