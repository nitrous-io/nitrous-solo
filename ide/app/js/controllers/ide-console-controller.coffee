App.IdeConsoleController = Em.ArrayController.extend App.IdeTabControllerMixin, App.TerminalConnectionMixin,
  needs: ['ideFileFinder', 'ide']

  modelClass: App.ConsoleSession
  content: null
  fullscreenBinding: 'App.consoleFullscreen'
  minimizedBinding: 'App.consoleMinimized'

  initialize: ->
    @set 'content', []
    @set 'fullscreen', (App.localStorage.get('fullscreenConsole') ? false)

  fullscreenDidChange: (->
    App.localStorage.set 'fullscreenConsole', @get('fullscreen')
  ).observes 'fullscreen'

  reset: ->
    @set 'content', null

  startAndAttach: (consoleSession, terminal) ->
    return unless socket = @get('controllers.ide.socket')

    start = =>
      onAttach = (subscriptionID, shellSessionID) =>
        consoleSession.setProperties
          subscriptionID: subscriptionID
          shellSessionID: shellSessionID
        terminal.focus()
        terminal.didOpen()
        terminal.onKeyDown = (data) =>
          socket.shellInput(shellSessionID, data)
        terminal.onResize = (cols, rows) =>
          socket.shellResize(shellSessionID, cols, rows)

      onData = (data) =>
        terminal.handleData(data);

      onError = (err) =>
        terminal.didFail(err)

      onExit = (data) =>
        terminal.didExit(data)

      termSize = terminal.getSize()

      socket.shellStart('sudo', [
        'login', '-f', 'nitrous', '-p', 'TERM=xterm-256color'
      ], [], '/', 'nitrous', termSize.cols, termSize.rows, onAttach, onData, onError, onExit)

    if socket.get('started')
      start()
    else
      socket.one('open', start)

  addTab: (tab, settings = {}) ->
    empty = @get('content.length') == 0
    @_super.apply this, arguments
    if empty
      @set 'minimized', false

  removeTab: (tab)->
    @_super.apply this, arguments
    if @get('content.length') == 0
      unless @get 'fullscreen'
        @set 'minimized', true
