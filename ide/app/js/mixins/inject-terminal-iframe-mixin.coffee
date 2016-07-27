App.InjectTerminalIframeMixin = Em.Mixin.create
  terminal: null

  injectTerminalIframe: (callback) ->
    $iframe = $('<iframe>')
    @$('.terminal-screen').append($iframe)
    iframeWindow = $iframe[0].contentWindow
    iframeDocument = iframeWindow.document

    # hack to make iframe ready for injection in firefox
    iframeDocument.open()
    iframeDocument.close()

    cssPath = "/assets/terminal.css"
    jsPath = "/assets/terminal.js"

    origin = location.protocol + "//" + location.host

    if !(cssPath.match("^//"))
      cssPath = origin + cssPath
    if !(jsPath.match("^//"))
      jsPath = origin + jsPath

    # use pure js dom api because jquery fails to use iframe's window/document obj
    linkEl = iframeDocument.createElement('link')
    linkEl.setAttribute('rel', 'stylesheet')
    linkEl.setAttribute('href', cssPath)
    iframeDocument.head.appendChild(linkEl)

    scriptEl = iframeDocument.createElement('script')
    scriptEl.setAttribute('src', jsPath)

    scriptEl.onload = =>
      scriptEl.onload = null
      terminal = iframeWindow.NitrousTerminal
      terminal.init()
      terminal.focus()
      @set 'terminal', terminal
      callback?(terminal)

    iframeDocument.body.appendChild(scriptEl)
