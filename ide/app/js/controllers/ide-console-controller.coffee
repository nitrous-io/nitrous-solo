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

  restoreState: ->
    return unless socket = @get('controllers.ide.socket')

    restore = =>
      socket.shellList (list) =>
        if list.length == 0
          @set('minimized', true)
        else
          @createTab(shellSessionID: v) for v in list

    if socket.get('started')
      restore()
    else
      socket.one('open', restore)

  attach: (consoleSession, terminal) ->
    return unless shellSessionID = consoleSession.get('shellSessionID')
    return unless socket = @get('controllers.ide.socket')

    attach = =>
      onAttach = (subscriptionID, scrollback) =>
        consoleSession.setProperties
          subscriptionID: subscriptionID
        terminal.focus()
        terminal.didOpen()

        if scrollback
          d = null
          try
            d = atob(scrollback)
          if d
            terminal?.handleData(d)

        terminal.onKeyDown = (data) =>
          socket.shellInput(shellSessionID, data)

        terminal.onResize = (cols, rows) =>
          socket.shellResize(shellSessionID, cols, rows)

      onData = (data) =>
        terminal?.handleData(data)

      onError = (err) =>
        terminal?.didFail(err)

      onExit = (data) =>
        terminal.didExit(data)
        terminal = null

      termSize = terminal.getSize()

      socket.shellAttach(shellSessionID, termSize.cols, termSize.rows, onAttach, onData, onError, onExit)

    if socket.get('started')
      attach()
    else
      socket.one('open', attach)

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
        terminal?.handleData(data);

      onError = (err) =>
        terminal?.didFail(err)

      onExit = (data) =>
        terminal.didExit(data)
        terminal = null

      termSize = terminal.getSize()

      socket.shellStart('sudo', [
        'login', '-f', 'nitrous', '-p', 'TERM=xterm-256color'
      ], [], '/', 'root', termSize.cols, termSize.rows, onAttach, onData, onError, onExit)

    if socket.get('started')
      start()
    else
      socket.one('open', start)

  removeTab: (tab) ->
    shellSessionID = tab.get('shellSessionID')
    socket = @get('controllers.ide.socket')

    if socket && socket.get('started') && shellSessionID
      socket.shellKill(shellSessionID)

    @_super.apply this, arguments
