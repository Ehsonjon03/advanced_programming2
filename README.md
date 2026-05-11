## System Architecture (Assignment 4)

This diagram illustrates the current architecture including gRPC communication, Redis caching (Cache-aside pattern), and the asynchronous event-driven flow from Assignment 3.

```mermaid
graph TD
    Client[Postman / Browser] -->|1. HTTP Request| OrderService[Order Service :8081]
    
    subgraph "Order Service (Go)"
        Handler[HTTP Handler] --> UC[Order UseCase]
        UC --> Repo[Order Repository]
    end

    subgraph "Infrastructure"
        DB[(PostgreSQL)]
        Redis[(Redis Cache)]
        Broker[RabbitMQ Broker]
    end

    %% Cache-aside Logic
    Repo -.->|2. Check Cache| Redis
    Repo -.->|3. If Missing: Read/Write| DB
    Repo -.->|4. Update Cache| Redis

    %% gRPC Communication
    UC -->|5. gRPC: ProcessPayment| PaymentService[Payment Service :50051]
    PaymentService --- DB
    
    %% Asynchronous Flow (EDA)
    PaymentService -->|6. Publish: payment.completed| Broker
    Broker -->|7. Consume| NotificationService[Notification Service]
    NotificationService -->|8. Log| Console[Console: Email Simulated]

    style Redis fill:#f9f,stroke:#333,stroke-width:2px
    style DB fill:#69f,stroke:#333,stroke-width:2px
    style Broker fill:#ff9,stroke:#333,stroke-width:2px