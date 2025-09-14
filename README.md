# Weight Tracker

A command-line weight tracking application built in Go with SQLite database, comprehensive statistics, and beautiful chart visualizations.

## Features

### Core CRUD Operations
- **Add** weight entries with date, unit, and notes
- **List** entries with filtering, sorting, and limiting options
- **Update** existing entries (partial updates supported)
- **Delete** entries with confirmation prompts

### Advanced Features
- **Statistics** command with comprehensive weight analytics
- **Chart Generation** with ASCII terminal charts and interactive HTML charts
- **Time-normalized** chart spacing based on actual entry intervals
- **Unit Support** for both kg and lbs
- **Flexible Date Handling** with automatic defaults

### Data Management
- **SQLite Database** with automatic migrations
- **Type-safe** database operations using `sqlc`
- **Comprehensive Validation** for all input data
- **Error Handling** with clear, actionable messages

## Getting Started

### Quick Start
1. **Install the application** (see Installation section below)
2. **Add your first weight entry**:
   ```bash
   ./weight-tracker add 75.5
   ```
3. **View your entries**:
   ```bash
   ./weight-tracker list
   ```
4. **See your progress with a chart**:
   ```bash
   ./weight-tracker list --graph
   ```
5. **Get detailed statistics**:
   ```bash
   ./weight-tracker stats --verbose
   ```

### First-Time Setup
1. **Create a `.env` file** in your project directory:
   ```bash
   echo "DATABASE_PATH=./weight_tracker.db" > .env
   echo "DATE_INPUT_FORMAT=dd-mm-yyyy" >> .env
   echo "DATE_DISPLAY_FORMAT=dd-mm-yyyy" >> .env
   echo "DEFAULT_UNIT=kg" >> .env
   ```
   
   **For WSL users**: If you prefer to store the database on your Windows filesystem:
   ```bash
   echo "DATABASE_PATH=/mnt/c/weight-tracker/db/weight.db" > .env
   echo "DATE_INPUT_FORMAT=dd-mm-yyyy" >> .env
   echo "DATE_DISPLAY_FORMAT=dd-mm-yyyy" >> .env
   echo "DEFAULT_UNIT=kg" >> .env
   mkdir -p /mnt/c/weight-tracker/db
   ```

2. **Run database migrations** to set up the schema:
   ```bash
   goose -dir migrations sqlite3 ./weight_tracker.db up
   ```
   
   **For WSL users**:
   ```bash
   goose -dir migrations sqlite3 /mnt/c/weight-tracker/db/weight.db up
   ```

3. **The application will automatically**:
   - Create a `charts/` directory for generated visualizations
   - Handle all database operations

4. **Start tracking your weight** - you're ready to go!

## Installation

