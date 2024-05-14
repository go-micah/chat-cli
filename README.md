# 💬 chat-cli 💬

A little terminal based program that lets you interact with LLMs available via [Amazon Bedrock](https://aws.amazon.com/bedrock).

## Prerequisites

1. You will need an [AWS account](https://aws.amazon.com)
2. You will need to enable the LLMs you wish to use in Amazon Bedrock via the [Model Access](https://us-east-1.console.aws.amazon.com/bedrock/home?region=us-east-1#/modelaccess) page in the AWS Console. The defualt LLMs for both Chat and Prompt commands are proivded by Anthropic, so it is recommended to enable these as a starting point.
3. You will need to install the [AWS CLI](https://docs.aws.amazon.com/cli/) tool and run `aws config` from the command line to set up credentials.

## Local Quickstart

    git clone git@github.com:go-micah/chat-cli.git
    cd chat-cli
    go run . prompt "Explain this code" < cmd/prompt.go

Example output (this will change, as LLMs generate new responses with every request):

> This code is a Cobra-based command-line interface (CLI) application that allows users to send prompts to an Amazon Bedrock Large Language Model (LLM). The application supports several different LLM models, including Anthropic's Claude, AI21 Labs' Jurassic, Cohere's Command, Meta's LLaMA, and Amazon's Titan.
>
> The main functionality is in the `promptCmd` Cobra command, which has the following features:
>
> 1. **Prompt Handling**: The command can accept a prompt as a command-line argument, and it can also read a document from the standard input (STDIN) and prepend it to the prompt.
>    ...

## Installation

At this time you can install chat-cli via pre-packaged binaries (thanks to [GoReleaser](https://goreleaser.com/)!) for your operating system/architecture combination of choice.

1. Head to https://github.com/go-micah/chat-cli/releases/latest to find the binary for your setup.
2. Download and unzip to find a pre-compiled binary file that should work on your system.

Notes:

- You won't need Go installed on your system to use the pre-packaged binaries.
- These are currently unsigned binary files. For most systems, this will not be an issue, but on MacOS you will need to [follow these instructions](https://support.apple.com/guide/mac-help/open-a-mac-app-from-an-unidentified-developer-mh40616/mac).

## Build from source

You will need [Go](https://go.dev) v1.22.1 installed on your system. You can type `go version` to ensure you have the correct version installed.

To build the project from source, clone this repo to your local machine and use [Make](https://www.gnu.org/software/make/manual/make.html) to build the binary.

    $ git clone git@github.com:go-micah/chat-cli.git
    $ cd chat-cli
    $ make

## Run

To run the program from within the same directory use the following command syntax. (If you downloaded a pre-packaged binary your path will be different.)

    $ ./bin/chat-cli <command> <args> <flags>

## Help

You can get help at anytime with the `--help` flag. Typing `--help` after any command will display args and flags available to that command.

## Commands

There are currently three ways to interact with foundation models through this interface.

1. Send a single prompt to an LLM from the command line using the `prompt` command
2. Start an interactive chat with an LLM using the `chat` command
3. Generate an image with the `image` command

## Prompt

You can send a one liner prompt like this:

    $ ./bin/chat-cli prompt "How are you today?"

You can also read in a file from `stdin` as part of your prompt like this:

    $ cat myfile.go | ./bin/chat-cli prompt "explain this code"

    or

    $ ./bin/chat-cli prompt "explain this code" < myfile.go

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

| Provider  | Model ID                                | Family Name | Streaming Capable | Base Model |
| --------- | --------------------------------------- | ----------- | ----------------- | ---------- |
| Anthropic | anthropic.claude-3-haiku-20240307-v1:0  | claude3     | yes               | yes        |
| Anthropic | anthropic.claude-3-sonnet-20240229-v1:0 | claude3     | yes               | no         |
| Anthropic | anthropic.claude-v2:1                   | claude      | yes               |            |
| Anthropic | anthropic.claude-v2                     | claude      | yes               |            |
| Anthropic | anthropic.claude-instant-v1             | claude      | yes               | yes        |
| Cohere    | cohere.command-light-text-v14           | command     | yes               | yes        |
| Cohere    | cohere.command-text-v14                 | command     | yes               |            |
| Amazon    | amazon.titan-text-lite-v1               | titan       | not yet           | yes        |
| Amazon    | amazon.titan-text-express-v1            | titan       | not yet           |            |
| AI21 Labs | ai21.j2-mid-v1                          | jurassic    | no                | yes        |
| AI21 Labs | ai21.j2-ultra-v1                        | jurassic    | no                |            |
| Meta      | meta.llama2-13b-chat-v1                 | llama       | yes               | yes        |
| Meta      | meta.llama2-70b-chat-v1                 | llama       | yes               |            |

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

## Anthropic Claude 3 Vision

With the latest models from Anthropic, Claude 3 can now support uploading an image. Images can be either png or jpg and must be less than 5MB. To upload an image do the following:

    $ ./bin/chat-cli prompt "Explain this image" --image IMG_1234.JPG

Please note this only works with models from Anthropic Claude 3.

## Image

With the `image` command you can generate images with any supported Foundation Model. Simply follow the syntax below:

    $./bin/chat-cli image "Generate an image of a cat eating cereal"

You can specify the model with the `--model-id` flag set to model's full model id or family name.
You can also specify an output filename with the `--filename` flag.

## Image Models

| Provider     | Model ID                         | Family Name | Base Model |
| ------------ | -------------------------------- | ----------- | ---------- |
| Stability AI | stability.stable-diffusion-xl-v1 | stability   | yes        |
| Stability AI | stability.stable-diffusion-xl-v0 | stability   |            |
| Amazon       | amazon.titan-image-generator-v1  | titan-image | yes        |
