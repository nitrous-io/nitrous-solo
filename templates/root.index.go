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
  </head>
  <body>
    <script>
      this.NitrousConfig = {{ .config }};
    </script>
    <script src="/assets/bundle.js"></script>
  </body>
</html>
{{ end }}
`)
