# Konkurransetilsynet
Discord bot for the elite of Norwegian olympiad participants, built with Golang and passion.

## Features
- Codeforces integration.
- Guess the Function game.
- Database storage using PostgreSQL. Proper migration system will be implemented soon.

See [Commands](#Commands) for more details.

## Commands
### Codeforces
These commands are related to the competitive programming platform [Codeforces](https://codeforces.com/).

To access these commands prefix the command with `!cf`.
- List upcoming contests. `contests`
- Authentication by submitting a compilation error to a randomly selected problem. `authenticate [your codeforces username]`
- Automatically sends leaderboard with every authenticated member of the Discord server when ratings are updated after a contest.
- Automatically sends contest reminders an hour before a contest starts.

### Guess the Function™
To access Guess the Function commands use the prefix `!gtf`
- Start a round. `start [lower bound] [upper bound] [function definition]`
- Query the current function. `query [value]`
- Guess the current function. `guess [function definition]`
- Parses any function limited to addition, subtraction, multiplication, division, exponents, parenthesis and the variable x.
- Evaluates numerically, meaning guessing `f(x) = 1` for the function `f(x) = x/x` is valid.  
#### Limitations
- The notation `10x` is not accepted, `10*x` is. This includes `10(...)` which should be `10*(...)`

## Tech Stack
- Golang - programming language.
- [discordgo](https://github.com/bwmarrin/discordgo) - Discord API bindings.
- PostgreSQL - database storage.
