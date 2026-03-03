#!/bin/sh

# Default values for environment variables
NGINX_HOST=${NGINX_HOST:-localhost}
BACKEND_PORT=${BACKEND_PORT:-8080}
EXPOSE_SWAGGER_UI=${EXPOSE_SWAGGER_UI:-false}
API_URL=${API_URL:-http://localhost:8080}
TRUSTED_PROXIES=${TRUSTED_PROXIES:-}
REAL_IP_HEADER=${REAL_IP_HEADER:-X-Real-IP}

# SSL certificate paths
SSL_CERT_DIR="/etc/nginx/ssl"
SSL_CERT="${SSL_CERT_DIR}/cert.crt"
SSL_KEY="${SSL_CERT_DIR}/cert.key"

# Create SSL directory if it doesn't exist
mkdir -p "$SSL_CERT_DIR"

# Generate self-signed certificate if no certificate exists
if [ ! -f "$SSL_CERT" ] || [ ! -f "$SSL_KEY" ]; then
    echo "No SSL certificate found. Generating self-signed certificate..."
    openssl req -x509 -nodes -days 365 -newkey rsa:2048 \
        -keyout "$SSL_KEY" \
        -out "$SSL_CERT" \
        -subj "/C=US/ST=State/L=City/O=Organization/CN=${NGINX_HOST}" \
        2>/dev/null

    if [ $? -eq 0 ]; then
        echo "Self-signed SSL certificate generated successfully"
        chmod 644 "$SSL_CERT"
        chmod 600 "$SSL_KEY"
    else
        echo "ERROR: Failed to generate SSL certificate"
        exit 1
    fi
else
    echo "Using existing SSL certificate"
fi

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

    # Inject the proxy config after BOTH server_name lines (HTTP and HTTPS)
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
echo "Testing nginx configuration..."
nginx -t

if [ $? -ne 0 ]; then
    echo "ERROR: nginx configuration test failed"
    exit 1
fi

echo "Starting nginx..."
# Start nginx in the foreground
nginx -g 'daemon off;'
