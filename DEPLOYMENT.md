# Deployment Guide

## Overview

This guide covers deployment procedures for the workflow management application, including environment setup, build processes, and production deployment strategies.

## Architecture

### Frontend (Next.js)
- **Framework**: Next.js 15 with React 19
- **Deployment**: Static export or Server-Side Rendering
- **CDN**: Optimized for global content delivery
- **PWA**: Progressive Web App capabilities

### Backend (Go)
- **Framework**: Go with Gin/Echo (if implemented)
- **Database**: PostgreSQL with connection pooling
- **Cache**: Redis for session management
- **API**: RESTful API with JWT authentication

## Environment Setup

### Required Environment Variables

#### Frontend (.env.local)
```bash
# Supabase Configuration
NEXT_PUBLIC_SUPABASE_URL=your_supabase_url
NEXT_PUBLIC_SUPABASE_ANON_KEY=your_supabase_anon_key

# Application Configuration
NEXT_PUBLIC_APP_ENV=production
NEXT_PUBLIC_API_URL=https://api.yourdomain.com

# Analytics (optional)
NEXT_PUBLIC_GA_ID=your_google_analytics_id
```

#### Backend (.env)
```bash
# Database
DATABASE_URL=postgresql://user:password@host:port/database
REDIS_URL=redis://host:port

# Authentication
JWT_SECRET=your_jwt_secret
GITHUB_CLIENT_ID=your_github_client_id
GITHUB_CLIENT_SECRET=your_github_client_secret

# Application
PORT=8080
APP_ENV=production
```

### Security Configuration
```bash
# Security Headers
CSP_ENABLED=true
HTTPS_REDIRECT=true
CORS_ORIGINS=https://yourdomain.com

# Rate Limiting
RATE_LIMIT_ENABLED=true
RATE_LIMIT_REQUESTS=100
RATE_LIMIT_WINDOW=900
```

## Build Process

### Frontend Build

#### Development
```bash
cd frontend
npm install
npm run dev
```

#### Production Build
```bash
cd frontend
npm install --production=false
npm run build
npm start
```

#### Static Export (for CDN deployment)
```bash
cd frontend
npm run build
npm run export
# Output: out/ directory
```

### Backend Build

#### Development
```bash
cd backend
go mod tidy
go run main.go
```

#### Production Build
```bash
cd backend
CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .
```

## Docker Deployment

### Frontend Dockerfile
```dockerfile
FROM node:18-alpine AS builder
WORKDIR /app
COPY package*.json ./
RUN npm ci --only=production && npm cache clean --force
COPY . .
RUN npm run build

FROM node:18-alpine AS runner
WORKDIR /app
ENV NODE_ENV=production
RUN addgroup --system --gid 1001 nodejs
RUN adduser --system --uid 1001 nextjs
COPY --from=builder /app/public ./public
COPY --from=builder --chown=nextjs:nodejs /app/.next/standalone ./
COPY --from=builder --chown=nextjs:nodejs /app/.next/static ./.next/static
USER nextjs
EXPOSE 3000
ENV PORT 3000
CMD ["node", "server.js"]
```

### Backend Dockerfile
```dockerfile
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .

FROM alpine:latest
RUN apk --no-cache add ca-certificates tzdata
WORKDIR /root/
COPY --from=builder /app/main .
EXPOSE 8080
CMD ["./main"]
```

### Docker Compose
```yaml
version: '3.8'

services:
  frontend:
    build: 
      context: ./frontend
      dockerfile: Dockerfile
    ports:
      - "3000:3000"
    environment:
      - NODE_ENV=production
      - NEXT_PUBLIC_API_URL=http://backend:8080
    depends_on:
      - backend
    networks:
      - app-network

  backend:
    build:
      context: ./backend  
      dockerfile: Dockerfile
    ports:
      - "8080:8080"
    environment:
      - APP_ENV=production
      - DATABASE_URL=postgresql://postgres:password@db:5432/workflow
      - REDIS_URL=redis://redis:6379
    depends_on:
      - db
      - redis
    networks:
      - app-network

  db:
    image: postgres:15-alpine
    environment:
      - POSTGRES_DB=workflow
      - POSTGRES_USER=postgres
      - POSTGRES_PASSWORD=password
    volumes:
      - postgres_data:/var/lib/postgresql/data
    networks:
      - app-network

  redis:
    image: redis:7-alpine
    volumes:
      - redis_data:/data
    networks:
      - app-network

volumes:
  postgres_data:
  redis_data:

networks:
  app-network:
    driver: bridge
```

## Cloud Deployment

### Vercel (Recommended for Frontend)

