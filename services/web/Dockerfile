FROM node:24-alpine AS build
RUN npm install --global pnpm@10.34.5
WORKDIR /repo
COPY package.json pnpm-workspace.yaml pnpm-lock.yaml ./
COPY services/engagement/package.json services/engagement/package.json
COPY services/web/package.json services/web/package.json
RUN pnpm install --frozen-lockfile
COPY services/web services/web
RUN pnpm --filter @cafe/web build

FROM nginx:1.28-alpine
COPY --from=build /repo/services/web/out /usr/share/nginx/html
COPY devops/ingress/default.conf /etc/nginx/conf.d/default.conf
