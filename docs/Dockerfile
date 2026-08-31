FROM node:24-alpine AS build-env

RUN apk add git ca-certificates
RUN corepack enable

WORKDIR /build

COPY package.json pnpm-workspace.yaml pnpm-lock.yaml ./
RUN pnpm install --frozen-lockfile

COPY . .

ENV NODE_ENV=production
RUN pnpm run build

FROM busybox
COPY --from=build-env /build/src/.vitepress/dist /build/out
RUN mkdir -p /www
CMD cp -r /build/out/* /www/
