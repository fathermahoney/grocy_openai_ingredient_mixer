# About
Want a shortcut to utilizing your fridge or pantry items? Don't mind if it's not 100% accurate? This is a tool to snag your Grocy inventory and have ChatGPT suggest meals based on the ingredients available. As with other ChatGPT observations, results are not always accurate, but are usually great conversation starters. A personality file is included to make the bot reply as familiar characters or evil monsters at random. Edit this file to experiment with different prompts.

*Note: Edit the prompt options to get replies in familiar voices.*

`personalities.json`
```
{
	"name":"norman_bates",
	"prompt":{
	  "season":["Spring", "Summer", "Fall", "Winter"],
	  "keywords":["psycho", "secluded", "loner", "desert", "knife"],
	  "OpenAI":"Address me as if you are Norman Bates from Psycho, and tell me about how great a cook your mother is. Go ever deeper into the psyche of a man who is obsessed with his mother. "
	}
  }
  ```
---
<img src="/images/norman_bates.png" alt="Alt text" title="Norman Bates">

---
## Requirements
* Grocy instance installed with Grocy API key. *See: [grocy.info](https://grocy.info/) for more info*
* Slack Account & Bot
* Slack Channel ID
* Go installed
---
## Quickstart
1. Make a copy of `SAMPLE.env` file as `.env` and fill in specifics
2. Adjust Personality prompts in `personalities.json`
3. `go run grocy_openai_ingredient_mixer.go`
---
### Future Me Problems
1. Suggest soon to be expired food first
2. Expand on a specific suggestion

