# chat-cli

A little terminal based program that talks to the Amazon Bedrock API.

![](images/claude-1.png)
![](images/claude-2.png)

## Prerequisites

1. You will need an AWS account.
2. You will need to enable Claude v2 in Amazon Bedrock
3. You will need to run `aws config` from the command line to set up your default profile

## Compile

    $ make cli

## Run

    $ ./bin/chat-cli

## Use

You can send messages to Claude in natural language. If the response is lengthy it may take a little while to appear. The responses are currently capped at 300 tokens, so things may get truncated if Claude gets long winded.

You can send a one liner prompt like this:

    $ ./bin/chat-cli prompt "How are you today?"

## Chat commands

You can start an interactive chat session like this:

    $ ./bin/chat-cli chat

- Type `quit` or `ctl-c` to quit the interactive chat session.
- Type `save` to save your current conversation to a file.
- Type `load` to load a previously saved conversation from a file.
