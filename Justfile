set quiet

alias b  := build
alias fb := force-build

# run full ci
[default]
[group('main')]
run: force-build log

# show logs from bot container
[group('dev')]
log:
    docker compose logs -f bot

# build docker
[group('dev')]
build *ARGS:
    docker compose up -d --build {{ARGS}}

# force rebuild docker
[group('dev')]
force-build:
    @just build --force-recreate bot