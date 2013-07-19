package main

import (
	"text/template"
)

var (
	TemplateSingleColumnTable = `
<table>
	<thead>
		<tr>
			<th></th>
	</thead>
	<tbody>
		{{range .}}
			<tr>
				<td>
					{{.message}}
				</td>
			</tr>
		{{end}}
	</tbody>
</table>
`

	TemplateDNATable = `
<table>
	<thead>
		<tr>
			<th></th>
	</thead>
	<tbody>
		{{range .}}
			<tr>
				<td>
					{{.name}}
				</td>
				<td>
					{{.dna}}
				</td>
			</tr>
		{{end}}
	</tbody>
</table>
`

	TemplateKawaiiTable = `
<table>
	<thead>
		<tr>
			<th></th>
	</thead>
	<tbody>
		{{range .}}
			<tr>
				<td>
					{{.name}}
				</td>
				<td>
					{{.face}}
				</td>
			</tr>
		{{end}}
	</tbody>
</table>
`
	TemplateSkeleton = `
<!doctype html>
<html>
	<head>
		<style type="text/css"> 
			a:visited{
				color:blue;
			}
			ul.nav li {
				display:inline;
			}
		</style>
	</head>
	<body>
		Guest: 
		<ul class="nav">
				<li><a href="/kawaii">Kawaiiface</a></li>
				<li><a href="/dna">DNA</a></li>
				<li><a href="/joke">Joke</a></li>
				<li><a href="/rape">Rape</a></li>
				<li><a href="/kill">Kill</a></li>
				<li><a href="/factoid">Factoid</a></li>
				<li><a href="/yomama">Yomama</a></li>
				<li><a href="/Gelbooru/Random">Gelbooru Random</a></li>
				<li><a href="/Files">Files</a></li>
				<li><a href="/Auth">Login</a></li>
				
				<!--<li><a href="/Irc">Irc</a></li>
				<li><a href="/Irc/Latest">IrcLog</a></li>--> 
		</ul>
		
		Admin: 
		<ul class="nav">
			<li><a href="/FileUpload">File Upload</a></li>
			<li><a href="/CreateUser">CreateUser</a></li>
			<li><a href="/plugins">Plugin Manager</a></li>
		</ul>
	
		{{template "content" .}}
	</body>
</html>`

	TemplatePluginlist = `
<h1> Plugin hub </h1>
<table border="1">
	<thead>
		<tr>
			<th>File Name</th>
			<th></th>
			<th>Functions</th>
			<th>Actions</th>
		</tr>
	</thead>
	<tbody>
		{{range .}}
		<tr>
			<td>{{.name}}</td>
			<td>{{if .file}}File Exists{{end}} {{if .loaded}}Loaded{{end}}</td>
			<td>{{range .functions}}{{.}}, {{end}}</td>
			<td>
				{{if .file}}<a href="/plugins/edit?name={{.name}}">Edit</a> 
				<a href="/plugins/delete?name={{.name}}">Delete</a>
				<a href="/plugins/load?name={{.name}}">(Re-)load</a>{{end}}
				{{if .loaded}}
				<a href="/plugins/unload?name={{.name}}">Unload</a>
				{{end}}
			</td>
		</tr>
		{{end}}
	</tbody>
</table>

Create new:

<form action="/plugins/edit" method="GET">
	<input type="text" name="name" /> <br />
	<input type="submit" value="Create" />
</form>
`

	TemplatePluginedit = `
<script type="text/javascript" src="/scripts/ace/ace.js"></script>
	
<h1> Edit plugin {{.name}} </h1>
<form action="/plugins/edit" method="POST" onsubmit="document.getElementById('contentT').value = ace.edit('content').getSession().getValue();">
	<input type="hidden" value="{{.name}}" name="name" />
	<textarea cols="100" rows="40" name="content" id="contentT">{{.content}}</textarea> <br />
	<div id="content" style="height: 550px; width: 900px;"></div> <br />
	Save: <input type="text" name="filename" value="{{.name}}" /><input type="submit" value="Submit" />
</form>

<script type="text/javascript"> 
	var editor = ace.edit("content");
	var textarea = document.getElementById("contentT");
    editor.setTheme("ace/theme/textmate");
    editor.getSession().setMode("ace/mode/javascript");
	textarea.style.display = "none";
	editor.getSession().setValue(textarea.value);
</script>

<h2> API: </h2>

	<p>
		<h3>Defining a function:</h3>
		use <pre> Subscribe(string trigger, function(sender, source, line, params) handler) </pre> to define a function. <br />
		The following trigger types are defined:
		
		<ul>
			<li>!command - These are your bog standard irc commands</li>
			<li>privmsg - This will trigger on all messages received</li>
			<li>join - Triggers if someone joins the channel</li>
			<li>part - Triggers when someone leaves the channel</li>
			<li>quit - Triggers when someone disconnects from irc</li>
			<li>regex: - Type a regex on which it will trigger after the colon (:)</li>
		</ul>
		
		The handler function has the following parameters:
		
		<ul>
			<li>sender - This is the nick name of the sender</li>
			<li>source - This is either the nick name of the sender on a direct message or the channel the sender posted in</li>
			<li>line - Either the complete message or the part after the !command</li>
			<li>params - In !commands this is a array of space delimited parameters. Parameters with a space in it can be consolidated with quotes. <br />
				<pre>!command param param21 "pa sdf asdfj dflk"</pre>
				is THREE parameters. 
				<ol>
					<li><pre>param</pre></li>
					<li><pre>param21</pre>and</li>
					<li><pre>pa sdf asdfj dflk</pre></li>
				<ol>
				Actual quotes can be escaped using \. Matching quotes will otherwise be removed.
			</li>
		</ul>	
			
	</p>

	<p>
		<h3>Function definitions</h3>
		
		<ul>
			<li>IRC
				<ul>
					<li>IRC.Privmsg(destination, message) - sends an IRC message to a channel or a nick</li>
					<li>IRC.Action(channel, message) - sends an "action" to a IRC channel</li>
					<li>IRC.Notice(nick, message) - sends a Notice towards a user</li>
					<li>IRC.Channel - Contains the main channel of the bot</li>
					<li>IRC.Server - Contains the current server name</li>
					<li>IRC.Nick - Contains the nick of the bot</li>
				</ul>
			</li>
			<li>
				DB
				<ul>
					<li>DB.Authenticate(user, pwd) - Authenticates against the user database</li>
					<li>DB.SaveToDb</li>
					<li>DB.ExistsInDB</li>
					<li>DB.GetRandomFromDB</li>
					<li>DB.GetAndDeleteFirstFromDB</li>
					<li>DB.GetNamedFromDB</li>
					<li>DB.SaveNamedToDB</li>
				</ul>
			</li>
		</ul>
	</p>
`

	TemplateLogin = `
	{{if .}} Error: {{.}} {{end}}
	<form method="POST">
		User: <input type="text" name="username" /> <br /> 
		Password: <input type="password" name="password" /> <br /> 
		<input type="submit" name="submit" value="Login" /> 
	</form>
`

	TemplateCreateUser = `
	<form method="POST" action="/CreateUser">
		Username: <input type="text" name="username" /> <br /> 
		Password: <input type="text" name="password" /> <br /> 
		<input type="submit" name="submit" value="Create" /> 
	</form>
`

	TemplateFiles = `
`

	TemplateFileUpload = `

`
)

func Template(subtmpl string) *template.Template {
	templ := template.New("main")
	templ.Parse(TemplateSkeleton)
	contenttmpl := templ.New("content")
	if subtmpl != "" {
		contenttmpl.Parse(subtmpl)
	}
	return templ
}
