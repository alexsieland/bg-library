#!/bin/sh

# Override the nginx configuration with environment variables
envsubst '$BACKEND_PORT $EXPOSE_SWAGGER_UI $NGINX_HOST' < /etc/nginx/conf.d/nginx.conf.template > /etc/nginx/conf.d/default.conf

# Inject API_URL into config.js
envsubst '$API_URL' < /usr/share/nginx/html/config.js.template > /usr/share/nginx/html/config.js

# Start nginx in the foreground
nginx -g 'daemon off;'
