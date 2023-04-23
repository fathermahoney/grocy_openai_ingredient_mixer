# About
A tool to snag your Grocy inventory and have ChatGPT suggest meals based on the ingredients available. A personality file is included to make the bot reply as familiar characters or evil monsters at random.
## Requirements
* Grocy instance installed with Grocy API key
* Slack Account & Bot
* Slack Channel ID
* Go installed
---
## Quickstart
1. Make a copy of `SAMPLE.env` file as `.env` and fill in specifics
2. Adjust Personality prompts in `personalities.json`
3. `go run grocy_openai_ingredient_mixer.go`

