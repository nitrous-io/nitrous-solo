App.IdeController = Em.ObjectController.extend
  needs: ['ideFileFinder', 'ideFindInProject']

  socket: null
  collabSocket: null
  collabEnabled: true

  repository: null
  errorDialog: null

  socketTimeoutAlert: null
  socketConnectionGracePeriod: 15 * 1000
  socketConnectionLost: false

  socketReadyBinding: Em.Binding.oneWay 'socket.started'
  socketNotReady: Em.computed.not 'socketReady'

  collabSocketReadyBinding: Em.Binding.oneWay 'collabSocket.started'
  collabSocketNotReady: Em.computed.not 'collabSocketReady'

  ideSocketsReady: (->
    return false if !@get('socketReady')

    if @get('collabEnabled')
      return @get('collabSocketReady')

    return true
  ).property 'socketReady', 'collabSocketReady', 'collabEnabled'

  ideSocketsNotReady: Em.computed.not 'ideSocketsReady'

  initialize: ->
    if App.localStorage.get('token')
      @set 'socket', socket = App.Socket.create()
      App.set('socket', socket)
      socket.connect()

      router = App.get 'router'
      router.get('ideLayoutController').initialize()
      router.get('ideMainMenuController').initialize()
      router.get('ideSidebarController').initialize()
      router.get('ideEditorController').initialize()
      router.get('ideConsoleController').initialize()
      @get('controllers.ideFileFinder').initialize()

  reset: ->
    if socket = @get 'socket'
      socket.disconnect()
      socket.destroy()

    router = App.get 'router'
    router.get('ideConsoleController').reset()
    router.get('ideEditorController').reset()
    router.get('ideSidebarController').reset()
    router.get('ideMainMenuController').reset()
    router.get('ideLayoutController').reset()

    @set 'socket', null
    @set 'repository', null
    @set 'content', null
    @set 'errorDialog', null

  socketDisconnected: (->
    unless @get('socketReady')
      socketTimeoutAlert = Em.run.later this, (->
        @set 'socketConnectionLost', true
      ), @get('socketConnectionGracePeriod')

      @set 'socketTimeoutAlert', socketTimeoutAlert
  ).observes 'socketReady'

  socketConnected: (->
    if @get('socketReady') && (socketTimeoutAlert = @get('socketTimeoutAlert'))
      Em.run.cancel socketTimeoutAlert
      @set 'socketConnectionLost', false
  ).observes 'socketReady'

  warnWhenOffline: (->
    if App.get 'offline'
      App.set 'flashMessage', { type: 'alert-error', text: "Unable to connect to the server. Changes will not be saved.", sticky: true }
    else
      App.set 'flashMessage', null
  ).observes 'App.offline'
