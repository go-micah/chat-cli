# chat-cli

A little terminal based program that lets you interact with LLMs available via [Amazon Bedrock](https://aws.amazon.com/bedrock).

## Prerequisites

1. You will need an AWS account
2. You will need to enable the LLMs you wish to use in Amazon Bedrock
3. You will need to run `aws config` from the command line to set up your default profile
4. You will need Go installed on your system

## Compile

    $ make

## Run

    $ ./bin/chat-cli <command> <args> <flags>

## Help

You can get help at anytime with the `--help` flag

## Use

There are currently two ways to interact with LLMs through this interface.

1. Send a single prompt from the command line using the `prompt` command
2. Start an interactive chat using the `chat` command

## Prompt

You can send a one liner prompt like this:

    $ ./bin/chat-cli prompt "How are you today?"

You can also pipe in a file as part of your prompt like this:

    $ cat myfile.go | ./bin/chat-cli prompt "explain this code"

This will add `<document></document>` tags arround your document ahead of your prompt. This syntax works especially well with [Anthropic Claude](https://www.anthropic.com/product). Other models may produce different results.

## Chat

You can start an interactive chat sessions which will remember your conversation as you chat back and forth with the LLM.

You can start an interactive chat session like this:

    $ ./bin/chat-cli chat

- Type `quit` to quit the interactive chat session.

## LLMs

Currently all text based LLMs available through Amazon Bedrock are supported. The LLMs you wish to use must be enabled within Amazon Bedrock. 

The default LLM is Anthropic Claude Instant v1. 

To switch LLMs, use the `--model-id` flag. You can supply a valid model id from the following list of currently supported models:

| Provider  | Model ID                      | Family Name | Streaming Capable | Base Model |
|-----------|-------------------------------|-------------|-------------------|------------|
| Anthropic | anthropic.claude-v2:1         | claude      | yes               |            |
| Anthropic | anthropic.claude-v2           | claude      | yes               |            |
| Anthropic | anthropic.claude-instant-v1   | claude      | yes               | yes        |
| Cohere    | cohere.command-light-text-v14 | command     | yes               | yes        |
| Cohere    | cohere.command-text-v14       | command     | yes               |            |
| Amazon    | amazon.titan-text-lite-v1     | titan       | not yet           | yes        |
| Amazon    | amazon.titan-text-express-v1  | titan       | not yet           |            |
| AI21 Labs | ai21.j2-mid-v1                | jurassic    | no                | yes        |
| AI21 Labs | ai21.j2-ultra-v1              | jurassic    | no                |            |
| Meta      | meta.llama2-13b-chat-v1       | llama       | yes               | yes        |
| Meta      | meta.llama2-70b-chat-v1       | llama       | yes               |            |



You can supply the exact model id from the list above like so:

    $ ./bin/chat-cli prompt "How are you today?" --model-id cohere.command-text-v14

Or, you can use the `Family Name` as a shortcut. Using the Family Name will select the `Base Model` as the least expensive option offered by each provider.

    $ ./bin/chat-cli prompt "How are you today?" --model-id titan

## Streaming Response

By default, responses will stream to the command line as they are generated. This can be dissabled using the `--no-stream` flag with the prompt command. Not all models offer a streaming response capability.

You can disable streaming like this:

    $ ./bin/chat-cli prompt "What is event driven architecture?" --no-stream

Only streaming response capable models can be used with the `chat` command. 

## Model Config

There are several flags you can use to overide the default config settings. Not all config settings are used by each model.

    --max-tokens defaults to 500
    --temperature defaults to 1
    --topP defaults to 0.999
    --topK defaults to 250
