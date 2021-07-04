# Reddit Comments Scrapper

This is web server to get comments from choosed subreddits.

## Api
### `GET`/start_observing/<subreddit_name> 
Preload some posts (defined in config file) and start getting updates on new comments in selected subreddit

### `GET`/stop_observing/<subreddit_name> 
Removes subreddit from observable ones. Change applies only after server reload.

## Models

### `Comment`
| Key  | Description | Type | Modifiers |
| ------------- | ------------- | ------------- | ------------- |
| FullName  | Full id of the comment | string | Primary Key |
| Id  | Id of the comment without t1_ prefix | string | |
| Author | Nickname of the author | string | |
| Body | Text of the comment | string | |
| Score | Diff between upvotes and downvotes | int32 | |
| SubredditID | The id of the subreddit in which the thing is located | string | |
| ParentID | Id of the thing this comment is a reply to, either the link or a comment in it | string | |

### Scrapper Agent
- UserAgent: Any string
- ClientId and ClientSecret: Go to https://ssl.reddit.com/prefs/apps/ and create new application (Choose script type). Copy this fields from created application.
- Username and Password: Your username and password on reddit.
