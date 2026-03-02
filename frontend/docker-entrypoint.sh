#!/bin/sh

# Override the nginx configuration with environment variables
envsubst '$BACKEND_PORT $EXPOSE_SWAGGER_UI $NGINX_HOST' < /etc/nginx/conf.d/nginx.conf.template > /etc/nginx/conf.d/default.conf

# Start nginx in the foreground
nginx -g 'daemon off;'
