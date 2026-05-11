graph TD
Client[Postman / Browser] -->|1. HTTP Request| OrderService[Order Service :8081]

    subgraph "Order Service Internal"
        Handler[HTTP Handler] --> UC[Order UseCase]
        UC --> Repo[Order Repository]
    end

    subgraph "Infrastructure"
        DB[(PostgreSQL)]
        Redis[(Redis Cache)]
        Broker[RabbitMQ Broker]
    end

    %% Логика Cache-aside
    Repo -.->|2. Check Cache| Redis
    Repo -.->|3. Read/Write| DB
    Repo -.->|4. Update Cache| Redis

    %% gRPC взаимодействие
    UC -->|5. gRPC: ProcessPayment| PaymentService[Payment Service :50051]
    PaymentService --- DB
    
    %% Асинхронный поток (из Assignment 3)
    PaymentService -->|6. Publish| Broker
    Broker -->|7. Consume| NotificationService[Notification Service]
    NotificationService -->|8. Log| Console[Console: Email Simulated]