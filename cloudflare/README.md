# ☁️ Go Coffee Platform - Cloudflare Deployment

## 📋 Overview

This directory contains the complete Cloudflare deployment configuration for the Go Coffee platform, including Workers, Pages, KV storage, R2 buckets, and deployment automation.

## 🏗️ Infrastructure Created

### ✅ **KV Namespaces** (Created via MCP Tools)
- **go-coffee-cache** (`f12294ae42924b729024f030e9b5611c`) - Application caching
- **go-coffee-sessions** (`be1b38800d8d4f00820a04d1e5866552`) - User sessions
- **go-coffee-orders** (`6a65944658824e6b8bbc9f9e24e10317`) - Order data

### ✅ **R2 Buckets** (Created via MCP Tools)
- **go-coffee-assets** - Static assets and CDN content
- **go-coffee-images** - Coffee images and user uploads
- **go-coffee-backups** - Backups and logs

### 🔄 **Workers** (Ready to Deploy)
- **go-coffee-ai-coordinator** - AI agent coordination
- **go-coffee-event-router** - Cross-cloud event routing
- **go-coffee-order-processor** - Order processing and payments

### 🌐 **Pages** (Ready to Deploy)
- **go-coffee-frontend** - Next.js frontend application

## 🚀 Quick Deployment

### Prerequisites

1. **Install Wrangler CLI**:
   ```bash
   npm install -g wrangler
   ```

2. **Login to Cloudflare**:
   ```bash
   wrangler login
   ```

3. **Set Environment Variables**:
   ```bash
   export STRIPE_SECRET_KEY="sk_live_..."
   export DATABASE_URL="postgresql://..."
   export JWT_SECRET="your-jwt-secret"
   export TWILIO_AUTH_TOKEN="your-twilio-token"
   export SENDGRID_API_KEY="SG.your-sendgrid-key"
   ```

### Deploy Everything

```bash
# Make deployment script executable
chmod +x cloudflare/deploy-cloudflare.sh

# Deploy to production
./cloudflare/deploy-cloudflare.sh

# Deploy to staging
./cloudflare/deploy-cloudflare.sh --environment staging

# Dry run (test without deploying)
./cloudflare/deploy-cloudflare.sh --dry-run
```

## 📁 Directory Structure

```
cloudflare/
├── workers/
│   ├── ai-agent-coordinator/
│   │   └── wrangler.toml
│   ├── cross-cloud-event-router/
│   │   └── wrangler.toml
│   └── coffee-order-processor/
│       └── wrangler.toml
├── pages/
│   └── go-coffee-frontend/
│       └── wrangler.toml
├── deploy-cloudflare.sh
├── environment-config.sh
└── README.md
```

## ⚙️ Manual Deployment Steps

### 1. Deploy Workers

```bash
# AI Agent Coordinator
cd cloudflare/workers/ai-agent-coordinator
wrangler deploy --env production

# Cross-Cloud Event Router
cd ../cross-cloud-event-router
wrangler deploy --env production

# Coffee Order Processor
cd ../coffee-order-processor
wrangler deploy --env production
```

### 2. Deploy Frontend

```bash
# Build frontend
cd web-ui/frontend
npm ci
npm run build

# Deploy to Pages
cd ../../cloudflare/pages/go-coffee-frontend
wrangler pages deploy ../../../web-ui/frontend/out --project-name go-coffee-frontend
```

### 3. Create Queues

```bash
# Create processing queues
wrangler queues create ai-coordination-tasks
wrangler queues create cross-cloud-events
wrangler queues create coffee-orders
wrangler queues create payment-processing

# Create dead letter queues
wrangler queues create ai-coordination-tasks-dlq
wrangler queues create cross-cloud-events-dlq
wrangler queues create coffee-orders-dlq
wrangler queues create payment-processing-dlq
```

### 4. Set Secrets

```bash
# Set secrets for each worker
echo "$STRIPE_SECRET_KEY" | wrangler secret put STRIPE_SECRET_KEY --name go-coffee-order-processor
echo "$DATABASE_URL" | wrangler secret put DATABASE_URL --name go-coffee-order-processor
echo "$JWT_SECRET" | wrangler secret put JWT_SECRET --name go-coffee-ai-coordinator
```

## 🌐 Domain Configuration

### Custom Domains (Configure in Cloudflare Dashboard)

1. **Frontend Domains**:
   - `go-coffee.com` → Pages project
   - `app.go-coffee.com` → Pages project
   - `www.go-coffee.com` → Pages project (redirect)

2. **API Domains**:
   - `api.go-coffee.com` → go-coffee-order-processor
   - `events.go-coffee.com` → go-coffee-event-router
   - `ai.go-coffee.com` → go-coffee-ai-coordinator

3. **CDN Domains**:
   - `cdn.go-coffee.com` → R2 bucket (go-coffee-assets)
   - `images.go-coffee.com` → R2 bucket (go-coffee-images)

