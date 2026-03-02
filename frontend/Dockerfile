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

COPY frontend/nginx.conf /etc/nginx/templates/default.conf.template
COPY swagger/api.yaml /etc/nginx/templates/api.yaml.template

# Install Swagger UI
ENV SWAGGER_UI_VERSION=5.18.2
RUN wget -qO- https://github.com/swagger-api/swagger-ui/archive/refs/tags/v$SWAGGER_UI_VERSION.tar.gz | tar xz \
    && mkdir -p /usr/share/nginx/html/swagger \
    && cp -R swagger-ui-$SWAGGER_UI_VERSION/dist/* /usr/share/nginx/html/swagger/ \
    && rm -rf swagger-ui-$SWAGGER_UI_VERSION \
    && sed -i 's|https://petstore.swagger.io/v2/swagger.json|/api.yaml|g' /usr/share/nginx/html/swagger/swagger-initializer.js

COPY --from=build /app/dist /usr/share/nginx/html

EXPOSE 80

CMD ["/bin/sh", "-c", "envsubst '${NGINX_HOST} ${BACKEND_PORT} ${EXPOSE_SWAGGER} ${API_URL}' < /etc/nginx/templates/api.yaml.template > /usr/share/nginx/html/api.yaml && /docker-entrypoint.sh nginx -g 'daemon off;'"]
