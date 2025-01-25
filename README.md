# Running Man Video Store Bot

This is a Telegram bot that enables users to buy **Running Man** videos using Telegram stars. Once purchased, the videos are permanently available for the users to watch as many times as they like without needing to repurchase.

## Features

- **Purchase Videos**: Users can purchase their favorite **Running Man** episodes through Telegram stars.
- **Permanent Access**: Videos bought by users are permanently available in their accounts.
- **Unlimited Viewing**: Users can watch purchased videos as many times as they wish.

## Technology Stack

- **Programming Language**: [Go](https://golang.org/) - Known for its simplicity and high performance.
- **Telegram Bot API**: Facilitates bot communication and user interactions.

## How It Works

1. **User Registration**: When a user starts the bot `@RunningManSeriesBot`, their Telegram ID is registered.
2. **Browse Videos**: Users can view a list of available **Running Man** videos through `/browse` command.
3. **Buy Videos**: Users can purchase videos using Telegram stars.
4. **Watch Videos**: Once purchased, videos are permanently added to the user's collection for unlimited viewing and can be accessed through `/collection` command.