#### Configuration (vercel.json)
```json
{
  "version": 2,
  "builds": [
    {
      "src": "frontend/package.json",
      "use": "@vercel/next"
    }
  ],
  "env": {
    "NEXT_PUBLIC_SUPABASE_URL": "@supabase_url",
    "NEXT_PUBLIC_SUPABASE_ANON_KEY": "@supabase_anon_key"
  },
  "headers": [
    {
      "source": "/(.*)",
      "headers": [
        {
          "key": "X-Frame-Options",
          "value": "DENY"
        },
        {
          "key": "X-Content-Type-Options", 
          "value": "nosniff"
        }
      ]
    }
  ]
}
```

#### Deployment Steps
```bash
# Install Vercel CLI
npm i -g vercel

# Login to Vercel
vercel login

# Deploy
cd frontend
vercel --prod
```

### Railway/Fly.io (Backend)

#### Railway Deployment
```bash
# Install Railway CLI
npm install -g @railway/cli

# Login
railway login

# Initialize project
railway init

# Deploy
railway up
```

#### Fly.io Configuration (fly.toml)
```toml
app = "workflow-backend"
primary_region = "nrt"

[build]
  builder = "heroku/buildpacks:20"

[[services]]
  http_checks = []
  internal_port = 8080
  processes = ["app"]
  protocol = "tcp"
  script_checks = []

  [services.concurrency]
    hard_limit = 25
    soft_limit = 20
    type = "connections"

  [[services.ports]]
    force_https = true
    handlers = ["http"]
    port = 80

  [[services.ports]]
    handlers = ["tls", "http"]
    port = 443

  [[services.tcp_checks]]
    grace_period = "1s"
    interval = "15s"
    restart_limit = 0
    timeout = "2s"
```

### AWS/GCP/Azure

#### AWS ECS with Fargate
```yaml
# ecs-task-definition.json
{
  "family": "workflow-app",
  "networkMode": "awsvpc",
  "requiresCompatibilities": ["FARGATE"],
  "cpu": "256",
  "memory": "512",
  "executionRoleArn": "arn:aws:iam::ACCOUNT:role/ecsTaskExecutionRole",
  "taskRoleArn": "arn:aws:iam::ACCOUNT:role/ecsTaskRole",
  "containerDefinitions": [
    {
      "name": "frontend",
      "image": "your-registry/workflow-frontend:latest",
      "portMappings": [
        {
          "containerPort": 3000,
          "protocol": "tcp"
        }
      ],
      "environment": [
        {
          "name": "NODE_ENV",
          "value": "production"
        }
      ],
      "logConfiguration": {
        "logDriver": "awslogs",
        "options": {
          "awslogs-group": "/ecs/workflow-frontend",
          "awslogs-region": "us-west-2",
          "awslogs-stream-prefix": "ecs"
        }
      }
    }
  ]
}
```

## CI/CD Pipeline

### GitHub Actions

#### Frontend Deployment
```yaml
name: Deploy Frontend

on:
  push:
    branches: [main]
    paths: ['frontend/**']

jobs:
  deploy:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      
      - name: Setup Node.js
        uses: actions/setup-node@v4
        with:
          node-version: '18'
          cache: 'npm'
          cache-dependency-path: frontend/package-lock.json
      
      - name: Install dependencies
        run: |
          cd frontend
          npm ci
      
      - name: Run tests
        run: |
          cd frontend
          npm run test:coverage
          npm run test:e2e
      
      - name: Build application
        run: |
          cd frontend
          npm run build
        env:
          NEXT_PUBLIC_SUPABASE_URL: ${{ secrets.SUPABASE_URL }}
          NEXT_PUBLIC_SUPABASE_ANON_KEY: ${{ secrets.SUPABASE_ANON_KEY }}
      
      - name: Deploy to Vercel
        uses: amondnet/vercel-action@v25
        with:
          vercel-token: ${{ secrets.VERCEL_TOKEN }}
          vercel-org-id: ${{ secrets.VERCEL_ORG_ID }}
          vercel-project-id: ${{ secrets.VERCEL_PROJECT_ID }}
          working-directory: frontend
          vercel-args: '--prod'
```

#### Backend Deployment
```yaml
name: Deploy Backend

on:
  push:
    branches: [main]
    paths: ['backend/**']

jobs:
  deploy:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      
      - name: Setup Go
        uses: actions/setup-go@v4
        with:
          go-version: '1.21'
      
      - name: Run tests
        run: |
          cd backend
          go test -v ./...
      
      - name: Build Docker image
        run: |
          cd backend
          docker build -t workflow-backend .
      
      - name: Deploy to Railway
        run: |
          railway deploy
        env:
          RAILWAY_TOKEN: ${{ secrets.RAILWAY_TOKEN }}
```

