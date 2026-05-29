// Standalone module — keeps the deploy tool's build independent of the main
// BuddieBot module's vendor/replace directives, so building it on the server
// is `cd scripts/pull-deploy && go build .` with no other setup.
module github.com/Beamer64/BuddieBot/scripts/pull-deploy

go 1.26
