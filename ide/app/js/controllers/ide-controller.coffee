freighterFeaturesRetry = 3000 # 3s

App.IdeController = Em.ObjectController.extend
  needs: ['ideFindInProject', 'ideFileFinder']

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
      socket.connect()

    router = App.get 'router'
    router.get('ideLayoutController').initialize()
    router.get('ideMainMenuController').initialize()
    router.get('ideSidebarController').initialize()
    router.get('ideEditorController').initialize()
    router.get('ideConsoleController').initialize()

    #pingerSocket = App.PingerSocket.create
    #  hostId: container.get('host_id')
    #  pingWhenVisible: false
    #App.set 'pingerSocket', pingerSocket
    #pingerSocket.connect()

    #@loadFreighterFeatures(container)

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
    @set 'resumeSessionDialog', null
    @set 'startProjectDialog', null

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

  collabSocketDidStartForTheFirstTime: (->
    if @get 'collabSocket.startedFirstTime'
      fileMon = App.ActiveFilesMonitor.create(collabSocket: @get('collabSocket'), username: App.get('user.username'))
      @set 'activeFilesMonitor', fileMon
  ).observes 'collabSocket.startedFirstTime'

  warnWhenOffline: (->
    if App.get 'offline'
      App.set 'flashMessage', { type: 'alert-error', text: "Unable to connect to the server. Changes will not be saved.", sticky: true }
    else
      App.set 'flashMessage', null
  ).observes 'App.offline'

  loadFreighterFeatures: (container)->
    @getFreighterFeatures(container).then =>
      if @loadFreighterFeaturesRetry
        window.clearTimeout @loadFreighterFeaturesRetry
      features = App.get('freighterdFeatures') ? []

      if features.contains('collab')
        @set 'collabEnabled', true
        @set 'collabSocket', collabSocket = App.CollabSocket.create(container: container)
        collabSocket.connect()
    , (jqxhr) =>
      return if @loadFreighterFeaturesRetry
      @loadFreighterFeaturesRetry = window.setTimeout(=>
        @loadFreighterFeaturesRetry = null
        @loadFreighterFeatures(container)
      , freighterFeaturesRetry)

  getFreighterFeatures: (container) ->
    protocol = location.protocol
    freighterdUrl = container.get 'freighterdUrl'
    url = "#{protocol}//#{freighterdUrl}/features"

    Em.RSVP.Promise.cast(Em.$.ajax(
      type: 'GET'
      url: url
      dataType: 'json'
    )).then (data) ->
      if Em.isArray(data.features)
        App.set 'freighterdFeatures', data.features
