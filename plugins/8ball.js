Doc("!8ball", "Try the mighty eightball.")
Subscribe("!8ball", function(source, commandline) {
	var possibilities = ["Signs point to yes. ", "Yes. ", "Reply hazy, try again. ", 
                        "Without a doubt. ", "My sources say no. ", "As I see it, yes. ", 
                        "You may rely on it. ", "Concentrate and ask again. ", 
                        "Outlook not so good. ", "It is decidedly so. ", 
                        "Better not tell you now. ", "Very doubtful. ", "Yes - definitely. ", 
                        "It is certain. ", "Cannot predict now.", "Most likely.", 
                        "Ask again later. ", "My reply is no.", "Outlook good.", "Don't count on it.",
                        "For this question I demand a boot to the head!", "For this question I demand a boot to the head!", 
                        "For this question I demand a boot to the head!", "For this question I demand a boot to the head!",
                        "For this question I demand a boot to the head!1"];
	
	var rand = Math.floor((Math.random()*possibilities.length))

	IRC.Privmsg(source.Source, possibilities[rand])

})

