# Configuration Management

This document describes the configuration management system for the Medical Rep API.

## Overview

The configuration system uses a layered approach with the following priority order (highest to lowest):

1. **Environment Variables** (highest priority)
2. **Environment-specific Config Files** (e.g., `config.dev.yaml`, `config.prod.yaml`)
3. **Base Config File** (`config.yaml`)
4. **Default Values** (lowest priority)

## Directory Structure

```
configs/
├── config.go              # Configuration structs and loading logic
├── config.yaml            # Base configuration file
├── config.dev.yaml        # Development environment overrides
├── config.prod.yaml       # Production environment overrides
├── config.staging.yaml    # Staging environment overrides (optional)
└── README.md              # This file

.env                        # Local environment variables (gitignored)
.env.example               # Example environment variables file
```

## Configuration Files

### Base Configuration (`config.yaml`)

Contains default settings that apply to all environments. Should include:
- Application metadata (name, version)
- Default server settings
- Default database configuration
- Default logging configuration

### Environment-Specific Configuration

Environment-specific files override base configuration:

- `config.dev.yaml` - Development environment
- `config.prod.yaml` - Production environment
- `config.staging.yaml` - Staging environment (optional)

These files should only contain settings that differ from the base configuration.

### Environment Variables

Environment variables follow the naming convention:
`MEDICAL_REP_<SECTION>_<SUBSECTION>_<KEY>`

Examples:
- `MEDICAL_REP_APP_NAME`
- `MEDICAL_REP_HTTP_PORT`
- `MEDICAL_REP_DATABASE_HOST`
- `MEDICAL_REP_AUTH_JWT_SECRET`

## Configuration Sections

### Application (`app`)
- `name`: Application name
- `version`: Application version
- `environment`: Runtime environment (development, staging, production)
- `debug`: Debug mode flag
- `shutdown.timeout`: Graceful shutdown timeout

### HTTP Server (`http`)
- `port`: Server port
- `host`: Server host/interface
- `read_timeout`: Request read timeout
- `write_timeout`: Response write timeout
- `idle_timeout`: Connection idle timeout
- `max_header_bytes`: Maximum header size
- `tls`: TLS configuration
- `cors`: CORS configuration
- `rate_limit`: Rate limiting configuration

### Database (`database`)
- `driver`: Database driver (postgres, mysql)
- `host`: Database host
- `port`: Database port
- `database`: Database name
- `username`: Database username
- `password`: Database password
- `ssl_mode`: SSL mode for connections
- `max_open_conns`: Maximum open connections
- `max_idle_conns`: Maximum idle connections
- `conn_max_lifetime`: Connection maximum lifetime
- `migrations_path`: Database migrations path

### Redis (`redis`)
- `host`: Redis host
- `port`: Redis port
- `password`: Redis password
- `database`: Redis database number
- `pool_size`: Connection pool size
- `dial_timeout`: Connection dial timeout
- `read_timeout`: Read operation timeout
- `write_timeout`: Write operation timeout

### Authentication (`auth`)
- `jwt_secret`: JWT signing secret
- `jwt_expiration`: JWT token expiration time
- `bcrypt_cost`: Bcrypt hashing cost

### Logging (`logging`)
- `level`: Log level (debug, info, warn, error)
- `format`: Log format (json, console)
- `output`: Log output (stdout, stderr, or file path)
- `max_size`: Log file max size in MB
- `max_backups`: Number of backup files to keep
- `max_age`: Max age of log files in days
- `compress`: Whether to compress rotated files

### Health Checks (`health`)
- `enabled`: Enable health checks
- `check_interval`: Health check interval
- `timeout`: Health check timeout
- `database_check`: Enable database health check
- `redis_check`: Enable Redis health check
- `external_checks`: List of external URLs to check

## Usage

### Loading Configuration

```go
import "medical-rep/configs"

func main() {
    // Load configuration
    if err := configs.Load(); err != nil {
        log.Fatal("Failed to load config:", err)
    }
    
    // Get configuration instance
    cfg := configs.Get()
    
    // Use configuration
    fmt.Printf("Server will start on port %d\n", cfg.HTTP.Port)
}
```

### Environment-Specific Setup

1. **Development**:
   - Copy `.env.example` to `.env`
   - Modify values as needed
   - Uses `config.dev.yaml` automatically

2. **Production**:
   - Set `MEDICAL_REP_APP_ENVIRONMENT=production`
   - Configure production-specific environment variables
   - Uses `config.prod.yaml` automatically

3. **Docker**:
   ```dockerfile
   ENV MEDICAL_REP_APP_ENVIRONMENT=production
   ENV MEDICAL_REP_HTTP_PORT=8080
   ENV MEDICAL_REP_DATABASE_HOST=db
   ```

4. **Kubernetes**:
   ```yaml
   env:
   - name: MEDICAL_REP_APP_ENVIRONMENT
     value: "production"
   - name: MEDICAL_REP_DATABASE_PASSWORD
     valueFrom:
       secretKeyRef:
         name: db-secret
         key: password
   ```

## Best Practices

### Security
- Never commit `.env` files to version control
- Use environment variables for sensitive data (passwords, secrets)
- Use different JWT secrets for each environment
- Enable TLS in production
- Use strong bcrypt costs in production

### Environment Management
- Keep environment-specific files minimal
- Override only what's necessary for each environment
- Use descriptive environment variable names
- Document any custom environment variables

### Configuration Validation
- Required fields are validated at startup
- Invalid configurations cause application startup failure
- Use appropriate data types and validation rules

### Performance
- Use connection pooling for databases
- Set appropriate timeouts
- Configure rate limiting in production
- Use caching where appropriate

## Example Configurations

### Development Environment Variables
```bash
MEDICAL_REP_APP_ENVIRONMENT=development
MEDICAL_REP_APP_DEBUG=true
MEDICAL_REP_HTTP_PORT=8080
MEDICAL_REP_DATABASE_HOST=localhost
MEDICAL_REP_DATABASE_PASSWORD=postgres
MEDICAL_REP_AUTH_JWT_SECRET=dev-secret
MEDICAL_REP_LOGGING_LEVEL=debug
```

### Production Environment Variables
```bash
MEDICAL_REP_APP_ENVIRONMENT=production
MEDICAL_REP_APP_DEBUG=false
MEDICAL_REP_HTTP_PORT=8080
MEDICAL_REP_HTTP_TLS_ENABLED=true
MEDICAL_REP_DATABASE_HOST=prod-db-host
MEDICAL_REP_DATABASE_PASSWORD=secure-password
MEDICAL_REP_AUTH_JWT_SECRET=super-secure-secret
MEDICAL_REP_LOGGING_LEVEL=warn
```

## Troubleshooting

### Common Issues

1. **Configuration not loading**:
   - Check file paths are correct
   - Verify YAML syntax
   - Check file permissions

2. **Environment variables not working**:
   - Verify naming convention: `MEDICAL_REP_<SECTION>_<KEY>`
   - Check for typos in variable names
   - Ensure variables are exported

3. **Database connection issues**:
   - Verify database credentials
   - Check network connectivity
   - Validate SSL mode settings

4. **Invalid configuration values**:
   - Check data types match expected values
   - Verify required fields are set
   - Review validation error messages

### Debugging Configuration

Enable debug logging to see configuration loading process:
```bash
MEDICAL_REP_LOGGING_LEVEL=debug ./crmserver
```

## Migration Guide

When updating configuration structure:

1. Update structs in `config.go`
2. Add new fields to base `config.yaml`
3. Update environment-specific overrides
4. Update `.env.example`
5. Update this README
6. Add validation for new required fields
7. Test with different environments