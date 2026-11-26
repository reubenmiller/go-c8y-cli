# Offline tests

Tests here send real requests (they create data the live tenant would keep, or need a running service).
They only run against the go-c8y fake server via `C8Y_OFFLINE_SUITE=offline tests/run-offline.sh [folder|file]`.
They are not part of the live suite (`tests/run-manual.sh`).
