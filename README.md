# SWIFT Codes API

This project is a solution to the Remitly internship home assignment.  
It provides a RESTful API to manage SWIFT (BIC) codes for banks and their branches.

---

## Features

- Parses SWIFT codes from CSV on app start
- Automatically detects headquarters and branches
- Stores data in PostgreSQL with proper indexing
- Provides REST API:
  - ✅ Get SWIFT code details
  - ✅ Get all codes by country
  - ✅ Add new SWIFT code (with validation)
  - ✅ Delete SWIFT code (safe HQ delete prevention)
- Automatically adds placeholder HQ if needed
- Full Dockerized setup

---

## Business Rules

- A SWIFT code with suffix `XXX` is treated as a headquarters
- Branches are matched to HQs via first 8 characters
- Placeholder HQs (`UNKNOWN`) are inserted for orphan branches
- Country codes and names are always stored uppercased
- Validation ensures data correctness and integrity

---

## Running the Project

### Prerequisites
- **Docker** and **docker-compose** installed

### 1. Clone the repo

```bash
git clone https://github.com/zapjur/swift-api.git
cd swift-api
```

### 2. Start with Docker 🐳

```bash
docker-compose up --build -d
```

The app will be available at: `http://localhost:8080`

CSV is parsed at startup from: `./assets/swift_codes.csv`

### 3. Stop the app

```bash
docker-compose down -v
```

---

## API Endpoints

### Get SWIFT code details

```
GET /v1/swift-codes/{swiftCode}
```

**Response Structure** for headquarter swift code:

```json
{
  "address": "",
  "bankName": "",
  "countryISO2": "",
  "countryName": "",
  "isHeadquarter": true,
  "swiftCode": "",
  "branches": [
    {
      "address": "",
      "bankName": "",
      "countryISO2": "",
      "isHeadquarter": false,
      "swiftCode": ""
    },
    {
      "address": "",
      "bankName": "",
      "countryISO2": "",
      "isHeadquarter": false,
      "swiftCode": ""
    }
  ]
}

```

**Response Structure** for branch swift code:

```json
{
  "address": "",
  "bankName": "",
  "countryISO2": "",
  "countryName": "",
  "isHeadquarter": false,
  "swiftCode": ""
}
```

---

### Get all SWIFT codes by country

```
GET /v1/swift-codes/country/{countryISO2}
```

**Response Structure**:

```json
{
  "countryISO2": "",
  "countryName": "",
  "swiftCodes": [
    {
      "address": "",
      "bankName": "",
      "countryISO2": "",
      "isHeadquarter": true,
      "swiftCode": ""
    },
    {
      "address": "",
      "bankName": "",
      "countryISO2": "",
      "isHeadquarter": false,
      "swiftCode": ""
    }
  ]
}

```

---

### Add new SWIFT code

```
POST /v1/swift-codes
```

```json
{
  "address": "Street 123",
  "bankName": "Bank S.A.",
  "countryISO2": "PL",
  "countryName": "POLAND",
  "isHeadquarter": true,
  "swiftCode": "BANKPLPWXXX"
}
```

**Response Structure**:

```json
{
  "message": ""
}
```

---

### Delete SWIFT code

```
DELETE /v1/swift-codes/{swiftCode}
```

**Response Structure**:

```json
{
  "message": ""
}
```

---

##  Sample `curl` Requests

```bash
# Add HQ
curl -X POST http://localhost:8080/v1/swift-codes \
  -H "Content-Type: application/json" \
  -d '{
    "address": "HQ Address",
    "bankName": "HQ Bank",
    "countryISO2": "PL",
    "countryName": "POLAND",
    "isHeadquarter": true,
    "swiftCode": "TESTPLHQXXX"
}'

# Add Branch
curl -X POST http://localhost:8080/v1/swift-codes \
  -H "Content-Type: application/json" \
  -d '{
    "address": "Branch Address",
    "bankName": "Branch Bank",
    "countryISO2": "PL",
    "countryName": "POLAND",
    "isHeadquarter": false,
    "swiftCode": "TESTPLHQ001"
}'

# Get HQ
curl -X GET http://localhost:8080/v1/swift-codes/TESTPLHQXXX

# Get Branch
curl -X GET http://localhost:8080/v1/swift-codes/TESTPLHQ001

# Get all by country
curl -X GET http://localhost:8080/v1/swift-codes/country/PL

# Delete HQ
curl -X DELETE http://localhost:8080/v1/swift-codes/TESTPLHQXXX

# Delete Branch
curl -X DELETE http://localhost:8080/v1/swift-codes/TESTPLHQ001
```

