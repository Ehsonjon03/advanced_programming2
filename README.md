## Assignment 3: Event-Driven Architecture (EDA)

### Project Overview
[cite_start]This project demonstrates an asynchronous event-driven flow between microservices using **RabbitMQ** as a message broker[cite: 17, 18].

### Architecture Diagram
```mermaid
graph TD
    Client[Postman / Browser] -->|HTTP Request| OrderService[Order Service :8081]
    OrderService -->|gRPC Call| PaymentService[Payment Service :50051]
    
    subgraph "Infrastructure"
        DB[(PostgreSQL)]
        Broker[RabbitMQ Broker]
    end

    OrderService --- DB
    PaymentService --- DB
    
    PaymentService -->|Publish: payment.completed| Broker
    Broker -->|Consume: payment.completed| NotificationService[Notification Service]
    
    NotificationService -->|Log| Console[Console: Email Simulated]