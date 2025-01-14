to build: 
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o bootstrap *.go

# Reading-Tracker Application

### Description
A serverless reading-tracker system that:

- Stores **book metadata** in DynamoDB (the “Books” table).
- Lets users create **personal lists** in another DynamoDB table (“UserLists”), referencing books by ISBN.
- Uses **AWS Cognito** for user authentication and an API Gateway **Cognito Authorizer** to secure routes.

### Key Features
1. **Books**  
   - **CRUD** operations on `/books` (Create, Read, Update, Delete).  
   - Data is stored in `Books` DynamoDB table (keyed by `isbn`).
2. **Personal Lists**  
   - **CRUD** operations on `/lists`.  
   - Each user can have multiple named lists keyed by `(userId, listName)`.
   - Uses a **Cognito** token (JWT) to identify the user.
3. **User Authentication**  
   - Endpoints at `/auth/signup`, `/auth/confirm`, `/auth/signin` for sign-up/confirmation/sign-in flows.  
   - Requires valid `Authorization: Bearer <token>` for protected routes.

### Architecture
1. **AWS SAM** manages CloudFormation resources:
   - **DynamoDB** tables for `Books` and `UserLists`.
   - **Lambda Functions**:
     - `BooksFunction` for `/books` routes.
     - `ListsFunction` for `/lists` routes.
     - Possibly an `AuthFunction` for `/auth` routes (or combined).
   - **Amazon Cognito** for user accounts.
   - **API Gateway** (named `BooksApi`) with `REGIONAL` endpoint & CORS support.
2. **Go** code structured in multiple handler files:
   - `booksHandler.go` for book operations.
   - `listHandler.go` for personal lists logic.
   - `authenticationHandler.go` for sign-up, confirm, sign-in.
3. **Main** router in `main.go` uses path-based logic to dispatch:
   - `/books` → `BooksHandler`  
   - `/lists` → `ListsHandler`  
   - `/auth` → `AuthHandler`

### Deployment & Setup
1. **Clone** the repository with the SAM template and Go source.
2. **Configure** environment variables (e.g., `USER_POOL_ID`, `USER_POOL_CLIENT_ID`) if needed.
3. **Build & Deploy**:
   ```bash
   sam build
   sam deploy --guided
   ```
4. **Check** the CloudFormation output for your API endpoint, Cognito IDs, etc.

### Usage
1. **Sign Up**:  
   ```bash
   POST /auth/signup
   {
     "username": "johndoe",
     "password": "MySecurePassword1!",
     "email": "john@example.com"
   }
   ```
2. **Confirm** (check your email for the code):
   ```bash
   POST /auth/confirm
   {
     "username": "johndoe",
     "code": "123456"
   }
   ```
3. **Sign In**:
   ```bash
   POST /auth/signin
   {
     "username": "johndoe",
     "password": "MySecurePassword1!"
   }
   ```
   - Response returns an `IdToken`, `AccessToken`, `RefreshToken`.
4. **Create a List**:
   ```bash
   POST /lists/MyFavorites
   Authorization: Bearer <ACCESS_TOKEN>
   {
     "isbns": ["0451526538", "12345ABC"],
     "description": "Favorites from 2023"
   }
   ```
5. **Get Your Lists**:
   ```bash
   GET /lists
   Authorization: Bearer <ACCESS_TOKEN>
   ```
6. **Add/Remove/Update** your list items with `PUT`, or delete an entire list with `DELETE /lists/MyFavorites`.