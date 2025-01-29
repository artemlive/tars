# TARS Bot
Task And Request Sorter for slack

TARS is a Slack bot designed to help track and manage reactions in Slack channels, categorize them, and store them in a database for analytics and reporting.

---

## Table of Contents

- [Features](#features)
- [Configuration](#configuration)
- [Installation](#installation)
- [Usage](#usage)
- [Development](#development)
- [License](#license)

---

## Features

- Automatically track reactions in Slack channels.
- Categorize reactions based on configurable rules.
- Save stats to a database for future analysis.
- Fetch and visualize statistics (e.g., pie charts).

---

## Configuration

The configuration file is written in YAML format and includes Slack, database, bot settings, and channel-specific rules. Below is an example configuration file:

```yaml
slack:
  app_token: "xapp"
  bot_token: "xoxb"
  signing_secret: ""

db:
  dsn: db/tars.db
  driver: sqlite

bot:
  log_level: "info"

channels:
  - name: "#support"
    id: "C089TUGAT9V"
    rules:
      - reaction: "cd"
        category: "CI/CD"
      - reaction: "bug"
        category: "Infra bug"
    beacon_reaction: ":eyes:"
```

### Configuration Fields

#### Slack Settings (`slack`)
- **app_token**: Slack app-level token.
- **bot_token**: Slack bot token.
- **signing_secret**: Slack signing secret.

#### Database Settings (`db`)
- **dsn**: Path to the database file or DSN for your database.
- **driver**: Database driver (e.g., `sqlite`).

#### Bot Settings (`bot`)
- **log_level**: Log verbosity level (`info`, `debug`, `error`).

#### Channels (`channels`)
- **name**: The display name of the Slack channel.
- **id**: The channel ID in Slack.
- **rules**: Define mapping between reactions and categories.
  - **reaction**: Emoji reaction (e.g., `cd` or `bug`).
  - **category**: The category associated with the reaction.
- **beacon_reaction**: A special reaction used as a beacon for monitoring.

---

## Installation

### Prerequisites
- Go (1.19+)
- SQLite (if using SQLite as the database) (it's the only supported database for now)
- Slack App with required tokens and permissions

### Steps
1. Clone the repository:
   ```bash
   git clone https://github.com/artemlive/tars.git
   cd tars
   ```
2. Build the project:
   ```bash
   make build
   ```
3. Run tests:
   ```bash
   make test
   ```
4. Configure the application:
   - Create a YAML configuration file based on the example above.

---

## Usage

### Running the Bot
```bash
./tars -config /path/to/config.yaml
```

### Features
- **Track Reactions**: Automatically monitor and categorize reactions in configured Slack channels.
- **Fetch Stats**: Generate and visualize statistics via Slack shortcuts.

---

## Development

### TODO

Here are the planned improvements and enhancements for the project:
- [ ] **Implement Multiple Log Levels**  
      Introduce more granular logging options to support different verbosity levels, such as debug, info, warning, and error.

- [ ] **Add More Features**  
      Introduce new features like JIRA ticket creation.  
      Add more chart types and interactive handlers. Currently, the bot only supports two shortcuts with a pie chart.

- [ ] **Add More Tests**  
      Expand the test suite to ensure better code coverage and improve overall reliability.

- [ ] **Support for Additional Storage Backends**  
      Extend storage compatibility beyond SQLite to include other databases like PostgreSQL or MySQL.

- [ ] **Implement CI/CD Pipeline**  
      Set up a continuous integration and deployment pipeline to automate testing and deployment processes.

- [ ] **Improve Documentation**


### Setting Up the Development Environment
1. Install dependencies:
   ```bash
   go mod tidy
   ```
2. Run tests:
   ```bash
   make test 
   make test-coverage
   ```

### Code Structure
- `pkg/slack`: Slack integration logic.
- `pkg/storage`: Database logic.
- `pkg/utils`: Configuration and utility functions.
- `cmd/tars`: Main application entry point.

---

## License

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.

---

## Contributing

Feel free to fork the repository and submit a pull request with your improvements or fixes. Contributions are welcome!

---

Happy monitoring with TARS! 🎉

