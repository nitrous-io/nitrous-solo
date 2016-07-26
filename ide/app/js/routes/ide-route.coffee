message = 'You will lose unsaved work if you close the IDE.'

App.IdeRoute = Em.Route.extend
  beforeModel: (transition) ->
    if !App.get('token')
      transition.abort()

      if t = App.localStorage.get("token")
        # token is found in local storage, check whether token is valid by fetching user
        $.ajax(
          type: 'GET'
          url: '/user'
          headers:
            Authorization: "Token #{t}"
        ).then ((res) =>
          # token is valid. set App.token and resume loading IDE
          App.set('token', t)
          App.set('user', App.User.create(res.user))
          transition.retry()
        ), (jqxhr) =>
          if jqxhr.responseJSON?.error == 'invalid_token'
            # token is invalid, remove token from storage and redirect to login
            App.localStorage.delete('token')
            window.location.href = '/#/login'
          else
            # unknown error. e.g. connection issue, internal server error etc.
            # set error message and redirect to error route
            errMsg = jqxhr.responseJSON?.error_description || jqxhr.responseJSON?.error || "unexpected response: #{jqxhr.status} #{jqxhr.statusText}"
            App.set('initError', errMsg.capitalize())
            window.location.href = '/#/error'
      else
        # token is not found in local storage, redirect to login
        window.location.href = '/#/login'
      return

  persistData: ->
    TabWatcher.syncChanges()
    App.get('user')?.save()

  activate: ->
    # warn the user before actually unload the window. This is helpful to
    # prevent mis-click close tab / cmd + w
    return if App.get 'redirecting'
    $(window).bind 'beforeunload.ide', (ev)=>
      @persistData()
      message

  deactivate: ->
    $(window).unbind 'beforeunload.ide'

  setupController: (controller) ->
    @_super.apply this, arguments
    controller.initialize()

  actions:
    willTransition: (transition) =>
      @persistData()
      return true if App.get 'isChromeWebView'
      return true if App.get 'redirecting'
      return true if window.confirm(message + '\n\nAre you sure you want to leave this page?')
      url = transition.router.recognizer.generate 'box.ide', box_id: this.get('controller.content.id')
      transition.router.updateURL(url)
      transition.abort()
