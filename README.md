# Function calling

## Overview
This project demonstrates the implementation of AI function calling using the OpenAI-compatible API interface. It allows you to create tools that can be invoked by an AI model during conversation, enabling the AI to perform specific tasks such as retrieving real-time weather information.

## Features
- Integration with OpenAI-compatible APIs (DashScope)
- Implementation of function calling with AI models
- Weather lookup tool for retrieving current weather and forecast information
- Conversation history management
- Multiple tool invocation in a single conversation

## Prerequisites
- Go (1.18 or higher)
- OpenAI-compatible API access (e.g., DashScope API key)

## Setup

1. **Clone the repository**
   ```bash
   git clone https://github.com/yourusername/function-calling.git
   cd function-calling
   ```

2. **Install dependencies**
   ```bash
   go mod download
   ```

3. **Configure environment variables**
   Copy the `.env.template` file to `.env` and update with your API credentials:
   ```bash
   cp .env.template .env
   ```
   Then modify the `.env` file with your specific values:
   ```
   DASH_SCOPE_API_KEY="your-api-key-here"
   DASH_SCOPE_URL="https://dashscope.aliyuncs.com/compatible-mode/v1"
   DASH_SCOPE_MODEL="qwen-turbo"
   ```

## Usage

Run the application:
```bash
go run main.go
```

The example in `main.go` demonstrates how to:
1. Define a weather tool
2. Send a user query to the AI
3. Process the AI's decision to call a function
4. Execute the function and return the results
5. Generate a final response incorporating the function results

## Example

The provided example asks for the current weather in Shenzhen and whether it's suitable for outdoor activities. The AI:
1. Recognizes that weather information is needed
2. Calls the GetWeather function with appropriate parameters
3. Receives the weather data
4. Generates a response based on the weather conditions

## Available Tools

### Weather Tool
- Retrieves current weather and forecast information from the OpenMeteo API
- Can request specific weather parameters (temperature, humidity, wind speed, etc.)
- Supports daily forecasts

## Extending with New Tools

To add new tools:

1. Create a new file in the `tools` directory
2. Define your tool function and its OpenAI Tool definition
3. Update the `toolsList` in `main.go` to include your new tool

## License

Apache License 2.0

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.