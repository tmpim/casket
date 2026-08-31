FROM node:24-alpine AS build-env

RUN apk add git ca-certificates
RUN corepack enable

WORKDIR /build

COPY package.json pnpm-workspace.yaml pnpm-lock.yaml ./
RUN pnpm install --frozen-lockfile

COPY . .

FROM build-env AS build
ENV NODE_ENV=production
RUN pnpm run build

FROM build-env AS dev
CMD pnpm run dev --host

FROM busybox
COPY --from=build /build/src/.vitepress/dist /build/out
RUN mkdir -p /www
CMD cp -r /build/out/* /www/
