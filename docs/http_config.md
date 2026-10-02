This is the HTTP request the client will do first upon logging in.

## Request
```http
GET /maxim_settings_us.asp?Version=1.00&BuildID=8039&Region=AMERICAN&Keyboard=1 HTTP/1.0
Host: [www.datelversions.com](https://www.datelversions.com)
```

| Parameter  | Value Observed | Data Type |
|:-----------|:---------------|:----------|
| `Version`  | `1.00`         | String    |
| `BuildID`  | `8039`         | Integer   |
| `Region`   | `AMERICAN`     | String    |
| `Keyboard` | `1` / `0`      | Boolean   |

---

## Response

The client expects a `HTTP/1.0 200 OK` response with a few config keys in the response body. Likely separated by `\r\n`.

| Key           | Format          | Example Value          | Description                                    |
|:--------------|:----------------|:-----------------------|:-----------------------------------------------|
| `chat_server` | IPv4 / Hostname | `192.168.137.1`        | IP of the chat server                          |
| `chat_port`   | Integer         | `2002`                 | Port of the chat server (doesn't seem to work) |
| `advert_url`  | HTTP URL        | `http://192.168.137.1` | ?                                              |


```http
HTTP/1.0 200 OK
Content-Type: text/plain
Content-Length: 72
Connection: close

chat_server=192.168.137.1
chat_port=2002
advert_url=http://192.168.137.1
```
