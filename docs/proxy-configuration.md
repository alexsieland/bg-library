# Proxy Configuration Guide

## Overview

The nginx frontend can be configured to trust proxy headers from reverse proxies like Cloudflare, AWS CloudFront, or custom load balancers. This is done via environment variables, making the configuration flexible for different deployment scenarios.

## Environment Variables

### `TRUSTED_PROXIES` (Optional)

A comma-separated list of IP addresses or CIDR ranges that should be trusted as proxies. When set, nginx will use these IPs to determine the real client IP address.

**Default**: Empty (no proxy trust configured)

**Examples**:
```bash
# Single IP
TRUSTED_PROXIES="203.0.113.1"

# Multiple IPs
TRUSTED_PROXIES="203.0.113.1,203.0.113.2"

# CIDR ranges
TRUSTED_PROXIES="10.0.0.0/8,172.16.0.0/12"

# Mix of IPs and CIDR ranges
TRUSTED_PROXIES="203.0.113.1,10.0.0.0/8,192.168.1.0/24"
```

### `REAL_IP_HEADER` (Optional)

The HTTP header that contains the real client IP address when behind a proxy.

**Default**: `X-Real-IP`

**Common values**:
- `X-Real-IP` - Standard reverse proxy header
- `CF-Connecting-IP` - Cloudflare specific header
- `X-Forwarded-For` - Standard proxy chain header (uses leftmost IP)
- `True-Client-IP` - Akamai, Cloudflare Enterprise

## Common Configurations

### No Proxy (Default)

When running without a reverse proxy, leave both variables unset:

```bash
# .env file - no proxy configuration needed
```

The application will use the direct connection IP addresses.

### Behind Cloudflare

Cloudflare provides the client IP in the `CF-Connecting-IP` header. You need to trust Cloudflare's IP ranges:

```bash
# .env file
TRUSTED_PROXIES=173.245.48.0/20,103.21.244.0/22,103.22.200.0/22,103.31.4.0/22,141.101.64.0/18,108.162.192.0/18,190.93.240.0/20,188.114.96.0/20,197.234.240.0/22,198.41.128.0/17,162.158.0.0/15,104.16.0.0/13,104.24.0.0/14,172.64.0.0/13,131.0.72.0/22
REAL_IP_HEADER=CF-Connecting-IP
```

For IPv6 support, add:
```bash
TRUSTED_PROXIES=173.245.48.0/20,103.21.244.0/22,103.22.200.0/22,103.31.4.0/22,141.101.64.0/18,108.162.192.0/18,190.93.240.0/20,188.114.96.0/20,197.234.240.0/22,198.41.128.0/17,162.158.0.0/15,104.16.0.0/13,104.24.0.0/14,172.64.0.0/13,131.0.72.0/22,2400:cb00::/32,2606:4700::/32,2803:f800::/32,2405:b500::/32,2405:8100::/32,2a06:98c0::/29,2c0f:f248::/32
REAL_IP_HEADER=CF-Connecting-IP
```

**Note**: Cloudflare IP ranges may change. Get the latest from: https://www.cloudflare.com/ips/

### Behind AWS CloudFront

```bash
# .env file
TRUSTED_PROXIES=CloudFront_IP_ranges  # Get from AWS documentation
REAL_IP_HEADER=X-Forwarded-For
```

AWS CloudFront IP ranges: https://docs.aws.amazon.com/AmazonCloudFront/latest/DeveloperGuide/LocationsOfEdgeServers.html

### Behind a Single Reverse Proxy

If you have a single nginx or HAProxy in front:

```bash
# .env file
TRUSTED_PROXIES=192.168.1.100  # IP of your reverse proxy
REAL_IP_HEADER=X-Real-IP
```

### Behind Docker Swarm/Kubernetes Ingress

Trust the internal network range:

```bash
# .env file
TRUSTED_PROXIES=10.0.0.0/8  # Docker/K8s internal network
REAL_IP_HEADER=X-Forwarded-For
```

## How It Works

When `TRUSTED_PROXIES` is set, the docker-entrypoint.sh script dynamically generates nginx configuration directives:

```nginx
set_real_ip_from 203.0.113.1;
set_real_ip_from 10.0.0.0/8;
real_ip_header X-Real-IP;
```

