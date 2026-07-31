# [JS-DOCKER-002] NGINX SPA Config

Statements: `[JS-DOCKER-002]`

```nginx
server {
    listen 8080;
    server_name localhost;

    root /usr/share/nginx/html;

    index index.html;

    location / {
        # This is critical for SPAs (History Mode).
        # If the requested file isn't found, fall back to index.html
        try_files $uri $uri/ /index.html;
    }

    # Basic error handling
    error_page 500 502 503 504 /50x.html;
    location = /50x.html {
        root /usr/share/nginx/html;
    }
}
```
