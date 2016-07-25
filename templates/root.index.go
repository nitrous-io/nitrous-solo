package templates

import "strings"

var rootIndex = strings.TrimSpace(`
{{ define "root.index" }}
<!DOCTYPE html>
<html lang='en'>
  <head>
    <meta charset='utf-8'>
    <title>Nitrous Solo</title>
		<meta http-equiv="X-UA-Compatible" content="IE=edge">
		<link rel="stylesheet" href="/assets/app.css" />
  </head>
  <body>
		<div id="webapp-loading">
			<div class="ui progress">
				<div class="bar"></div>
			</div>
		</div>
		<div id="app"></div>
    <script>
      this.NitrousConfig = {{ .config }};
    </script>
    <script src="/assets/bundle.js"></script>
  </body>
</html>
{{ end }}
`)
