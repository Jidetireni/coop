# Cooperative Management System API

A RESTful API for managing a cooperative society. The system supports features such as member management, savings, loan applications, repayments, and reports generation.

## Features

- **Member Management**: Add, view, update, and delete members (Admin only).
- **Savings Management**: Add and view savings for members.
- **Loan Management**: Apply for loans, view loan status, and update loan status (Admin only).
- **Repayment Management**: Record and view repayments for loans.
- **Reports**: Generate detailed reports for the cooperative admin.

---

## API Flow

### 1. **Member Flow**
- **View Savings**: `GET /savings/{member_id}`
- **Apply for Loan**: `POST /loans`
- **View Loan Status**: `GET /loans/{loan_id}`
- **Record Repayment**: `POST /repayments`

### 2. **Admin Flow**
- **Manage Members**: 
  - `GET /members`
  - `POST /members`
  - `PUT /members/{member_id}`
  - `DELETE /members/{member_id}`
- **Approve Loan**: `PUT /loans/{loan_id}`
- **View Reports**: `GET /reports`

---

## How to Run

1. Clone the repository:
   ```bash
   git clone https://github.com/Jidetireni/coop.git
   ```
2. Navigate to the project directory:
   ```bash
   cd coop
   ```
3. Copy the environment variables template:
   ```bash
   cp .env.example .env
   ```
4. Edit the `.env` file and provide the correct values for your PostgreSQL database and other configuration settings. 

5. Install the required dependencies:
   ```bash
   go mod tidy
   ```

6. Start the server:
   ```bash
   go run main.go
   ```
8. The API will be accessible at `http://localhost:<PORT>`.

---

## Technologies Used

- **Go**: Programming language for backend development.
- **Gin**: Web framework for building the API.
- **PostgreSQL**: Relational database for storing data.
- **Gorm**: ORM library for interacting with the PostgreSQL database.
- **JWT**: Authentication and authorization.

---

## Contributing

Contributions are welcome! To contribute:
1. Fork the repository.
2. Create a new branch for your feature or bug fix.
3. Submit a pull request with detailed explanations of your changes.

---

## License

This project is licensed under the MIT License. See the LICENSE file for details.