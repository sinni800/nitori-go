Doc("!dice <xdy>", "Rolls x dice with y sides")

Subscribe("!dice", function(source, line, params) {
	var out = ""
	var done = false
	if (params.length == 1) {
		var spl = params[0].split("d")

		if (spl.length == 2) {
			
			if (spl[0] < 15 && spl[1] < 10000) {
				out = "Rolling " + spl[0] + " " + spl[1] + "-sided dice: "
				for (var x = 0; x < spl[0]; x++) {
					out += Math.floor((Math.random()*spl[1])+1) + ", "
				}
				done = true
			}
		}
	} 
	
	if (!done) {
		out = "Usage: !dice 1d6   where 1 is the amount of dice and 6 is the highest possible number. Amount of dice < 15; Maximum number < 10000"
	}
	
	IRC.Privmsg(source.Source, out)
})