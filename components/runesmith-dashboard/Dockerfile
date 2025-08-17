FROM node:22.18-alpine AS build
WORKDIR /app

COPY components/runesmith-dashboard/package*.json ./
RUN --mount=type=cache,target=/root/.npm npm ci

COPY components/runesmith-dashboard/ .
ARG FULL_VERSION="dev"
ENV VITE_APP_VERSION=$FULL_VERSION
RUN npm run build

FROM nginx:1.27-alpine

ARG FULL_VERSION="dev"
LABEL name="runesmith-dashboard" \
      org.opencontainers.image.source="https://github.com/fukaraca/runesmith" \
      org.opencontainers.image.authors="@fukaraca" \
      org.opencontainers.image.version="${FULL_VERSION}"

COPY components/runesmith-dashboard/nginx.conf /etc/nginx/conf.d/default.conf
COPY --from=build /app/dist /usr/share/nginx/html

EXPOSE 80
CMD ["nginx", "-g", "daemon off;"]
