# Linux CI image: the GTK 4 / WebKitGTK 6 build dependencies Wails needs, the snippet runtimes the BDD suite runs
# (curl, python3 with requests) and the tools the cache actions use. Go and Node come from setup-go / setup-node.
# Jobs run as uid 1001 (the hosted runner's user) so workspace permissions and read-only-path scenarios behave as
# on a plain runner. ci.yml tags the image with this file's hash and rebuilds it only when the file changes.
FROM ubuntu:24.04

ENV DEBIAN_FRONTEND=noninteractive

RUN rm -f /var/lib/man-db/auto-update \
    && printf 'path-exclude /usr/share/doc/*\npath-exclude /usr/share/man/*\npath-exclude /usr/share/info/*\n' > /etc/dpkg/dpkg.cfg.d/01-nodoc \
    && apt-get -o Acquire::Retries=3 update \
    && apt-get -o Acquire::Retries=3 install -y --no-install-recommends \
        build-essential pkg-config libgtk-4-dev libwebkitgtk-6.0-dev \
        ca-certificates curl git make openssh-client python3 python3-requests xz-utils zstd \
    && rm -rf /var/lib/apt/lists/* \
    && pkg-config --exists gtk4 webkitgtk-6.0

RUN userdel -r ubuntu 2>/dev/null || true \
    && useradd --create-home --uid 1001 --shell /bin/bash runner

USER runner
