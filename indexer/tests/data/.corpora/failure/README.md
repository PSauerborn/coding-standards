# Failure-path fixture corpora

Every directory here is a complete, deliberately broken standards corpus. Each
one is indexed as its own `--source` root by `indexer/failure_paths_test.go`,
which asserts that the run exits non-zero and reports a diagnostic naming the
offending file.

**Do not repair these corpora.** A fixture that stops being broken makes the
test asserting its diagnostic pass vacuously.

**Do not move them out of `.corpora`.** `indexer/tests/data` lies inside the
source root the tool walks when it indexes this repository, and the dot prefix
on `.corpora` is the only thing keeping these documents out of that walk. Moved
anywhere the walker can reach, they would make the indexer report its own
repository as a broken corpus.
