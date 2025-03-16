# The "local" environment represents local and pushes changes to atlas cloud registry
env "local" {
  url = "postgres://postgres:postgres@:5432/postgres?search_path=public&sslmode=disable"
  migration {
    dir = "atlas://movie-manager-bot"
  }
}

env "prod" {
  # url must be given from cli as --url flag
  migration {
    dir = "file://migrations"
  }
}