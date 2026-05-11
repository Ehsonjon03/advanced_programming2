graph TD
Client[Postman / Browser] -->|HTTP Request| OrderService[Order Service :8081]

    subgraph "Infrastructure"
        DB[(PostgreSQL)]
        Redis[(Redis Cache)]
        Broker[RabbitMQ Broker]
    end

    %% Логика Cache-aside (Assignment 4)
    OrderService -.->|1. Читает| Redis
    OrderService -.->|2. Пишет (при промахе)| Redis
    OrderService --- DB

    %% gRPC взаимодействие (Assignment 4)
    OrderService -->|3. Синхронный gRPC Call| PaymentService[Payment Service :50051]
    PaymentService --- DB

    %% Асинхронный поток (Assignment 3, сохранен)
    PaymentService -->|4. Publish: payment.completed| Broker
    Broker -->|5. Consume: payment.completed| NotificationService[Notification Service]
    
    NotificationService -->|6. Log| Console[Console: Email Simulated]