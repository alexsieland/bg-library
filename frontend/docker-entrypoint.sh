#!/bin/sh

# Default values for environment variables
NGINX_HOST=${NGINX_HOST:-localhost}
BACKEND_PORT=${BACKEND_PORT:-8080}
EXPOSE_SWAGGER_UI=${EXPOSE_SWAGGER_UI:-false}
API_URL=${API_URL:-http://localhost:8080}
TRUSTED_PROXIES=${TRUSTED_PROXIES:-}
REAL_IP_HEADER=${REAL_IP_HEADER:-X-Real-IP}

# Start building the nginx config
CONFIG_FILE="/etc/nginx/conf.d/default.conf"

# First, substitute environment variables in the template
envsubst '$BACKEND_PORT $EXPOSE_SWAGGER_UI $NGINX_HOST' < /etc/nginx/conf.d/nginx.conf.template > "$CONFIG_FILE.tmp"

# If TRUSTED_PROXIES is set, inject the proxy trust configuration after the server_name line
if [ -n "$TRUSTED_PROXIES" ]; then
    echo "Configuring trusted proxies: $TRUSTED_PROXIES"
    echo "Using real IP header: $REAL_IP_HEADER"

    # Create a temporary file with proxy configuration
    PROXY_CONFIG=$(mktemp)

    # Add each trusted proxy IP/CIDR to the config
    echo "$TRUSTED_PROXIES" | tr ',' '\n' | while read -r proxy; do
        proxy=$(echo "$proxy" | xargs)  # Trim whitespace
        if [ -n "$proxy" ]; then
            echo "    set_real_ip_from $proxy;" >> "$PROXY_CONFIG"
        fi
    done

    # Add the real_ip_header directive
    echo "    real_ip_header $REAL_IP_HEADER;" >> "$PROXY_CONFIG"
    echo "" >> "$PROXY_CONFIG"

    # Inject the proxy config after the server_name line
    awk -v proxy_config="$(cat $PROXY_CONFIG)" '
        /server_name/ { print; print proxy_config; next }
        { print }
    ' "$CONFIG_FILE.tmp" > "$CONFIG_FILE"

    rm "$PROXY_CONFIG"
else
    mv "$CONFIG_FILE.tmp" "$CONFIG_FILE"
fi

# Inject API_URL into config.js
envsubst '$API_URL' < /usr/share/nginx/html/config.js.template > /usr/share/nginx/html/config.js

# Remove the template file after use
rm /usr/share/nginx/html/config.js.template

# Test nginx configuration
nginx -t

# Start nginx in the foreground
nginx -g 'daemon off;'
