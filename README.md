Jira TUI
===

very much WIP

View jira issues from the terminal


Usage
===

```bash
make build
PAT="access-token" JIRA=instance-domain ./jira-cli ISSUE-12345
```

Key bindings
===

1. Search mode (starts there)
	* `enter`: finds and display issue
	* `tab`: go to issue body
2. Issue body:
  * `up`/`down` arrows: move
  * `pgup`/`pgdown`: move
 	* `tab`: go to meta block
 	* `escape`: go to search mode
3. Meta block:
  * `up`/`down` arrows: move
  * `pgup`/`pgdown`: move
 	* `tab`: go to meta block
 	* `escape`: go to search mode
