# Go Medical Rep CRM: Technical Documentation

## Table of Contents
1. [Introduction](#introduction)
2. [System Architecture](#system-architecture)
3. [Core Modules](#core-modules)
4. [Database Schema](#database-schema)
5. [API Design](#api-design)
6. [Deployment & Infrastructure](#deployment--infrastructure)
7. [Development Guidelines](#development-guidelines)
8. [Troubleshooting](#troubleshooting)
9. [Future Considerations](#future-considerations)

## Introduction
This document provides a technical overview of the Go Medical Rep CRM system, focusing on its modular architecture and the underlying PostgreSQL database schema. It is intended for developers, database administrators, and architects involved in the development, maintenance, and future evolution of the system.

### Key Features
- Modular monolith architecture
- PostgreSQL database with advanced features
- Role-based access control (RBAC)
- RESTful API design
- Event-driven communication
- Real-time notifications
- Geographic territory management

### Technology Stack
- Backend: Go 1.21+
- Database: PostgreSQL 15+
- Cache: Redis
- Message Queue: RabbitMQ
- API Documentation: Swagger/OpenAPI
- Testing: Go testing framework, testify
- CI/CD: GitHub Actions

## System Architecture

### High-Level Overview
The system follows a modular monolith architecture, which provides:
- Single deployable unit
- Clear separation of concerns
- Independent module development
- Simplified testing and maintenance
- Path to microservices if needed

### Architecture Diagram
```mermaid
graph TD
    A[API Layer] --> B[Service Layer]
    B --> C[Domain Layer]
    C --> D[Repository Layer]
    D --> E[Database]
    B --> F[Event Bus]
    F --> G[Event Handlers]
```

### Key Design Principles
1. **Domain-Driven Design (DDD)**
   - Clear bounded contexts
   - Rich domain models
   - Ubiquitous language

2. **SOLID Principles**
   - Single Responsibility
   - Open/Closed
   - Liskov Substitution
   - Interface Segregation
   - Dependency Inversion

3. **Clean Architecture**
   - Independent of frameworks
   - Testable
   - Independent of UI
   - Independent of database
   - Independent of external agencies

## Core Modules
The CRM is conceptually divided into several core modules, each responsible for a distinct set of functionalities. These modules are designed to be highly cohesive internally and loosely coupled with other modules, primarily interacting through well-defined interfaces or, where appropriate, an event-driven mechanism.

UserAuth & Permissions:

Manages all aspects of user identity, including secure creation of user accounts (medical reps, managers, admins), robust authentication mechanisms (login/logout, session management), and fine-grained authorization through a system of roles and permissions. This module serves as the gatekeeper for accessing system functionalities.

Key Tables: users, roles, permissions, user_roles, role_permissions.

Territory Management:

Handles the complex geographical hierarchy (e.g., Country > Region > Governorate > City > Area > Brick) used for organizing sales territories and assigning them to medical reps. This module is fundamental for segmenting client bases and distributing workload.

Key Tables: countries, regions, governorates, cities, areas, bricks, user_territory_assignments.

Client Management:

Manages comprehensive information about Healthcare Professionals (HCPs), clinics, hospitals, pharmacies, and other client entities. This includes demographic data, professional details (like specialty), classification (grading), and interaction history. This module is the heart of any CRM, providing the "C" in Customer Relationship Management.

Key Tables: client_types, specialties, clients.

Visit Planning & Logging:

Provides core functionality for medical representatives to plan their client visits (e.g., weekly or monthly schedules), and subsequently log the details and outcomes of these visits. This includes capturing visit objectives, discussion points, feedback, and follow-up actions. This module is critical for tracking rep activity and effectiveness.

Key Tables: plans, plan_days, plan_day_clients, call_types, visits.

Product Catalog & Feedback:

Manages detailed information about medical products, including categorization, descriptions, and potentially pricing or promotional materials. It also tracks specific product-related feedback received from clients during visits, which is invaluable for marketing and product development.

Key Tables: product_categories, products, visit_products_feedback.

Activity & Expense Management:

Allows reps to log various work-related activities beyond direct client visits (e.g., office work, training sessions, internal meetings) and to submit expense reports for reimbursement, complete with approval workflows. This provides a fuller picture of a rep's time utilization and operational costs.

Key Tables: activities, expenses.

Order Management:

Facilitates the creation, tracking, and management of product orders placed by clients or submitted by reps on behalf of clients. This includes order status updates, item details, and potentially pricing and discount application.

Key Tables: orders, order_items.

Communication (Notifications & Messaging):

Handles system-generated notifications to users (e.g., reminders, alerts for approvals) and provides internal messaging capabilities for communication between team members or between reps and managers.

Key Tables: notifications, messages.

System Settings:

Manages global application settings and configurations that control various aspects of the CRM's behavior, such as feature flags, default values, or integration parameters.

Key Table: settings.

## Database Schema
The database schema is meticulously designed to robustly support the functionalities of the core modules. It utilizes PostgreSQL and leverages its advanced features like UUIDs for primary keys, custom ENUM types for ensuring controlled vocabularies, JSONB for flexible data storage, and TIMESTAMPTZ for accurate, timezone-aware time tracking.

### General Principles
Primary Keys: Most tables employ UUID (Universally Unique Identifiers) as primary keys, generated via DEFAULT uuid_generate_v4(). This strategy ensures global uniqueness, which is highly beneficial for preventing key collisions in distributed environments, simplifying data merging from different sources, and making record identifiers less predictable than sequential integers.

Timestamps: Standard created_at and updated_at columns (both TIMESTAMPTZ) are present in most tables, recording the creation and last modification times. The use of TIMESTAMPTZ (timestamp with time zone) is critical for applications that might operate across different geographical regions or need to accurately represent historical data irrespective of server or client time zone settings. updated_at is automatically managed via a database trigger (update_updated_at_column) to ensure it always reflects the latest change.

Soft Deletes: A deleted_at TIMESTAMPTZ column is consistently used for implementing soft deletes. This allows data to be marked as deleted (and typically hidden from regular application queries) without physical removal from the database. This approach is valuable for data recovery, auditing, and maintaining historical integrity, though it requires careful query construction to exclude "deleted" records.

ENUM Types: Custom PostgreSQL ENUM types are defined for fields that have a fixed, predefined set of possible values (e.g., plan_status_enum, visit_status_enum). This enforces data integrity at the database level, preventing invalid values, and can simplify application-level validation logic, making the data more self-describing and consistent.

Indexing: Appropriate indexes are created not only on foreign keys (which is standard practice) but also on other frequently queried columns, particularly those used in WHERE clauses, JOIN conditions, or for ORDER BY operations. This is crucial for optimizing query performance and ensuring the application remains responsive as data volume grows. Indexing strategies will be continuously reviewed based on query patterns.

### Key Tables and Relationships (Grouped by Conceptual Module)
(Refer to the crm_postgres_schema_v1.sql artifact in the Canvas for the complete DDL and specific column definitions.)

UserAuth & Permissions Module
users: Stores essential user account information, including credentials (hashed passwords) and profile details. The last_login_at field is useful for tracking user activity and identifying potentially dormant accounts.

roles: Defines distinct user roles within the system (e.g., 'medical-rep', 'district-manager', 'admin'), forming the basis of role-based access control (RBAC).

permissions: Defines granular operational permissions (e.g., 'client:create', 'report:view_all'), allowing for fine-tuned control over what actions users can perform.

user_roles (Junction): A many-to-many mapping table that assigns roles to users. A user can have multiple roles.

role_permissions (Junction): A many-to-many mapping table that assigns permissions to roles. This defines the set of capabilities for each role.

Territory Module
countries, regions, governorates, cities, areas, bricks: These tables collectively represent a deeply hierarchical geographical structure. Each level references its parent (e.g., regions has a country_id). The areas and bricks tables can store polygon_data (using JSONB) for geographical mapping and visualization purposes. Querying across this hierarchy (e.g., finding all clients in a region) might involve recursive queries or careful joining strategies.

user_territory_assignments: Links users (typically medical reps) to specific territories at various levels (country, region, city, etc.). The assigned_territory_id is a generic UUID; its specific meaning (i.e., which territory table it refers to) is determined by the territory_level_enum column. This requires careful application logic to resolve and maintain consistency.

Client Management Module
client_types: Defines categories of clients, such as 'Doctor', 'Pharmacy', 'Hospital', allowing for tailored interactions and reporting.

specialties: Stores medical specialties, primarily for HCP clients, enabling targeted marketing and visit planning.

clients: The central table for all client information. It links to client_types, specialties, and bricks (for geographical assignment/location). It contains contact details, classification (e.g., grade A, B, C), and other relevant CRM data.

Visit Planning & Logging Module
plans: Represents high-level visit plans for medical reps, often structured on a monthly or cyclical basis, outlining objectives and target clients.

plan_days: Breaks down a plan into individual daily schedules, potentially with distinct AM/PM shifts, allowing for detailed day-to-day organization.

plan_day_clients: The crucial link between planning and execution. It associates specific clients with a plan_day, including a planned sequence for visits and any pre-visit notes or objectives.

call_types: Defines different types or purposes of visits/calls (e.g., 'Standard Detail', 'Follow-up', 'Sample Drop'), allowing for better categorization and analysis of interactions.

visits: Records the actual client visits that take place. This table links to users (the rep), clients, plan_day_clients (to compare planned vs. actual), and call_types. It captures vital data like actual start/end times, GPS location data for check-in/out (if applicable), and detailed feedback or outcomes of the visit.

Product & Feedback Module
product_categories: Allows for a hierarchical categorization of medical products, aiding in organization and targeted promotion.

products: Contains detailed information about each medical product, such as name, SKU, description, and links to its category.

visit_products_feedback: A vital table for capturing product-specific intelligence. It records feedback, physician interest levels, and quantities of samples dropped for each product discussed during a visit.

Activity & Expense Module
activities: Logs non-visit activities undertaken by users, such as office work, training sessions, or internal team meetings, providing a more complete view of time allocation.

expenses: Tracks expenses submitted by users for reimbursement. It includes details like expense type, amount, date, associated visit (if any), status (pending, approved, rejected), and approval history.

Order Management Module
orders: Contains header information for client orders, including the client, the rep who placed it, order date, overall status, total amounts, and shipping/billing details.

order_items: Details each line item within an order, linking to a specific product and recording quantity, price at the time of order, and subtotal.

Communication Module
notifications: Stores system-generated notifications intended for users, such as reminders for planned visits, alerts for plan approvals, or updates on order statuses.

messages: Supports internal user-to-user messaging, potentially including basic threading (via parent_message_id) for conversational context.

System Settings Module
settings: A flexible key-value store for managing application-wide configurations and parameters. The value_type column guides the application on how to interpret the value (stored as text).

### Data Integrity and Relationships
Foreign Keys: Referential integrity is strictly enforced using foreign key constraints across all related tables. This is fundamental to preventing orphaned records and ensuring that relationships between entities (e.g., a visit always belongs to a valid client and user) are maintained.

ON DELETE Behavior: The behavior upon deletion of a parent record is carefully defined:

ON DELETE CASCADE: Used when child records are intrinsically dependent on the parent and have no meaning or utility if the parent is removed. For instance, plan_days are cascaded if their parent plan is deleted, as daily schedules are meaningless without the overarching plan. Similarly, user_roles entries are removed if either the user or the role is deleted.

ON DELETE RESTRICT (or the default NO ACTION): Employed where child records should prevent the deletion of a parent record if they exist, safeguarding important related data. For example, a client cannot be deleted if they have associated orders or visits, ensuring historical transaction data isn't inadvertently lost.

ON DELETE SET NULL: Applied in cases where the relationship is optional. If the referenced parent record is deleted, the foreign key in the child record is set to NULL. An example is clients.specialty_id; if a specialty is removed from the system, existing clients previously assigned to it will have their specialty_id nulled rather than being deleted or preventing the specialty's deletion.

Triggers: An update_updated_at_column() trigger is implemented to automatically refresh the updated_at timestamp on any row modification for tables that include this column. While some prefer handling such audit fields in the application layer, for a universally required field like updated_at, a database trigger offers simplicity and guarantees consistency regardless of how the data modification occurs.

### Important Considerations
user_territory_assignments.assigned_territory_id: This field presents a design challenge akin to a polymorphic association in an RDBMS. Since assigned_territory_id can refer to an ID from one of several distinct territory tables (countries, regions, etc.), a direct foreign key constraint to all of them is not feasible. The territory_level enum column is crucial here, as the application logic must use its value to determine which specific territory table the assigned_territory_id corresponds to. This requires diligent implementation in the service layer to ensure data consistency and resolve these relationships correctly.

settings.value: Storing all setting values as TEXT provides flexibility but shifts the responsibility of type interpretation and validation to the application layer. The value_type column serves as a guide for the application to correctly parse, cast, and validate the setting's value before use, preventing runtime errors due to type mismatches.

JSONB Usage: The polygon_data field in areas and bricks utilizes PostgreSQL's JSONB type. This offers significant benefits, such as storing complex, schemaless (within the JSON structure itself) geographical coordinate data (e.g., GeoJSON). PostgreSQL also provides powerful functions for querying and indexing JSONB content. However, it's important to note that while the overall structure is validated as JSON, the consistency and schema within the JSON object (e.g., ensuring correct GeoJSON format) become the responsibility of the application logic during serialization and deserialization.

## API Design

### RESTful Endpoints
The system exposes RESTful APIs following these conventions:
- Resource-based URLs
- HTTP methods for operations
- JSON request/response format
- Proper status codes
- Pagination support
- Filtering and sorting

Example API structure:
```go
// User Management
GET    /api/v1/users
POST   /api/v1/users
GET    /api/v1/users/{id}
PUT    /api/v1/users/{id}
DELETE /api/v1/users/{id}

// Territory Management
GET    /api/v1/territories
POST   /api/v1/territories
GET    /api/v1/territories/{id}/clients
```

### Authentication & Authorization
- JWT-based authentication
- Role-based access control
- API key management
- Rate limiting
- Request validation

## Deployment & Infrastructure

### Requirements
- Go 1.21 or higher
- PostgreSQL 15 or higher
- Redis 6 or higher
- RabbitMQ 3.9 or higher
- Docker & Docker Compose

### Environment Variables
```env
DB_HOST=localhost
DB_PORT=5432
DB_NAME=medical_rep
DB_USER=postgres
DB_PASSWORD=secret

REDIS_URL=redis://localhost:6379
RABBITMQ_URL=amqp://guest:guest@localhost:5672/

JWT_SECRET=your-secret-key
API_PORT=8080
```

### Docker Deployment
```yaml
version: '3.8'
services:
  app:
    build: .
    ports:
      - "8080:8080"
    environment:
      - DB_HOST=db
      - REDIS_URL=redis://redis:6379
    depends_on:
      - db
      - redis
      - rabbitmq

  db:
    image: postgres:15
    environment:
      - POSTGRES_DB=medical_rep
      - POSTGRES_USER=postgres
      - POSTGRES_PASSWORD=secret

  redis:
    image: redis:6

  rabbitmq:
    image: rabbitmq:3.9-management
```

## Development Guidelines

### Code Style
- Follow Go standard formatting
- Use meaningful variable names
- Write comprehensive comments
- Keep functions small and focused
- Use interfaces for abstraction

### Testing
- Unit tests for all packages
- Integration tests for critical paths
- End-to-end tests for key workflows
- Mock external dependencies
- Maintain high test coverage

### Git Workflow
1. Feature branches from main
2. Pull request reviews
3. Automated testing
4. Code quality checks
5. Merge to main

## Troubleshooting

### Common Issues
1. Database Connection
   - Check connection string
   - Verify network access
   - Confirm credentials

2. Authentication
   - Validate JWT token
   - Check role permissions
   - Verify API keys

3. Performance
   - Monitor query execution
   - Check index usage
   - Review connection pools

### Logging
- Structured logging with levels
- Request tracing
- Error tracking
- Performance metrics

## Future Considerations
Data Migration: A detailed and thoroughly tested data migration strategy will be paramount for transitioning data from the existing PHP/Filament CRM to this new PostgreSQL schema. This will likely involve scripting for extraction, transformation (to align with the new schema and data types), and loading (ETL), along with comprehensive validation checks post-migration.

Schema Evolution: As the application matures and new requirements emerge, the database schema will inevitably need to evolve. Changes will be managed systematically using a database migration tool like golang-migrate/migrate. This ensures that schema modifications are version-controlled, repeatable, and can be applied consistently across different environments (development, staging, production), minimizing risks associated with manual changes. Thought must be given to backward compatibility or planning for controlled downtime if breaking schema changes are unavoidable.

Performance Tuning: Ongoing monitoring of database performance will be essential, especially as data volumes increase. This includes analyzing query execution plans for critical or slow queries, optimizing indexes (adding new ones, removing unused ones, or modifying existing ones), and potentially refactoring queries or data structures if bottlenecks are identified. Regular database maintenance tasks will also be scheduled.

SaaS Multi-tenancy: If the long-term vision includes evolving the CRM into a Software-as-a-Service (SaaS) offering, architectural considerations for multi-tenancy must be addressed. While the current UUID-based design and modular structure provide a good foundation, specific strategies for data isolation will be needed. Common approaches include a shared database with a tenant_id discriminator column in most tables (often preferred for modular monoliths due to simpler management) or, less commonly for this architecture, entirely separate databases per tenant. Each approach has trade-offs in terms of data isolation, cost, complexity, and regulatory compliance.

This documentation provides a snapshot of the system's technical design as of its current iteration. It is a living document and will be updated to reflect significant changes and advancements as the project progresses.