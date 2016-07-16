App.IndexRoute = Em.Route.extend
  beforeModel: (transition) ->
    transition.abort()
    window.location.href = '/#/ide'
