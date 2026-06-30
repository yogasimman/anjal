# Test Env Resolution

```http
GET {{.API_HOST}}/todos/1
Accept: application/json
```

# Test with Auth Placeholder

```http
GET {{.API_HOST}}/posts/1
@auth bearer {{.MY_TOKEN}}
Accept: application/json
```