---

## Testing

This project includes unit and integration tests for parsing, repository operations, validation, and API handlers.

All tests are run inside a Docker container to ensure environment consistency.
Test results `test-report.txt` and test coverage `coverage.html` are automatically saved locally after each run.

### Run tests

#### Linux/MacOS

```bash
# Make sure the script is executable
chmod +x run_tests.sh

# Run tests
./run_tests.sh
````

#### Windows

```bash
bash run_tests.sh
```

### What happens during the test run?

- It spins up a fresh PostgreSQL database and API container.
- It builds the project.
- It runs all Go tests (go test ./... -v) inside the container.
- It generates:
    - `test-report.txt` with detailed test results
    - `coverage.html` with interactive test coverage report
- It copies these files to your local filesystem.
- It automatically cleans up all test containers and network.

### After running the tests, you can check the results:

- Test logs here: `./test-report.txt`
- Test coverage report here: `./coverage.html`

You can open `coverage.html` in your browser to check **which parts of code are covered.**

---

## Project Structure

```
.
├── cmd/                  # App entrypoint
├── pkg/
│   ├── handlers/         # HTTP handlers
│   ├── parser/           # CSV parsing
│   ├── repository/       # DB access
│   └── models/           # Shared models
│── internal/
│   └──database/          # DB connection logic
├── scripts/              # DB Schema
├── assets/               # Input CSV file
├── docker-compose.yml
└── Dockerfile
```

---

## Built With

- **Go** 1.24
- **PostgreSQL**
- **net/http** + gorilla/mux
- **Docker** & **docker-compose**

---

## Design Decisions

- **Single Table Schema**: All SWIFT codes — both headquarters and branches — are stored in a single `swift_codes` table. A boolean flag `is_headquarter` clearly distinguishes between them. This simplifies database design, querying, and indexing, especially for country- or HQ-specific lookups.
- **Data Integrity First**: The schema enforces constraints (e.g. foreign key on `headquarter_swift_code`) and indexing (e.g. by `country_iso2`) to ensure data consistency and fast access. This is further reinforced at the application level with strong validation.
- **Placeholder HQ Insertion**: When a branch is added before its headquarter exists, either via CSV or API, a placeholder headquarter is created. This prevents insert failures due to foreign key constraints while maintaining referential integrity.
- **CSV Parsing on Startup**: Instead of a CLI or separate migration command, the CSV is parsed and imported when the application starts. This simplifies deployment and ensures that the app can be bootstrapped easily with the correct data.
- **Repository Pattern**: The use of a `repository` layer abstracts database logic away from the HTTP layer. This promotes clean architecture and makes the codebase easier to test, maintain, and evolve.
- **Raw SQL (No ORM)**: To maximize performance, readability, and full control over query structure, raw SQL is used over an ORM. This is especially suitable for small, focused projects like this one, where the data model is stable and not overly complex.

---

## Validation Rules

To ensure only valid and consistent data enters the system, the following validations are applied:

- `swiftCode` must be exactly **11 characters**
- `swiftCode` ending in `"XXX"` **must** be marked as a **headquarter**
- Non-headquarter `swiftCode` **must not** end with `"XXX"`
- `countryISO2` must be a **2-letter uppercase** ISO-3166 country code
- All fields are **required** and cannot be empty
- A `swiftCode` **must not already exist** in the database
- When adding a **branch**, its headquarter is inferred from the first 8 characters + `"XXX"`
    - If no such headquarter exists in the database, a placeholder HQ is inserted to maintain referential integrity.
    - The placeholder contains minimal information: fields like bankName, townName, and countryName are set to `UNKNOWN`, address is set to `null`, and timezone defaults to `Etc/UTC`.
    - If the real headquarter is later added via the API `POST /v1/swift-codes`, it **automatically replaces** the placeholder with the actual data.
- A headquarter **cannot be deleted** if branches referencing it still exist — the API returns an error

---
