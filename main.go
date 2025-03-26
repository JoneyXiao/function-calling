package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	tools "function-calling/tools"

	godotenv "github.com/joho/godotenv"
	openai "github.com/sashabaranov/go-openai"
)

// ChatMessages represents a collection of chat messages for OpenAI API
type ChatMessages []openai.ChatCompletionMessage

// MessageStore holds the conversation history for the current session
var MessageStore ChatMessages

// Constants for message roles to improve code readability and avoid typos
const (
	RoleSystem    = "system"
	RoleUser      = "user"
	RoleAssistant = "assistant"
	RoleTool      = "tool"
)

func (cm *ChatMessages) Clear() {
	*cm = make([]openai.ChatCompletionMessage, 0)
}

// init loads environment variables from .env file
func init() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	MessageStore = make(ChatMessages, 0)
	MessageStore.Clear()
}

func NewOpenAiClient() *openai.Client {
	apiKey := os.Getenv("DASH_SCOPE_API_KEY")
	config := openai.DefaultConfig(apiKey)
	config.BaseURL = os.Getenv("DASH_SCOPE_URL")

	return openai.NewClientWithConfig(config)
}

// AppendMessage adds a new message to the chat history with specified role and content
func (cm *ChatMessages) AppendMessage(role string, msg string, toolCalls []openai.ToolCall) {
	*cm = append(*cm, openai.ChatCompletionMessage{
		Role:      role,
		Content:   msg,
		ToolCalls: toolCalls,
	})
}

// GetMessages returns a copy of all messages in the chat history
// This creates a new slice to avoid modifying the original messages
func (cm *ChatMessages) GetMessages() []openai.ChatCompletionMessage {
	ret := make([]openai.ChatCompletionMessage, len(*cm))
	copy(ret, *cm)
	return ret
}

// AddTool adds a tool to the chat history
func (cm *ChatMessages) AddTool(msg string, name string, CallId string) {
	*cm = append(*cm, openai.ChatCompletionMessage{
		Role:       RoleTool,
		Content:    msg,
		Name:       name,
		ToolCallID: CallId,
	})
}

// Chat sends the message history to the AI model and returns the response
// It handles any errors that might occur during the API call
func Chat(message []openai.ChatCompletionMessage) openai.ChatCompletionMessage {
	// Create a new OpenAI client for this request
	client := NewOpenAiClient()

	// Send the chat completion request to the API
	rsp, err := client.CreateChatCompletion(context.TODO(), openai.ChatCompletionRequest{
		Model:    os.Getenv("DASH_SCOPE_MODEL"),
		Messages: message,
	})
	if err != nil {
		log.Println(err)
		return openai.ChatCompletionMessage{}
	}

	// Return the first (and typically only) message from the choices
	return rsp.Choices[0].Message
}

// ChatWithTools sends the message history to the AI model with tools and returns the response
// It handles any errors that might occur during the API call
func ChatWithTools(message []openai.ChatCompletionMessage, tools []openai.Tool) openai.ChatCompletionMessage {
	// Create a new OpenAI client for this request
	client := NewOpenAiClient()

	// Send the chat completion request to the API
	rsp, err := client.CreateChatCompletion(context.TODO(), openai.ChatCompletionRequest{
		Model:      os.Getenv("DASH_SCOPE_MODEL"),
		Messages:   message,
		Tools:      tools,
		ToolChoice: "auto",
	})

	if err != nil {
		log.Println(err)
		return openai.ChatCompletionMessage{}
	}

	// Return the first (and typically only) message from the choices
	return rsp.Choices[0].Message
}

// printDebugInfo prints debug information about the message store
func printDebugInfo() {
	fmt.Println("# Message Store Debug:")
	messages := MessageStore.GetMessages()
	fmt.Printf("Number of messages: %d\n", len(messages))
	for i, msg := range messages {
		fmt.Printf("Message %d: Role=%s, Content length=%d\n", i, msg.Role, len(msg.Content))
	}
}

func main() {
	// Example with chat history
	// MessageStore.AppendMessage(RoleSystem, "你是一名 AIOps 专家，请尽可能地帮我回答与 AIOps 相关的问题。")
	// MessageStore.AppendMessage(RoleUser, "AIOps 是什么？")
	// MessageStore.AppendMessage(RoleAssistant, "AIOps 是 AI 和 Ops 的结合，通过 AI 技术帮助运维人员更好地管理 IT 基础设施。")
	// MessageStore.AppendMessage(RoleUser, "它的典型应用场景有哪些？")

	// response := Chat(MessageStore.GetMessages())
	// fmt.Println(response.Content)

	toolsList := make([]openai.Tool, 0)
	toolsList = append(toolsList, tools.WeatherToolDefine)

	// MessageStore.AppendMessage(RoleSystem, "You are a weather expert, please help me answer questions about weather.", nil)
	prompt := "What's the weather in Shenzhen? Is it suitable for outdoor activities?"
	// prompt := "帮我查询一下深圳当前的天气情况，今天适合出去游玩吗？ Let's think step by step."
	MessageStore.AppendMessage(RoleUser, prompt, nil)

	response := ChatWithTools(MessageStore.GetMessages(), toolsList)
	toolCalls := response.ToolCalls

	maxLoops := 5
	loopCount := 0

	for {
		fmt.Printf("-------------- The %d round response ------------------\n", loopCount)
		printDebugInfo()

		if toolCalls == nil || loopCount >= maxLoops {
			fmt.Println("Final response from LLM: ", response.Content)
			break
		} else {
			fmt.Println("Response from LLM: ", response.Content)
			fmt.Println("Selected Tool by LLM: ", toolCalls)
			fmt.Println("Tool Call Arguments: ", toolCalls[0].Function.Arguments)

			var result string
			var args tools.WeatherParams
			err := json.Unmarshal([]byte(toolCalls[0].Function.Arguments), &args)
			if err != nil {
				log.Fatalln("Failed to unmarshal tool call arguments: ", err.Error())
				return
			}

			// Call the tool
			if toolCalls[0].Function.Name == tools.WeatherToolDefine.Function.Name {
				result, err = tools.GetWeather(args)
				if err != nil {
					log.Fatalln("Failed to get weather: ", err.Error())
					return
				}
			}

			fmt.Println("Result from tool: \n", result)

			// Append the response from LLM and the tool call result to the message history
			MessageStore.AppendMessage(RoleAssistant, response.Content, toolCalls)
			MessageStore.AddTool(string(result), toolCalls[0].Function.Name, toolCalls[0].ID)

			response = ChatWithTools(MessageStore.GetMessages(), toolsList)
			toolCalls = response.ToolCalls
			loopCount++
		}
	}
}
