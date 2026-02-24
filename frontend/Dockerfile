# Build stage
FROM node:25-alpine AS build

WORKDIR /app

COPY frontend/package*.json ./
COPY swagger/api.yaml api.yaml
RUN npm install

COPY frontend .

RUN npm openapi-typescript api.yaml --output src/generated/libray-api.d.ts
RUN npm run build

# Production stage
FROM nginx:alpine

COPY --from=build /app/dist /usr/share/nginx/html

EXPOSE 80

CMD ["nginx", "-g", "daemon off;"]
