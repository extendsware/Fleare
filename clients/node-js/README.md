# Fleare Node Js Documentation

The `FleareClient` is a TypeScript/JavaScript library for interacting with an Fleare server. It provides methods to manage JSON data using string-based keys and optional or multi-level paths.

---

## Installation

To install the library, use npm:

```bash
npm install fleare
```

---

## Usage

First, import the `FleareClient` class and create an instance:

```javascript
const { FleareClient } = require("fleare");

const fleare = new FleareClient("localhost", 1234, {
  username: "user",
  password: "pass",
});
```

---

## API Reference

### `connect()`
Connects to the server and authenticates the fleare.

**Returns**: `Promise<void>`

**Example**:
```javascript
fleare.connect()
  .then(() => {
    console.log("Connected and authenticated");
  })
  .catch((err) => {
    console.error("Connection failed:", err);
  });
```

---

### `close()`
Closes the connection to the server.

**Example**:
```javascript
fleare.close();
console.log("Connection closed");
```

---

### `KEYS()`
Retrieves all keys stored on the server.

**Returns**: `Promise<string[]>`

**Example**:
```javascript
fleare.KEYS()
  .then((keys) => {
    console.log("Keys:", keys);
  })
  .catch((err) => {
    console.error("Failed to retrieve keys:", err);
  });
```

**Output**:
```json
["user", "config", "data"]
```
---

### `SET(key, data)`
Sets a key-value pair on the server.

**Example 1**:
```javascript
fleare.SET("str-1", "This is my first key")
  .then(() => {
    console.log("Key set successfully");
  })
  .catch((err) => {
    console.error("Failed to set key:", err);
  });
```

**Example 2**:
```javascript
fleare.SET("user", { name: "Alice", age: 25 })
  .then(() => {
    console.log("Key set successfully");
  })
  .catch((err) => {
    console.error("Failed to set key:", err);
  });
```
---

### `GET(key)`
Retrieves the value associated with a key.

**Example**:
```javascript
fleare.GET("user")
  .then((value) => {
    console.log("Value:", value);
  })
  .catch((err) => {
    console.error("Failed to retrieve value:", err);
  });
```
**Output**:
```json
{
  name: "Alice",
  age: 25
}
```
---

### `DELETE(key, path?)`
Deletes a key or a specific path inside a JSON object.

**Example Output**:
```javascript
fleare.DELETE("user", "age")
  .then(() => {
    console.log("Deleted key/path successfully");
  })
  .catch((err) => {
    console.error("Failed to delete key/path:", err);
  });
```

---

### `UPDATE(key, path?, data)`
Updates a value inside a JSON object.

**Example Output**:
```javascript
fleare.UPDATE("user", "name", "Bob")
  .then(() => {
    console.log("Value updated successfully");
  })
  .catch((err) => {
    console.error("Failed to update value:", err);
  });
```

---

### `MAP_SET(key, path?, data)`
Sets a value inside a JSON map.

**Example Output**:
```javascript
fleare.MAP_SET("config", "theme.color", "dark")
  .then(() => {
    console.log("Map value set");
  })
  .catch((err) => {
    console.error("Failed to set map value:", err);
  });
```

---

### `MAP_GET(key, path?)`
Retrieves a value from a JSON map.

**Example Output**:
```javascript
fleare.MAP_GET("config", "theme.color")
  .then((value) => {
    console.log("Map value:", value); // Output: "dark"
  })
  .catch((err) => {
    console.error("Failed to retrieve map value:", err);
  });
```

---

### `MAP_DELETE(key, path?)`
Deletes a value from a JSON map.

**Example Output**:
```javascript
fleare.MAP_DELETE("config", "theme.color")
  .then(() => {
    console.log("Value deleted from map");
  })
  .catch((err) => {
    console.error("Failed to delete value from map:", err);
  });
```

---

This documentation provides methods for managing JSON data efficiently using string keys and optional or multi-level paths. Happy coding!