These directives tell nginx:
1. **Trust these proxy IPs** - Only accept real IP headers from these sources
2. **Use this header** - Extract the real client IP from this HTTP header

The real client IP is then available in nginx logs and passed to the backend via `X-Real-IP` and `X-Forwarded-For` headers.

## Security Considerations

### ⚠️ Important: Only Trust Known Proxies

**Do NOT** trust all IPs:
```bash
# DANGEROUS - Never do this!
TRUSTED_PROXIES=0.0.0.0/0
```

This would allow anyone to spoof their IP address by setting the real IP header.

### Trust Only What You Control

Only add IP ranges for proxies you control or trust:
- ✅ Your own reverse proxy
- ✅ CDN service you're using (Cloudflare, CloudFront)
- ✅ Load balancer in your infrastructure
- ❌ Random IP ranges
- ❌ Client IP ranges

### Keep Proxy IP Lists Updated

If using a CDN like Cloudflare:
1. Subscribe to their IP range updates
2. Periodically check for changes
3. Update your `.env` file accordingly

## Testing

### Verify Real IP Detection

1. **Check nginx access logs**:
   ```bash
   docker compose logs frontend | grep "GET"
   ```
   You should see real client IPs, not proxy IPs.

2. **Test with curl**:
   ```bash
   # Without trusted proxy configured
   curl -H "X-Real-IP: 1.2.3.4" http://your-app.com/api/endpoint
   # Should ignore the header and use actual connection IP

   # With trusted proxy configured and coming from that proxy
   # Should use 1.2.3.4 as the real IP
   ```

3. **Check backend logs**:
   Your backend should receive the correct client IP in the `X-Real-IP` header.

## Troubleshooting

### Issue: Still seeing proxy IPs in logs

**Cause**: Proxy IPs not in `TRUSTED_PROXIES` list or wrong `REAL_IP_HEADER`

**Solution**:
1. Verify the proxy IP: `docker compose logs frontend | grep "connect"`
2. Add that IP to `TRUSTED_PROXIES`
3. Rebuild: `docker compose down && docker compose up -d`

### Issue: Seeing unexpected IPs

**Cause**: Wrong proxy IP range or header

**Solution**:
1. Check what header your proxy uses
2. Verify the proxy IP ranges
3. Update `.env` with correct values

### Issue: Configuration not applying

**Cause**: Changes not picked up

**Solution**:
```bash
# Rebuild and restart
docker compose down
docker compose up -d

# Check the generated nginx config
docker compose exec frontend cat /etc/nginx/conf.d/default.conf
```

## Examples by Use Case

### Development (Local)
```bash
# .env
# No proxy configuration needed
```

### Staging (Behind single nginx proxy)
```bash
# .env
TRUSTED_PROXIES=192.168.1.10
REAL_IP_HEADER=X-Real-IP
```

### Production (Behind Cloudflare)
```bash
# .env
TRUSTED_PROXIES=173.245.48.0/20,103.21.244.0/22,103.22.200.0/22,103.31.4.0/22,141.101.64.0/18,108.162.192.0/18,190.93.240.0/20,188.114.96.0/20,197.234.240.0/22,198.41.128.0/17,162.158.0.0/15,104.16.0.0/13,104.24.0.0/14,172.64.0.0/13,131.0.72.0/22
REAL_IP_HEADER=CF-Connecting-IP
```

### Production (Behind AWS ALB + CloudFront)
```bash
# .env
TRUSTED_PROXIES=10.0.0.0/8  # VPC CIDR
REAL_IP_HEADER=X-Forwarded-For
```

## Additional Resources

- [Cloudflare IP Ranges](https://www.cloudflare.com/ips/)
- [AWS CloudFront IP Ranges](https://docs.aws.amazon.com/AmazonCloudFront/latest/DeveloperGuide/LocationsOfEdgeServers.html)
- [Nginx ngx_http_realip_module](http://nginx.org/en/docs/http/ngx_http_realip_module.html)
- [X-Forwarded-For Header](https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/X-Forwarded-For)

## Summary

The proxy trust configuration is **optional** and should only be configured when running behind a reverse proxy or CDN. For direct connections, leave both `TRUSTED_PROXIES` and `REAL_IP_HEADER` unset or empty.

This design allows the same Docker image to work in any environment without modification - just configure the environment variables appropriately for your deployment scenario.

