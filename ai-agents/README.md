# Go Coffee Co. AI Agent Ecosystem

This directory contains the implementation of the AI agent ecosystem for "Go Coffee Co.", designed to automate and coordinate various operational aspects of the coffee shops.

## Agents Implemented:

- **Beverage Inventor Agent**: Generates new drink recipes.
- **Tasting Coordinator Agent**: Schedules and manages tasting sessions.
- **Inventory Manager Agent**: Tracks real-time inventory levels.
- **Notifier Agent**: Disseminates critical information and alerts.
- **Feedback Analyst Agent**: Collects, analyzes, and summarizes customer feedback.
- **Scheduler Agent**: Manages daily operational schedules.
- **Inter-Location Coordinator Agent**: Facilitates communication and coordination between locations.
- **Task Manager Agent**: Creates, assigns, and tracks tasks in ClickUp.
- **Social Media Content Agent**: Generates engaging social media content.

## Structure:

Each agent resides in its own subdirectory (e.g., `beverage-inventor-agent/`) and typically contains:

- `main.go`: The main application logic for the agent.
- `config.yaml`: Configuration parameters specific to the agent.

## Getting Started:

To run an individual agent, navigate to its directory and execute:

```bash
go run main.go
```

## Next Steps:

The current implementations are basic skeletons. Future development will involve:

- Implementing the core logic for each agent based on the detailed plan.
- Integrating with external systems (ClickUp, Google Sheets, Airtable, Slack, Social Media Platforms, AI/LLM Services, Supplier & Delivery Systems).
- Setting up a communication bus (e.g., Kafka) for inter-agent communication.
- Developing a central orchestration layer.
- Containerizing agents using Docker and deploying with Kubernetes.