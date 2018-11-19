# Default Provider

This is the default auth provider and provides basic HTTP authentication using bearer tokens

Configuration:

|  Key              | Example                     | Required   | Description |
|-------------------|  -----------                | ----       |   -----------|
| `api-url`         | https://my.functions.com/   | Yes | The API endpoint to contact for accessing the service API |
| `call-url`        | https://my.functions.com/   | No |  The call endpoint  base URL for calling functions- this defaults to  `api-url` |
| `token`           | 0YHQtdC60YHRg9Cw0LvRjNC90YvQuSDQsdCw0L3QsNC9Cg== | No (Unless server requires authentication | The Bearer token to use for API auth |
