ARG CADDY_VERSION=2.5.1

FROM caddy:${CADDY_VERSION}-builder AS builder

COPY . /workspace

RUN xcaddy build \
      --with github.com/thefrozenfire/caddy-ens=/workspace

FROM caddy:${CADDY_VERSION}

COPY --from=builder /usr/bin/caddy /usr/bin/caddy
