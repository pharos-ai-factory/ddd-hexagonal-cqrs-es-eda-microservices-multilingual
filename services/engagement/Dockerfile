FROM node:24-alpine AS build
RUN npm install --global pnpm@10.34.5
WORKDIR /repo
COPY package.json pnpm-workspace.yaml pnpm-lock.yaml ./
COPY services/engagement/package.json services/engagement/package.json
COPY services/web/package.json services/web/package.json
RUN pnpm install --frozen-lockfile
COPY services/engagement services/engagement
RUN pnpm --filter @cafe/engagement build
USER node
WORKDIR /repo/services/engagement
CMD ["node", "dist/apps/service.js"]
