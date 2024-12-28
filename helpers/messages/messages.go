package messages

const (
	StartCommand       = "/start command received"
	InfoCommand        = "/info command received"
	MovieCommand       = "/sm command received"
	TVShowCommand      = "/stv command received"
	WatchlistCommand   = "/w command received"
	MovieEmptyPayload  = "After /sm, a movie title must be provided. Example: /sm Deadpool"
	TVShowEmptyPayload = "After /stv, a movie title must be provided. Example: /sm Supernatural"
)

const (
	PrivacyPolicy       = "By using this bot, you agree to our [Privacy Policy](https://example.com/privacy-policy)"
	UseHelp             = "use /help for assistance"
	Registered          = "You have been successfully registered.\nNow you can use this bot. Use /help for assistance)"
	Loading             = "Loading..."
	InfoFirstMessage    = "What you want info about?"
	MovieSelected       = "Selected the movie!"
	TVShowSelected      = "Selected the TV show!"
	WatchedMovie        = "You have already watched this movie"
	NoSearchResult      = "No search results found"
	BackToSearchResults = "Returning to search results"
	WatchedSeason       = "You already watched this season, please select later seasons"
	WatchlistSelectType = "Which type of watchlist do you want?"
	NoWatchlistData     = "No records found"
	NoChanges           = "No changes to display"
	PageUpdated         = "Page updated"
)

const (
	InternalError       = "Something went wrong, please try again"
	MalformedData       = "Malformed data received"
	UnknownAction       = "Unknown action"
	InvalidSeason       = "Invalid season number received"
	InvalidPageNumber   = "Invalid page number"
	WatchlistCheckError = "Something went wrong while checking your watchlist."
)
