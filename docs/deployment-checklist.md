# Docker Deployment Checklist

## Pre-Deployment Checklist

### Code Changes
- [x] GitHub Actions workflow created (`.github/workflows/docker-deploy.yml`)
- [x] Frontend Dockerfile configured with environment variable support
- [x] Backend Dockerfile ready
- [x] docker-entrypoint.sh configured for dynamic API_URL injection
- [x] config.js template created for frontend
- [x] Local development Makefiles use `:dev` tag

### Repository Configuration
- [ ] Verify your GitHub repository is set up and connected
- [ ] Check that your repository owner name is correct
- [ ] Optional: Make repository public if using ghcr.io without authentication

## First Deployment Steps

1. **Commit and Push to Main**
   ```bash
   git add .github/workflows/docker-deploy.yml
   git commit -m "feat: add GitHub Actions Docker deployment workflow"
   git push origin main
   ```

2. **Monitor the Workflow**
   - Go to your GitHub repository
   - Click on the "Actions" tab
   - You should see "Docker Build and Push" workflow running
   - Wait for both backend and frontend images to complete

3. **Verify Images Were Pushed**
   - Once the workflow completes, visit your packages:
     - `https://github.com/YOUR_USERNAME/bg-library/pkgs/container/bg-library-backend`
     - `https://github.com/YOUR_USERNAME/bg-library/pkgs/container/bg-library-frontend`

## Updating compose.yaml (When Ready)

When you want to use the automatically deployed images, update your `compose.yaml`:

```yaml
services:
  frontend:
    image: ghcr.io/YOUR_GITHUB_USERNAME/bg-library-frontend:latest
    # ... rest of config
  
  backend:
    image: ghcr.io/YOUR_GITHUB_USERNAME/bg-library-backend:latest
    # ... rest of config
```

## Local Development (No Changes Needed)

Continue using the local dev images:
```bash
make up    # Uses :dev tag images built locally
make down  # Stop containers
```

## Environment Variables Reference

### For compose.yaml deployment with GHCR images

```bash
# Frontend
export API_URL=http://localhost:8080
export BACKEND_PORT=8080
export EXPOSE_SWAGGER_UI=true
export NGINX_HOST=localhost

# Backend  
export DB_HOST=db
export DB_PORT=5432
export DB_USER=postgres
export DB_PASSWORD=your_password
export DB_NAME=bglib
export GIN_MODE=release
export CORS_ALLOWED_ORIGIN=*

# Optional
export APP_VERSION=latest
export FRONTEND_PORT=80
```

## Troubleshooting Failed Workflows

1. **Check the workflow logs**
   - GitHub → Actions → Docker Build and Push → Click on the failed run

2. **Common issues**
   - **Dockerfile not found**: Verify paths in `.github/workflows/docker-deploy.yml`
   - **Build failure**: Check if all COPY commands reference correct paths
   - **Authentication failure**: The `GITHUB_TOKEN` should work automatically

3. **Manual testing locally**
   ```bash
   # Build images locally with the same Docker commands
   cd /path/to/bg-library
   docker build -f backend/Dockerfile -t ghcr.io/your-username/bg-library-backend:latest .
   docker build -f frontend/Dockerfile -t ghcr.io/your-username/bg-library-frontend:latest .
   ```

## Next: Adding Version Tags (Future)

When you're ready to add semantic versioning:

1. Create a new workflow for tagged releases
2. Use git tags like `v1.0.0` to trigger image builds with version tags
3. Update the metadata step in the workflow to include version tags

See `docs/github-actions-setup.md` for more details.


