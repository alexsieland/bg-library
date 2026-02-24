# This dockerfile is expected to be built from the root of the project, so the context includes both the frontend and sawgger directories.

# Build stage
FROM node:25-alpine AS build

WORKDIR /app

COPY frontend/package*.json ./
COPY swagger/api.yaml api.yaml
RUN npm install

COPY frontend/ ./

RUN npx openapi-typescript api.yaml --output src/generated/api.d.ts
RUN npm run build
RUN npm prune --production

# Production stage
FROM nginx:stable-alpine

COPY frontend/nginx.conf /etc/nginx/conf.d/default.conf

COPY --from=build /app/build /usr/share/nginx/html

EXPOSE 80

CMD ["nginx", "-g", "daemon off;"]
