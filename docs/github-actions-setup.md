# GitHub Actions Docker Deployment Setup

## Overview
Your GitHub Actions workflow is now configured to automatically build and push Docker images to GitHub Container Registry (ghcr.io) on every merge to the `main` branch.

## What Was Set Up

### Workflow File
- **Location**: `.github/workflows/docker-deploy.yml`
- **Trigger**: Automatically runs on any push to the `main` branch
- **Images Built**:
  - `bg-library-backend` → `ghcr.io/<your-username>/bg-library-backend`
  - `bg-library-frontend` → `ghcr.io/<your-username>/bg-library-frontend`
- **Build Strategy**: Both images build in parallel for faster deployment

### Image Tagging
Images are tagged with:
- `latest` - on the main branch (default branch)
- `main-<commit-sha>` - short commit SHA for traceability

### Container Registry
- **Registry**: GitHub Container Registry (ghcr.io)
- **Authentication**: Uses `GITHUB_TOKEN` (automatically available in GitHub Actions)
- **Permissions**: No additional secrets needed to configure

## What You Don't Need to Do
✅ No Docker Hub credentials needed  
✅ No additional secrets to configure  
✅ No manual image builds or pushes  

## Next Steps When You're Ready

### 1. Update compose.yaml (When Ready to Use GHCR Images)
Replace the Docker Hub image references with your GHCR images:

```yaml
services:
  frontend:
    image: ghcr.io/<your-github-username>/bg-library-frontend:latest
  backend:
    image: ghcr.io/<your-github-username>/bg-library-backend:latest
```

### 2. Make Your Repository Public (Optional)
If your GitHub repository is private, you'll need to authenticate when pulling images. For public repos, no authentication is needed.

### 3. Local Development
Continue using the `:dev` tag locally for development:
```bash
make up          # Builds and runs :dev images locally
make build       # Just builds :dev images
```

The Makefiles are already configured with:
- Backend: `alexsieland/bg-library-backend:dev`
- Frontend: `alexsieland/bg-library-frontend:dev`

## How It Works

1. You push/merge code to `main` branch
2. GitHub Actions automatically triggers the workflow
3. Both Docker images are built in parallel
4. Images are pushed to ghcr.io with appropriate tags
5. Images are ready to deploy

## Viewing Your Pushed Images

After the first workflow run, you can find your images at:
- https://github.com/<your-username>/bg-library/pkgs/container/bg-library-backend
- https://github.com/<your-username>/bg-library/pkgs/container/bg-library-frontend

## Environment Variables for Deployment

When you deploy these images, remember to set:

**Frontend**:
- `API_URL` - The backend API endpoint (e.g., `http://your-backend:8080`)
- `BACKEND_PORT` - The backend port (default: 8080)
- `EXPOSE_SWAGGER_UI` - Whether to expose Swagger UI (default: false)
- `NGINX_HOST` - The nginx server name (default: localhost)

**Backend**:
- `DB_HOST` - Database host
- `DB_PORT` - Database port
- `DB_USER` - Database user
- `DB_PASSWORD` - Database password
- `DB_NAME` - Database name
- `GIN_MODE` - Gin framework mode (release/debug)
- `CORS_ALLOWED_ORIGIN` - CORS origin (default: *)

## Future Enhancements

When you're ready to add versioning:
1. Add git tag handling in the workflow (e.g., `v1.0.0`)
2. Update the tags section in the metadata step
3. Consider semantic versioning for images

## Troubleshooting

If the workflow fails:
1. Check the GitHub Actions tab in your repository for logs
2. Verify your repository has the necessary permissions (Settings → Actions)
3. Ensure the Dockerfiles are valid (they should be working locally with `make build`)
4. Check that all required files are being copied in the Dockerfiles


