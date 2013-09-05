Subscribe("!ping", function(source) {
    IRC.Privmsg(source.Source, "Pong!")
})