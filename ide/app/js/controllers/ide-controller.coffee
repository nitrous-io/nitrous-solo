freighterFeaturesRetry = 3000 # 3s

App.IdeController = Em.ObjectController.extend
  needs: ['ideFindInProject', 'ideFileFinder']

  opsSocket: null
  collabSocket: null
  collabEnabled: true

  repository: null
  errorDialog: null

  opsTimeoutAlert: null
  opsConnectionGracePeriod: 15 * 1000 # TODO: This should come from a KeyValue object from the server-side
  opsConnectionLost: false

  opsSocketReadyBinding: Em.Binding.oneWay 'opsSocket.started'
  opsSocketNotReady: Em.computed.not 'opsSocketReady'

  collabSocketReadyBinding: Em.Binding.oneWay 'collabSocket.started'
  collabSocketNotReady: Em.computed.not 'collabSocketReady'

  ideSocketsReady: (->
    return false if !@get('opsSocketReady')

    if @get('collabEnabled')
      return @get('collabSocketReady')

    return true
  ).property 'opsSocketReady', 'collabSocketReady', 'collabEnabled'

  ideSocketsNotReady: Em.computed.not 'ideSocketsReady'

  initialize: ->
    router = App.get 'router'
    router.get('ideLayoutController').initialize()
    router.get('ideMainMenuController').initialize()
    router.get('ideSidebarController').initialize()
    router.get('ideEditorController').initialize()
    router.get('ideConsoleController').initialize()

  reset: ->
    if opsSocket = @get 'opsSocket'
      opsSocket.disconnect()
      opsSocket.destroy()

    router = App.get 'router'
    router.get('ideConsoleController').reset()
    router.get('ideEditorController').reset()
    router.get('ideSidebarController').reset()
    router.get('ideMainMenuController').reset()
    router.get('ideLayoutController').reset()

    @set 'opsSocket', null
    @set 'repository', null
    @set 'content', null
    @set 'errorDialog', null
    @set 'resumeSessionDialog', null
    @set 'startProjectDialog', null

  opsSocketDisconnected: (->
    unless @get('opsSocketReady')
      opsTimeoutAlert = Em.run.later this, (->
        @set 'opsConnectionLost', true
      ), @get('opsConnectionGracePeriod')

      @set 'opsTimeoutAlert', opsTimeoutAlert
  ).observes 'opsSocketReady'

  opsSocketConnected: (->
    if @get('opsSocketReady') && (opsTimeoutAlert = @get('opsTimeoutAlert'))
      Em.run.cancel opsTimeoutAlert
      @set 'opsConnectionLost', false

      if repo = @get 'repository'
        if repo.get('state') == 'failed'
          cacheKey = "repo:#{repo.get('id')}:shownError"
          if !App.localStorage.get(cacheKey)
            App.set 'flashMessage', { type: 'error', text: "Failed to clone repository: #{repo.get('failure_reason')}" }
            App.localStorage.set(cacheKey, true)
  ).observes 'opsSocketReady'

  warnWhenOffline: (->
    if App.get 'offline'
      App.set 'flashMessage', { type: 'alert-error', text: "Unable to connect to the server. Changes will not be saved.", sticky: true }
    else
      App.set 'flashMessage', null
  ).observes 'App.offline'
