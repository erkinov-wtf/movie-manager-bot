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
	PrivacyPolicy     = "By using this bot, you agree to our [Privacy Policy](https://example.com/privacy-policy)"
	UseHelp           = "use /help for assistance"
	Registered        = "You have been successfully registered.\nBut for using this bot you need to get TMDB Token.\nPlease click */token* to get you token"
	TokenInstructions = "To use this bot, you need a TMDB API Key. Please follow these steps:\n" +
		"1. Visit the [TMDB Account Page](https://www.themoviedb.org/account/signup) and create an account if you don’t already have one.\n" +
		"2. After signing in, go to the [API Settings](https://www.themoviedb.org/settings/api/new?type=developer).\n" +
		"3. Read the *Terms of Use* and click the *Accept* button at the bottom.\n" +
		"4. Follow the instructions and fill out the required form as needed.\n" +
		"5. Once you have your *API Key* (not the API Read Access Token), return to this bot and paste the key here to test and save it.\n\n" +
		"If you’d prefer assistance, feel free to contact [me](https://t.me/erkinov_wiz), and I'll help you get your key."
	TokenTestFailed         = "Your token failed the test. Ensure it is correct and write/paste it here again."
	TokenSaved              = "Your token has been successfully tested and saved. From now on, this token will be used exclusively for you in this bot. \nPlease use the /help command to get started."
	TokenAlreadyExists      = "Your Api Token has been already saved, no need to worry"
	Loading                 = "Loading..."
	InfoFirstMessage        = "What you want info about?"
	MovieSelected           = "Selected the movie!"
	TVShowSelected          = "Selected the TV show!"
	WatchedMovie            = "You have already watched this movie"
	NoSearchResult          = "No search results found"
	BackToSearchResults     = "Returning to search results"
	WatchedSeason           = "You already watched this season, please select later seasons"
	WatchlistSelectType     = "Which type of watchlist do you want?"
	NoWatchlistData         = "No records found"
	NoChanges               = "No changes to display"
	PageUpdated             = "Page updated"
	RegistrationRequired    = "You need to register to use this bot. Please type /start to continue"
	MenuSearchTVResponse    = "Write TV Show Title"
	MenuSearchMovieResponse = "Write Movie Title"
)

const (
	InternalError       = "Something went wrong, please try again"
	MalformedData       = "Malformed data received"
	UnknownAction       = "Unknown action"
	InvalidSeason       = "Invalid season number received"
	InvalidPageNumber   = "Invalid page number"
	WatchlistCheckError = "Something went wrong while checking your watchlist."
)
