## API DESIGN

# User Profile
1. **Create User Profile**
    - Endpoint: `POST /api/users`
    - Authorization: -
    - Request:
```
    {
    "nickname": "unique_nickname",
    "first_name": "John",
    "last_name": "Doe",
    "password": "user_password"
    }
```
    - Response:
```
    {
    "oid": "UUID",
    "message": "User profile created successfully."
    }
```

2. **Log In**
    - Endpoint: `POST /api/users/login`
    - Authorization: Basic Auth
    - Request:
```
    {
    "nickname": "unique_nickname",
    "password": "user_password"
    }
```
    - Response:
```
    {
    "token": "JWT_token",
    "message": "Successfully logged in"
    }
```

3. **Update User Profile**
    - Endpoint: PUT `/api/users/{user_id}`
    - Authorization: Bearer(JWT)
    - Request:
```
    {
    "nickname": "unique_nickname",
    "first_name": "UpdatedJohn",
    "last_name": "UpdatedDoe"
    }
```
    - Response:
```
    {
    "message": "User profile updated successfully."
    }
```
4. **Change Password**
    - Endpoint: `PUT /api/users/{user_id}/password`
    - Authorization: Bearer(JWT)
    - Request:
```
    {
    "password": "new_password"
    }
```
    - Response:
```     
    {
    "message": "Password updated successfully."
    }
```
5. **Get User Profile**
    - Endpoint: Endpoint: `GET /api/users/{user_id}`
    - Authorization: -
    - Request: -
    - Response:
```
    {
    "oid": "UUID",
    "nickname": "unique_nickname",
    "first_name": "John",
    "last_name": "Doe",
    "created_at": "timestamp",
    "updated_at": "timestamp",
    "state": 1
    }
```
6. **List User Profiles (with Pagination)**
    - Endpoint: `GET /api/users?page={page_number}&limit={page_size}`
    - Authorization: -
    - Request: - 
    - Response:
```
    {
    "total_users": "total_users_count",
    "page": "current_page_number",
    "users": [
        {
        "oid": "UUID",
        "nickname": "unique_nickname",
        "first_name": "John",
        "last_name": "Doe",
        "created_at": "timestamp",
        "updated_at": "timestamp",
        "state": 1
        },
        //another user profiles
    ]
    }
```

7. **Delete User Profile**
    - Endpoint: Endpoint: `DELETE /api/users/{user_id}`
    - Authorization: Bearer(JWT)
    - Request: -
    - Response:
```
    {
        "message": "Profile successfully deleted"
    }
```
7. **Vote**
- Endpoint: Endpoint: `POST /api/vote/{user_id}`
- Authorization: Bearer(JWT)
- Request: 
```
{
    "oid": "oid of user that you want to rate",
    "value": 1 (or -1)
}
```
- Response:

```
    {
        "message": "Your vote has been submitted"
    }
```

8. **Change vote**
- Endpoint: Endpoint: `PUT /api/vote/`
- Authorization: Bearer(JWT)
- Request: 
```
{
    "oid": "oid of user that you want to rate",
    "value": 1/-1 (or 0 if you want deactivate your vote)
}
```
- Response:
```
    {
        "message": "Your vote has been changed"
    }
```

## Database Tables:
1. User Profiles Table:
    - id (Primary Key) int
    - oid UUID
    - nickname (Unique) string
    - first_name string
    - last_name string
    - password string (hash)
    - created_at timestamp
    - updated_at timestamp
    - state int
    - user_role int
    - rating
2. Votes Table:
    - id (Primary Key) int
    - from_oid UUID (Foreign Key for oid from, user profiles table)
    - to_oid UUID 
    - value int
    - voted_at timestamp