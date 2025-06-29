# **🛍️ E-Commerce Backend API**

A robust and scalable backend API built with **Go**, **Fiber**, and **Firebase**, providing essential services for an e-commerce platform, including user authentication, authorization, and data management.

## ✨ Features

- **🚀 High Performance:** Utilizes Go and Fiber for optimized speed and efficiency.
- **🔒 Secure Authentication:** Employs Firebase Authentication for secure user management.
- **🗄️ Data Management:** Leverages Firestore for flexible and scalable data storage.
- **🛡️ Middleware Support:** Includes authentication middleware for protected routes.
- **⚙️ Environment Configuration:** Uses `.env` files for easy configuration.
- **🐳 Docker Support:** Ready for containerization with a provided Dockerfile.
- **⏰ Automated Tasks**: Includes a cron job to keep the backend service alive, ensuring continuous operation.

## 🛠️ Technologies Used

| Technology  | Description                                      |
| :---------- | :----------------------------------------------- |
| Go          | Backend logic and API development             |
| Fiber       | Web framework for handling HTTP requests        |
| Firebase    | Authentication and real-time database services |
| Firestore   | NoSQL cloud database for storing user data   |
| Docker      | Containerization for deployment                |
| go-dotenv   | Loading environment variables from .env files |
| cron        | Job scheduling for automated tasks |

## 📦 Installation