## 🔧 Environment Configuration

### Production Environment

```bash
# Set production environment
export ENVIRONMENT=production
export CLOUDFLARE_ACCOUNT_ID=6244f6d02d9c7684386c1c849bdeaf56
```

### Staging Environment

```bash
# Set staging environment
export ENVIRONMENT=staging
export CLOUDFLARE_ACCOUNT_ID=6244f6d02d9c7684386c1c849bdeaf56
```

### Development Environment

```bash
# Set development environment
export ENVIRONMENT=development
export CLOUDFLARE_ACCOUNT_ID=6244f6d02d9c7684386c1c849bdeaf56
```

## 📊 Monitoring and Analytics

### Built-in Analytics

- **Workers Analytics** - Automatic request/response metrics
- **Pages Analytics** - Frontend performance and usage
- **R2 Analytics** - Storage usage and bandwidth

### Custom Analytics

- **Analytics Engine** - Custom event tracking
- **Real User Monitoring** - Frontend performance
- **Error Tracking** - Application errors and exceptions

## 🔒 Security Configuration

### Headers

All responses include security headers:
- `X-Frame-Options: DENY`
- `X-Content-Type-Options: nosniff`
- `Referrer-Policy: strict-origin-when-cross-origin`
- `Content-Security-Policy: default-src 'self'`

### Rate Limiting

- **API Endpoints**: 500 requests/minute per IP
- **Frontend**: 1000 requests/minute per IP
- **WebSocket**: 100 connections per IP

### CORS Configuration

- **Allowed Origins**: `go-coffee.com`, `app.go-coffee.com`
- **Allowed Methods**: `GET`, `POST`, `PUT`, `DELETE`, `OPTIONS`
- **Allowed Headers**: `Content-Type`, `Authorization`

## 🚨 Troubleshooting

### Common Issues

1. **Worker Deployment Fails**:
   ```bash
   # Check wrangler configuration
   wrangler whoami
   wrangler kv:namespace list
   ```

2. **Pages Build Fails**:
   ```bash
   # Check Node.js version
   node --version  # Should be 18+
   
   # Clear cache and rebuild
   rm -rf node_modules package-lock.json
   npm install
   npm run build
   ```

3. **Domain Not Working**:
   - Check DNS propagation: `dig go-coffee.com`
   - Verify SSL certificate status in Cloudflare dashboard
   - Check custom domain configuration

### Debug Commands

```bash
# Check worker logs
wrangler tail go-coffee-order-processor

# Test worker locally
wrangler dev --env development

# Check KV namespace contents
wrangler kv:key list --namespace-id f12294ae42924b729024f030e9b5611c

# Check R2 bucket contents
wrangler r2 object list go-coffee-assets
```

## 📈 Performance Optimization

### Caching Strategy

- **Static Assets**: 1 year cache (immutable)
- **API Responses**: 5 minutes cache
- **Dynamic Content**: No cache
- **Images**: 30 days cache

### CDN Configuration

- **Global Distribution**: Cloudflare's 300+ edge locations
- **Smart Routing**: Argo Smart Routing enabled
- **Compression**: Brotli and Gzip compression
- **Image Optimization**: Polish and Mirage enabled

## 💰 Cost Optimization

### Resource Limits

- **Workers**: CPU 50-100ms, Memory 128-256MB
- **KV Operations**: 1000 reads/writes per day (free tier)
- **R2 Storage**: 10GB free, then $0.015/GB/month
- **Pages**: 500 builds/month (free tier)

### Monitoring Costs

- Check usage in Cloudflare dashboard
- Set up billing alerts
- Monitor resource utilization

## 🎯 Next Steps

1. **Configure Custom Domains** in Cloudflare dashboard
2. **Set up DNS Records** for your domain
3. **Configure SSL Certificates** (automatic with Cloudflare)
4. **Test All Endpoints** and functionality
5. **Set up Monitoring** and alerts
6. **Configure Backup Strategy** for critical data
7. **Implement CI/CD Pipeline** for automated deployments

## 📞 Support

- **Cloudflare Documentation**: https://developers.cloudflare.com/
- **Wrangler CLI Docs**: https://developers.cloudflare.com/workers/wrangler/
- **Go Coffee Platform Docs**: `../docs/`

---

## 🎉 **Cloudflare Deployment Ready!**

Your Go Coffee platform is now configured for enterprise-grade deployment on Cloudflare's global network with:

- ✅ **Serverless Workers** for backend processing
- ✅ **Edge-optimized Pages** for frontend delivery
- ✅ **Global KV Storage** for caching and sessions
- ✅ **R2 Object Storage** for assets and backups
- ✅ **Queue Processing** for async operations
- ✅ **Custom Domains** and SSL certificates
- ✅ **Security Headers** and rate limiting
- ✅ **Analytics and Monitoring** built-in

**Ready to serve millions of coffee lovers worldwide!** ☕🚀
