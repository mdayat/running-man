# Running Man Bot

This is a Telegram bot that enables users to subscribe to access **Running Man** videos. Subscribers can stream videos unlimited times during their active subscription period.

## Features

- **Subscription Access**: Users subscribe to watch videos as many times as they want as long as their subscription remains active.
- **Stream Videos**: Enjoy smooth streaming of **Running Man** episodes.
- **Easy Navigation**: Browse available videos using the `/browse` command.

## Technology Stack

- **Programming Language**: [Go](https://golang.org/)
- **Telegram Bot API Go framework**: https://github.com/go-telegram/bot
- **Payment Gateway**: [Tripay](https://tripay.co.id) - Handles subscription payments securely.
- **Webhook Integration**: Processes payment confirmation events from Tripay to activate subscriptions.

> **Developer Note:**  
> Before running the project, ensure you have installed the following code analysis tools:
>
> - `govulncheck`
> - `staticcheck`
> - `revive`
>
> These tools are required to execute the code checks script defined in the `Makefile`.  
> Additionally, developers must provide the environment variables as described in the `.example.env` file for **each service**.

## How It Works

1. **Browse Videos**: Users can view a list of available **Running Man** videos using the `/browse` command.
2. **Watch or Subscribe**:
   - If a user already has an active subscription, they can select and stream the video immediately.
   - If not, the user will be prompted to subscribe before streaming the video.
3. **Subscription Management**: Once subscribed, users gain access to stream any video until their subscription expires.

## Services

1. **Telegram Bot**: The primary service that interacts with users and facilitates video streaming.
2. **Webhook Service**: Listens for Tripay payment confirmation events to activate subscriptions.
