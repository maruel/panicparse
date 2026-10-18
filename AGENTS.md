# panicparse

Run `make help` for available targets. Before committing, run `make fix`,
`make verify`, and `make test`. Install Git hooks with `make git-hooks`.

`make test` runs coverage and benchmark smoke checks.
Use `make test-race` for race checks.

Run Python tools and scripts with `uv run --locked`. Refresh the file index with
`uv run --locked python scripts/update_agents_file_index.py` after adding or
removing indexed files or changing their summaries.