### Prerequisites
- [Go](https://go.dev/dl/) (version 1.24 or higher)
- [Docker](https://www.docker.com/get-started) (optional, for containerization)
- [Firebase project](https://console.firebase.google.com/)

### Steps

1.  **Clone the Repository:**
    ```bash
    git clone https://github.com/samueltuoyo15/Shopping-Backend.git
    cd shopping-backend
    ```
2.  **Install Dependencies:**
    ```bash
    go mod download
    ```
3.  **Set up Environment Variables:**

    -   Create a `.env` file in the project root.
    -   Add the following variables with your specific values:
        ```
        PORT=5000
        FRONTEND_DOMAIN=http://localhost:3000
        FIREBASE_API_KEY=YOUR_FIREBASE_API_KEY
        BACKEND_DOMAIN=http://localhost:5000
        GOLANG_ENV=development
        ```
    - Replace `YOUR_FIREBASE_API_KEY` with your actual Firebase API Key from your Firebase project settings.
4.  **Firebase Service Account Setup:**

    -   Download your Firebase service account key JSON file and place it in the project root as `serviceAccountKey.json`.
5.  **Run the Application:**
    ```bash
    go run ./pkg/main.go
    ```
    Or with Air for hot reloading:
     ```bash
     air
     ```
## ⚙️ Environment Variables

| Variable          | Description                                                                                       | Example                       |
| :---------------- | :------------------------------------------------------------------------------------------------ | :---------------------------- |
| `PORT`            | Port number for the server to listen on                                                          | `5000`                        |
| `FRONTEND_DOMAIN` | The domain of the frontend application, used for CORS configuration                               | `http://localhost:3000`       |
| `FIREBASE_API_KEY`| API key for authenticating with Firebase services                                                | `AIzaSyDOCakljv2sEXAMPLE`    |
| `BACKEND_DOMAIN`  | The backend domain, used by the cron job to send a keep-alive request to prevent the service from idling | `http://localhost:5000`      |
|`GOLANG_ENV`      | Used to configure secure cookies. If set to `production` cookies will be secure                  | `development` or `production`|

## 📖 API Documentation

### Base URL

```
https://shopping-backend-9cf2.onrender.com
```

### Endpoints

#### POST /api/auth/register
Registers a new user.

**Request**:
```json
{
    "email": "user@example.com",
    "full_name": "John Doe",
    "password": "securePassword"
}
```

**Response**:
```json
{
    "message": "User created successfully. Go ahead and login into your account"
}
```

**Errors**:
- 400: Invalid request body or validation errors
- 409: User with this email already exists
- 500: Failed to create user or user record

#### POST /api/auth/login
Logs in an existing user.

**Request**:
```json
{
    "email": "user@example.com",
    "password": "securePassword"
}
```

**Response**:
```json
{
    "message": "Login successful",
    "uid": "firebase_user_uid",
    "accessToken": "generated_access_token"
}
```

**Errors**:
- 400: Invalid request body or validation errors
- 404: User not found
- 401: Invalid credentials

#### GET /api/user/me
Retrieves the current user's information. Requires a valid access token.

**Request**:
Headers:
```
Authorization: Bearer <access_token>
```
or uses the accessToken cookie.

**Response**:
```json
{
    "email": "user@example.com",
    "fullname": "John Doe",
    "createdAt": "timestamp"
}
```

**Errors**:
- 401: Unauthorized (missing or invalid access token)
- 404: User not found


#### GET /categories/getCategories
Retrieves the list of available list of categories of products e.g Electronics. Results are cached using Redis to improve performance by the way.


**Response**:
```json
{"count":15,
"list":["Baby's & Toy's","Camera","Computers","Electronics","Gaming","Groceries & Pets","HeadPhone","Health & Beauty","Home & Lifestyle","Man's Fashion","Medicine","Phones","SmartWatch","Sports & Outdoors","Woman's Fashion"],"source":"firestore",
"success":true}
```


**Caching**:
- This endpoint is cached using Redis to make the response time faster
- cached results are automatically invalidated after 5 minutes

**Errors**:
- 500: Failed to fetch categories or Internal Server Error



#### GET /categories/getProducts
Retrieves the list of available products. Results are cached using Redis to improve performance by the way.


**Response**:
```json
[
  {
    "id": 1,
    "title": "Fjallraven - Foldsack No. 1 Backpack",
    "price": 109.95,
    "description": "Your perfect pack for everyday use and walks in the forest.",
    "category": "men's clothing",
    "image": "https://fakestoreapi.com/img/81fPKd-2AYL._AC_SL1500_.jpg",
    "rating": {
      "rate": 3.9,
      "count": 120
    }
  },
  ...
]
```

### Optional Query Parameters
This endpoint supports query parameters to filter or limit the results, which will also affect caching. If different query parameters are provided, a separate cache will be created for that unique query.

Examples:
```
GET /api/categories/getProducts?limit=5
Returns only 5 products from the list
```

```
GET /api/categories/getProducts?sort=desc
Returns products sorted in descending order 
```

Note: Any combination of query params (e.g. ?sort=asc&limit=3) will generate a unique cache key, ensuring accurate responses.


**Caching**:
- This endpoint is cached using Redis to make the response time faster
- cached results are automatically invalidated after 5 minutes

**Errors**:
- 500: Failed to fetch products or Internal Server Error


#### GET /api/health-check
health check endpoint.

**Response**:
```json
{
    "status": "Ok",
    "uptime": "1m30s",
    "memoryUsage": {
        "alloc": 2137344,
        "totalAlloc": 3317248,
        "sys": 18454560,
        "numGC": 3
    }
}
```

## 🐳 Docker

1.  **Build the Docker Image:**
    ```bash
    docker build -t shopping-backend .
    ```
2.  **Run the Docker Container:**
    ```bash
    docker run -p 5000:5000 -e PORT=5000 -e FRONTEND_DOMAIN=http://localhost:3000 -e FIREBASE_API_KEY=<YOUR_FIREBASE_API_KEY> shopping-backend
    ```
    Remember to replace `<YOUR_FIREBASE_API_KEY>` with your Firebase API key.

## 📜 License

This project is licensed under the [MIT License](LICENSE).


[![Readme was generated by Dokugen](https://img.shields.io/badge/Readme%20was%20generated%20by-Dokugen-brightgreen)](https://www.npmjs.com/package/dokugen)