## Monitoring & Observability

### Health Checks

#### Frontend Health Check
```typescript
// pages/api/health.ts
export default function handler(req: NextApiRequest, res: NextApiResponse) {
  const healthcheck = {
    uptime: process.uptime(),
    message: 'OK',
    timestamp: Date.now(),
    checks: {
      database: 'OK', // Add actual database check
      redis: 'OK',    // Add actual Redis check
    }
  }
  
  res.status(200).json(healthcheck)
}
```

#### Backend Health Check
```go
func healthCheck(c *gin.Context) {
    c.JSON(200, gin.H{
        "status":    "OK",
        "timestamp": time.Now().Unix(),
        "version":   os.Getenv("APP_VERSION"),
    })
}
```

### Logging

#### Structured Logging
```typescript
// utils/logger.ts
import pino from 'pino'

export const logger = pino({
  level: process.env.LOG_LEVEL || 'info',
  transport: {
    target: 'pino-pretty',
    options: {
      colorize: true,
      translateTime: true,
      ignore: 'pid,hostname',
    },
  },
})
```

### Metrics

#### Performance Monitoring
- **Frontend**: Vercel Analytics, Google Analytics 4
- **Backend**: Prometheus metrics, Grafana dashboards
- **Database**: PostgreSQL monitoring, query performance
- **Cache**: Redis monitoring, hit rates

## Security

### HTTPS Configuration
```nginx
# nginx.conf
server {
    listen 443 ssl http2;
    server_name yourdomain.com;
    
    ssl_certificate /path/to/cert.pem;
    ssl_certificate_key /path/to/private.key;
    
    ssl_protocols TLSv1.2 TLSv1.3;
    ssl_ciphers HIGH:!aNULL:!MD5;
    
    add_header Strict-Transport-Security "max-age=31536000; includeSubDomains" always;
    add_header X-Frame-Options DENY always;
    add_header X-Content-Type-Options nosniff always;
    
    location / {
        proxy_pass http://frontend:3000;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }
    
    location /api/ {
        proxy_pass http://backend:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }
}
```

### Database Security
```sql
-- Create application user with limited privileges
CREATE USER workflow_app WITH PASSWORD 'secure_password';
GRANT CONNECT ON DATABASE workflow TO workflow_app;
GRANT USAGE ON SCHEMA public TO workflow_app;
GRANT SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA public TO workflow_app;
GRANT USAGE, SELECT ON ALL SEQUENCES IN SCHEMA public TO workflow_app;
```

## Backup & Recovery

### Database Backup
```bash
#!/bin/bash
# backup-db.sh
DATE=$(date +%Y%m%d_%H%M%S)
pg_dump $DATABASE_URL | gzip > "backup_${DATE}.sql.gz"

# Upload to S3 (optional)
aws s3 cp "backup_${DATE}.sql.gz" s3://your-backup-bucket/db-backups/
```

### Automated Backups
```yaml
# .github/workflows/backup.yml
name: Database Backup

on:
  schedule:
    - cron: '0 2 * * *'  # Daily at 2 AM

jobs:
  backup:
    runs-on: ubuntu-latest
    steps:
      - name: Backup Database
        run: |
          pg_dump $DATABASE_URL | gzip > backup.sql.gz
        env:
          DATABASE_URL: ${{ secrets.DATABASE_URL }}
      
      - name: Upload to S3
        uses: aws-actions/configure-aws-credentials@v2
        with:
          aws-access-key-id: ${{ secrets.AWS_ACCESS_KEY_ID }}
          aws-secret-access-key: ${{ secrets.AWS_SECRET_ACCESS_KEY }}
          aws-region: us-west-2
```

## Troubleshooting

### Common Issues
1. **Build failures**: Check environment variables
2. **Database connection**: Verify connection strings
3. **CORS issues**: Configure allowed origins
4. **Performance**: Monitor Core Web Vitals
5. **SSL certificates**: Verify certificate chain

### Emergency Procedures
1. **Rollback**: Use git tags for quick rollbacks
2. **Scale down**: Reduce resource usage if needed
3. **Database recovery**: Restore from latest backup
4. **CDN purge**: Clear cache for immediate updates

## Resources

- [Next.js Deployment](https://nextjs.org/docs/deployment)
- [Vercel Documentation](https://vercel.com/docs)
- [Docker Best Practices](https://docs.docker.com/develop/dev-best-practices/)
- [AWS ECS Documentation](https://docs.aws.amazon.com/ecs/)
- [Railway Documentation](https://docs.railway.app/)
- [Fly.io Documentation](https://fly.io/docs/)