### Prerequisites
- Go 1.21 or later
- [goose](https://github.com/pressly/goose) for database migrations

### Build from Source
```bash
git clone https://github.com/BlochLior/weight-tracker.git
cd weight-tracker
go mod download
go build -o weight-tracker github.com/BlochLior/weight-tracker
```

## Usage

### Basic Commands

#### Add Weight Entry
```bash
# Basic weight entry (uses today's date and kg unit)
./weight-tracker add 75.5

# With specific date and unit
./weight-tracker add 165.3 --date 15-01-2024 --unit lbs

# With note
./weight-tracker add 75.5 --note "After workout"

# All options
./weight-tracker add 75.5 --date 15-01-2024 --unit kg --note "Morning weight"
```

#### List Entries
```bash
# List all entries
./weight-tracker list

# Filter by date range
./weight-tracker list --from 01-01-2024 --to 31-01-2024

# Filter by unit
./weight-tracker list --unit kg

# Sort by weight (ascending/descending)
./weight-tracker list --sort weight
./weight-tracker list --sort weight --desc

# Sort by date
./weight-tracker list --sort date

# Limit results
./weight-tracker list --limit 10

# Complex filtering
./weight-tracker list --unit kg --sort weight --desc --limit 5
```

#### Update Entry
```bash
# Update weight only
./weight-tracker update 1 --weight 76.0

# Update multiple fields
./weight-tracker update 1 --weight 76.0 --note "Updated note"

# Update date
./weight-tracker update 1 --date 16-01-2024

# Update unit
./weight-tracker update 1 --unit lbs
```

#### Delete Entry
```bash
# Delete with confirmation prompt
./weight-tracker delete 1

# Force delete without confirmation
./weight-tracker delete 1 --force

# Confirm deletion
./weight-tracker delete 1 --confirm
```

### Statistics Command

#### Basic Statistics
```bash
./weight-tracker stats
```
Shows:
- Total entries count
- Average weight
- Weight range (min to max)
- Minimum and maximum weights with entry IDs
- Time span from first to last entry

#### Verbose Statistics
```bash
./weight-tracker stats --verbose
```
Shows all basic statistics plus full entry details for:
- Minimum weight entry (ID, date, weight, note)
- Maximum weight entry (ID, date, weight, note)
- Time span entries (from and to entries with full details)

### Chart Generation

#### ASCII Terminal Charts
```bash
# Generate ASCII chart in terminal
./weight-tracker list --graph
```

#### HTML Charts
```bash
# Generate HTML chart with default filename
./weight-tracker list --graph --output html

# Generate HTML chart with custom filename
./weight-tracker list --graph --output html --file my-weight-chart.html

# Generate PNG chart (converts from HTML)
./weight-tracker list --graph --output png --file chart.png
```

Charts are saved in the `charts/` directory with:
- **Time-normalized spacing**: X-axis reflects actual time intervals between entries
- **Interactive features**: Hover for details, zoom, pan
- **Clean layout**: No text overlaps or formatting issues
- **Professional appearance**: High-quality rendering suitable for reports

## Configuration

### Database
The application uses SQLite with automatic database creation and migrations. The database path is configured via environment variables.

### Environment Variables
Create a `.env` file in your project directory to customize the database path:

**Default setup** (database in current directory):
```
DATABASE_PATH=./weight_tracker.db
```

**WSL setup** (database on Windows filesystem):
```
DATABASE_PATH=/mnt/c/weight-tracker/db/weight.db
```

**Custom path**:
```
DATABASE_PATH=/path/to/your/database.db
```

**Date format configuration**:
```
DATE_INPUT_FORMAT=dd-mm-yyyy    # Input format for CLI commands
DATE_DISPLAY_FORMAT=dd-mm-yyyy  # Display format for output
```

**Weight unit configuration**:
```
DEFAULT_UNIT=kg                 # Default weight unit (kg or lbs)
```

Supported date formats:
- `dd-mm-yyyy` (default) - 15-09-2024
- `mm-dd-yyyy` - 09-15-2024  
- `yyyy-mm-dd` - 2024-09-15
- `dd/mm/yyyy` - 15/09/2024
- `mm/dd/yyyy` - 09/15/2024
- `yyyy/mm/dd` - 2024/09/15

Supported weight units:
- `kg` (default) - Kilograms
- `lbs` - Pounds

The application will automatically create the database file and any necessary directories at the specified path.

## Project Structure

```
weight-tracker/
├── main.go                  # Application entry point
├── cmd/tracker/             # CLI commands and business logic
│   ├── add.go              # Add command implementation
│   ├── add_test.go         # Add command tests (integration + CLI)
│   ├── list.go             # List command with filtering/sorting/graphing
│   ├── list_test.go        # List command tests (integration + CLI + graph)
│   ├── update.go           # Update command with partial updates
│   ├── update_test.go      # Update command tests (integration + CLI)
│   ├── delete.go           # Delete command with confirmations
│   ├── delete_test.go      # Delete command tests (integration + CLI)
│   ├── stats.go            # Statistics command
│   ├── stats_test.go       # Statistics command tests
│   ├── graph.go            # Chart generation logic (ASCII, HTML, PNG)
│   ├── graph_test.go       # Chart generation tests
│   ├── store.go            # Database interface and implementation
│   ├── store_test.go       # Store interface and validation tests
│   ├── store_mock.go       # Mock store for testing
│   ├── app_config.go       # Application configuration (dates, units)
│   ├── app_config_test.go  # Configuration tests with dependency injection
│   ├── helpers.go          # Utility functions for printing
│   ├── helpers_test.go     # Test helper functions
│   └── root.go             # Root command setup
├── internal/db/            # Database connection and migrations
│   ├── db.go              # Database connection logic
│   └── sqlc/              # Generated database code
├── migrations/             # Database schema migrations
│   ├── 20250823093835_create_weights_table.sql
│   └── 20250825105156_alter_weights_table.sql
├── charts/                 # Generated chart files (gitignored)
├── queries.sql            # SQL queries for sqlc generation
├── sqlc.yaml              # sqlc configuration
├── go.mod                 # Go module dependencies
├── go.sum                 # Go module checksums
└── README.md              # This file
```

## Development

### Prerequisites for Development
- Go 1.21 or later
- [goose](https://github.com/pressly/goose) for database migrations
- [sqlc](https://sqlc.dev/) for type-safe SQL code generation

### Running Tests
```bash
# Run all tests
go test ./cmd/tracker

# Run specific test suites
go test ./cmd/tracker -v -run "TestAddCommand"
go test ./cmd/tracker -v -run "TestStatsCommand"

# Run with coverage
go test ./cmd/tracker -cover
```

### Database Migrations
```bash
# Run migrations
goose -dir migrations sqlite3 weight_tracker.db up

# Create new migration
goose -dir migrations create migration_name sql
```

### Code Generation
```bash
# Generate sqlc code (only needed when modifying queries.sql)
sqlc generate
```

## 🤝 Contributing

### Clone the Repository

```bash
git clone https://github.com/BlochLior/weight-tracker.git
cd weight-tracker
```

### Build the Project

```bash
# Download dependencies
go mod download

# Build the application
go build -o weight-tracker github.com/BlochLior/weight-tracker
```

### Run the Project

```bash
# Run with Go (development)
go run main.go add 75.5

# Or run the built binary
./weight-tracker add 75.5
```

### Run the Tests

```bash
# Run all tests
go test ./cmd/tracker

# Run with verbose output
go test ./cmd/tracker -v

# Run with coverage
go test ./cmd/tracker -cover

# Run specific test suites
go test ./cmd/tracker -run "TestAddCommand"
go test ./cmd/tracker -run "TestStatsCommand"
```

### Development Setup

1. **Set up your environment**:
   ```bash
   # Create .env file for local development
   echo "DATABASE_PATH=./weight_tracker_dev.db" > .env
   echo "DATE_INPUT_FORMAT=dd-mm-yyyy" >> .env
   echo "DATE_DISPLAY_FORMAT=dd-mm-yyyy" >> .env
   echo "DEFAULT_UNIT=kg" >> .env
   ```

2. **Run database migrations**:
   ```bash
   goose -dir migrations sqlite3 ./weight_tracker_dev.db up
   ```

3. **Generate sqlc code** (if you modify `queries.sql`):
   ```bash
   sqlc generate
   ```

### Submit a Pull Request

If you'd like to contribute:

1. **Fork the repository** on GitHub
2. **Create a feature branch** from `main`:
   ```bash
   git checkout -b feature/your-feature-name
   ```
3. **Make your changes** and ensure all tests pass:
   ```bash
   go test ./cmd/tracker
   ```
4. **Commit your changes** with clear, descriptive messages
5. **Push to your fork** and open a pull request to the `main` branch

### Contribution Guidelines

- **Code Style**: Follow Go conventions and run `gofmt`
- **Testing**: Add tests for new functionality
- **Documentation**: Update README.md for new features
- **Commits**: Use clear, descriptive commit messages
- **Pull Requests**: Provide a clear description of changes

## Testing

The application includes comprehensive test coverage:

### Unit Tests
- **MockStore**: Fast, isolated testing of business logic
- **Validation**: Input validation and error handling
- **Statistics**: Calculation accuracy and edge cases
- **Chart Generation**: ASCII and HTML chart creation

### Integration Tests
- **Database Operations**: Real SQLite testing with in-memory databases
- **CLI Commands**: End-to-end command execution
- **Migration Testing**: Database schema evolution
- **Error Scenarios**: Invalid inputs and edge cases

### Test Categories
- **CRUD Operations**: Add, list, update, delete functionality
- **Filtering & Sorting**: Complex query scenarios
- **Statistics**: Mathematical calculations and data processing
- **Chart Generation**: Visual output and file handling
- **Validation**: Input sanitization and error messages

## Dependencies

### Core Dependencies
- **Cobra**: CLI framework for command structure
- **SQLite**: Embedded database for data persistence
- **sqlc**: Type-safe SQL code generation
- **goose**: Database migration management

### External Chart Dependencies
- **go-echarts**: HTML chart generation
