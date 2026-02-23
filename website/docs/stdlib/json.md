# JSON

## Parsing JSON

```haira
import "json"

text = '{"name": "Alice", "age": 30}'
data, err = json.parse(text)
if err != nil {
    io.println("Parse error: ${err}")
}

name = data["name"]  // "Alice"
```

## From HTTP Responses

```haira
resp, err = http.get("https://api.example.com/user")
data = resp.json()

name = data["name"]
email = data["email"]
```

## Encoding JSON

```haira
data = {
    "name": "Alice",
    "age": 30,
    "active": true
}
text = json.encode(data)
// '{"name":"Alice","age":30,"active":true}'
```

## Working with Nested Data

```haira
resp, err = http.get("https://api.weather.com/current")
data = resp.json()

// Navigate nested structures
temp = data["main"]["temp"]
description = data["weather"][0]["description"]
```
