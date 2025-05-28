# 🏥 Go Medical Rep CRM

[![Go Report Card](https://goreportcard.com/badge/github.com/rixtrayker/medical-rep)](https://goreportcard.com/report/github.com/rixtrayker/medical-rep)
[![GoDoc](https://godoc.org/github.com/rixtrayker/medical-rep?status.svg)](https://godoc.org/github.com/rixtrayker/medical-rep)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?logo=go)](https://golang.org)
[![PostgreSQL](https://img.shields.io/badge/PostgreSQL-15+-336791?logo=postgresql)](https://www.postgresql.org)

> An advanced, modular, and scalable Customer Relationship Management (CRM) system meticulously tailored for the dynamic needs of Medical Representatives, built with the power and efficiency of Go.

## 🌟 Overview

This project is engineered to provide a truly robust platform for managing sophisticated client interactions, from initial contact to ongoing engagement. It handles intricate visit planning, streamlined order processing, and insightful reporting. The system prioritizes:

- ⚡ Exceptional performance for demanding workloads
- 🛠️ Enhanced maintainability through clean code
- 🔄 Clear boundaries for better scalability
- 🚀 Future extensibility towards SaaS model

## ✨ Key Features

### 🏗️ Modular Architecture
- Clear separation of concerns
- Independent module development
- Scalable component design
- Future-proof architecture

### 👥 User Management
- 🔐 Role-based access control (RBAC)
- 🔑 Secure authentication
- 📝 Detailed permission settings
- 🔒 Future: 2FA & audit trails

### 🗺️ Territory Management
- Hierarchical geographical structure
- Performance tracking
- Workload distribution
- Sales strategy execution

### 👨‍⚕️ Client Management
- 360-degree HCP profiles
- Detailed demographics
- Interaction history
- Communication preferences
- Client segmentation

### 📅 Visit Planning & Logging
- Optimized routing
- Compliance reporting
- Field intelligence capture
- Visit feedback tracking

### 📦 Product & Order Management
- Detailed product catalog
- Inventory integration
- Order processing
- Pricing tier management
- Status tracking

### 📊 Activity & Expense Tracking
- Non-visit activity logging
- Expense report management
- Approval workflows
- Budget control
- Operational insights

### 📈 Reporting & Analytics (Planned)
- KPI tracking
- Trend analysis
- Comparative dashboards
- Basic forecasting

### 🔔 Notifications & Messaging (Planned)
- Real-time alerts
- Visit reminders
- Approval requests
- Team coordination
- Internal messaging

## 🛠️ Technology Stack

| Category | Technology | Description |
|----------|------------|-------------|
| Backend | Go (Golang) | High performance, concurrency, standard library |
| Database | PostgreSQL | ACID compliance, JSONB support, extensibility |
| API | GraphQL (gqlgen) | Flexible data fetching, type safety |
| Auth | JWT | Stateless authentication |
| AuthZ | Casbin | Flexible access control models |
| Web Framework | Gin/Echo/Chi | Lightweight, efficient routing |
| ORM/DB | GORM/Ent/sqlx | Database interaction |
| Messaging | NATS/RabbitMQ | Async communication |

## 📂 Project Structure

```
.
├── api/              # API definitions (GraphQL, OpenAPI)
├── cmd/              # Application entry points
├── configs/          # Configuration files
├── internal/         # Private application code
│   ├── app/         # Application services
│   ├── domain/      # Business logic
│   ├── handler/     # HTTP/GraphQL handlers
│   ├── infra/       # Infrastructure implementations
│   └── platform/    # Shared utilities
├── pkg/             # Public library code
├── migrations/      # Database migrations
├── graph/           # GraphQL generated code
└── scripts/         # Helper scripts
```

## 🚀 Getting Started

### Prerequisites

- Go 1.21+
- PostgreSQL 15+
- Docker & Docker Compose
- golang-migrate CLI
- make (optional)

### Quick Start

```bash
# Clone the repository
git clone https://github.com/rixtrayker/medical-rep.git

# Navigate to project directory
cd medical-rep

# Start development environment
make dev

# Run migrations
make migrate-up

# Start the server
make run
```

### Development

```bash
# Run tests
make test

# Build binary
make build

# Run linter
make lint
```

## 📚 Documentation

- [API Documentation](docs/api.md)
- [Database Schema](docs/schema.md)
- [Development Guide](docs/development.md)
- [Deployment Guide](docs/deployment.md)

## 🤝 Contributing

We welcome contributions! Please see our [Contributing Guide](CONTRIBUTING.md) for details.

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

- [Go](https://golang.org)
- [PostgreSQL](https://www.postgresql.org)
- [gqlgen](https://github.com/99designs/gqlgen)
- [Casbin](https://casbin.org)
- And all other open-source projects that made this possible!

---

<div align="center">
Made with ❤️ by Your Team Name
</div>